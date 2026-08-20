package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/DavidGamba/dgtools/cueutils"
	"github.com/DavidGamba/dgtools/jsql/repl"
	"github.com/DavidGamba/go-getoptions"
	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/nyaosorg/go-readline-ny"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

var DBNAME = "jsql.duckdb"

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Bool("quiet", false)
	opt.SetUnknownMode(getoptions.Pass)
	get := opt.NewCommand("get", "Run provider's Get command to retrieve data").SetCommandFn(GetRun)
	get.HelpSynopsisArg("<provider_name>", "provider to use")
	get.HelpSynopsisArg("<args>...", "provider arguments")

	opt.NewCommand("query", "Run query's on the DB present in the CWD").SetCommandFn(QueryRun)

	opt.NewCommand("config", "Show parsed config").SetCommandFn(ConfigRun)

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
		cueutils.Logger.SetOutput(io.Discard)
	} else {
		// cueutils.Logger.SetOutput(os.Stderr)
	}
	Logger.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

type outputMode string

const (
	outputModePretty     outputMode = "pretty"
	outputModeSingleLine outputMode = "single_line"
	outputModeTable      outputMode = "table"
	outputModeCSV        outputMode = "csv"
)

type queryOption int

const (
	queryOptionClear      queryOption = 1 << iota // No options set
	queryOptionAutoNumber                         // set autonumber
)

func QueryRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")

	d := &Config{}
	err := ReadConfig(d, nil)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	mode := outputModePretty
	qo := queryOptionAutoNumber

	conn, err := dbConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	history, err := repl.NewHistoryFile("jsql", "history")
	if err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}

	r := repl.New(history, completionCandidates)
	r.SubmitOnEnterWhenEndsOn(";")

	r.Ed.Highlight = append(r.Ed.Highlight, readline.Highlight{
		Pattern: regexp.MustCompile(`(?:\b|^)(?i)(` + strings.Join(AllKeywords(), "|") + `)(?:\b|$)`), Sequence: "\x1B[36;49;1m",
	})

	// Ignore Ctrl-C in SQL repl
	r.IgnoreSIGINT = true
	signal.Ignore(syscall.SIGINT)

	writer := os.Stdout

	for lines, err := range repl.Interactive(ctx, r) {
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		query := strings.Join(lines, "\n")
		fmt.Println("----")
		fmt.Println(query)
		fmt.Println("----")
		err := history.Add(strings.Join(lines, "⏎"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to add to history: %v\n", err)
		}

		if strings.HasPrefix(lines[0], ".help") {
			fmt.Printf("%s\n", repl.DefaultHeader())
			fmt.Printf(`
.mode <pretty|single_line|table|csv>    - set output mode
.output <stdout|file <filename>>        - set output target
.option clear                           - clear all options
.option autonumber                      - add a numbering column
.help                                   - show this message

Common use DuckDB syntax:

Aggregate Functions: sum, count, list, string_agg
`)
			continue
		}

		if strings.HasPrefix(lines[0], ".output") {
			fileRegex := regexp.MustCompile(`(?s)(?i)\.output\s+file\s+(.+)\s*;`)
			switch {
			case regexp.MustCompile(`(?s)(?i)\.output\s+(?:stdout|terminal|cli)`).MatchString(query):
				writer = os.Stdout
			case fileRegex.MatchString(query):
				matches := fileRegex.FindStringSubmatch(query)
				if len(matches) > 1 {
					filename := matches[1]
					fh, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						fmt.Fprintf(os.Stderr, "ERROR: failed to open file: %v\n", err)
						continue
					}
					defer fh.Close()
					writer = fh
				}
			default:
				fmt.Printf(`Valid outputs:

stdout: (default) print to stdout
file <filename>: save to file
`)
			}
			continue
		}

		if strings.HasPrefix(lines[0], ".mode") {
			switch {
			case regexp.MustCompile(`(?s)(?i)\.mode\s+pretty`).MatchString(query):
				mode = outputModePretty
			case regexp.MustCompile(`(?s)(?i)\.mode\s+single_line`).MatchString(query):
				mode = outputModeSingleLine
			case regexp.MustCompile(`(?s)(?i)\.mode\s+table`).MatchString(query):
				mode = outputModeTable
			case regexp.MustCompile(`(?s)(?i)\.mode\s+csv`).MatchString(query):
				mode = outputModeCSV
			default:
				fmt.Printf(`Valid modes:

pretty: (default) json marshal results
table: pretty print tables and json marshal nested data
single_line: json marshal into one record per line
`)
			}
			continue
		}

		if strings.HasPrefix(lines[0], ".option") {
			switch {
			case regexp.MustCompile(`(?s)(?i)\.option\s+autonumber`).MatchString(query):
				qo |= queryOptionAutoNumber
			case regexp.MustCompile(`(?s)(?i)\.option\s+clear`).MatchString(query):
				qo = queryOptionClear
			default:
				fmt.Printf(`Valid options:

clear:      clear all options
autonumber: add a numbering column
`)
			}
			continue
		}

		// if strings.HasPrefix(lines[0], ".kget") {
		// 	resourceRegex := regexp.MustCompile(`(?s)(?i)\.kget\s+(.+)\s*;`)
		// 	switch {
		// 	case resourceRegex.MatchString(query):
		// 		matches := resourceRegex.FindStringSubmatch(query)
		// 		if len(matches) > 1 {
		// 			resource := matches[1]
		//
		// 			contextName, namespace, err := GetK8sContext(ctx)
		// 			if err != nil {
		// 				return fmt.Errorf("failed to get k8s context: %w", err)
		// 			}
		// 			Logger.Printf("Current context: %s, namespace: %s", contextName, namespace)
		//
		// 			cacheDir, err := createCacheDir(contextName)
		// 			if err != nil {
		// 				return fmt.Errorf("failed to get cache dir: %w", err)
		// 			}
		// 			Logger.Printf("Using cache dir: %s", cacheDir)
		//
		// 		}
		// 	default:
		// 		fmt.Printf(`Usage: .kget <resource> `)
		// 	}
		// 	continue
		// }

		err = runQuery(ctx, writer, conn, mode, qo, query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
}
