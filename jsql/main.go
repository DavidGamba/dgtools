package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/DavidGamba/dgtools/jsql/repl"
	"github.com/DavidGamba/dgtools/run"
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
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.SetUnknownMode(getoptions.Pass)
	get := opt.NewCommand("get", "description").SetCommandFn(GetRun)
	get.Bool("all-namespaces", false, opt.Alias("A"))
	get.HelpSynopsisArg("<resource-types>...", "type of the resources to get")

	opt.NewCommand("query", "description").SetCommandFn(QueryRun)

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
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
)

func QueryRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")

	mode := outputModeTable

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

		if strings.HasPrefix(lines[0], ".mode") {
			switch {
			case regexp.MustCompile(`(?s)(?i)\.mode\s+pretty`).MatchString(query):
				mode = outputModePretty
			case regexp.MustCompile(`(?s)(?i)\.mode\s+single_line`).MatchString(query):
				mode = outputModeSingleLine
			case regexp.MustCompile(`(?s)(?i)\.mode\s+table`).MatchString(query):
				mode = outputModeTable
			default:
				fmt.Printf(`Valid modes:

table: (default) pretty print tables and json marshal nested data
pretty: json marshal results
single_line: json marshal into one record per line
`)
			}
			continue
		}

		if strings.HasPrefix(lines[0], ".help") {
			fmt.Printf("%s\n", repl.DefaultHeader())
			fmt.Printf(`
.mode <pretty|single_line|table>    - set output mode
.help                               - show this message
`)
			continue
		}

		err = runQuery(ctx, conn, mode, query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
}

func GetRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")
	allNamespaces := opt.Value("all-namespaces").(bool)

	contextName, namespace, err := GetK8sContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get k8s context: %w", err)
	}
	Logger.Printf("Current context: %s, namespace: %s", contextName, namespace)

	cacheDirBase, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get user cache dir: %w", err)
	}
	cacheDir := filepath.Join(cacheDirBase, "jsql", contextName)
	Logger.Printf("Using cache dir: %s", cacheDir)
	os.MkdirAll(cacheDir, 0755)

	// TODO: Don't donwload every time but use a flag to force cache invalidation, otherwise re-use cache and only invalidate after a given age.

	for _, rt := range args {
		cmd := []string{"kubectl", "get", "-o", "json", rt}
		if allNamespaces {
			cmd = append(cmd, "-A")
		}
		out, err := run.CMD(cmd...).Log().STDOutOutput()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
		cmd = []string{"qq", ".items", "-o", "json"}
		out, err = run.CMD(cmd...).In(out).STDOutOutput()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
		filename := filepath.Join(cacheDir, rt+".json")
		fh, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer fh.Close()
		_, err = fh.Write(out)
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}

		cmds := []string{
			fmt.Sprintf("DROP TABLE IF EXISTS %s;", rt),
			fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM '%s';", rt, filename),
			fmt.Sprintf("ALTER TABLE %s ADD COLUMN name VARCHAR;", rt),
			fmt.Sprintf("UPDATE %s SET name = metadata.name;", rt),
			fmt.Sprintf("ALTER TABLE %s ADD COLUMN namespace VARCHAR;", rt),
			// Use cast to allow for null values
			fmt.Sprintf("UPDATE %s SET namespace = CAST(metadata AS JSON)->>'namespace';", rt),
		}
		for _, e := range cmds {
			cmd := []string{"duckdb", DBNAME, "-s", e}
			err = run.CMD(cmd...).Log().Run()
			if err != nil {
				return fmt.Errorf("failed: %w", err)
			}
		}

	}

	cmds := []string{
		"CREATE SCHEMA IF NOT EXISTS k8s;",

		`CREATE OR REPLACE MACRO k8s.age(x) AS
	CASE
		WHEN x IS NULL THEN NULL
		WHEN date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC') >= (24*60) THEN
			(date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC')//(24*60))::VARCHAR || 'd' ||
			(date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC')%(60*24)//60)::VARCHAR || 'h'
		WHEN date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC') >= 60 THEN
			(date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC')%(60*24)//60)::VARCHAR || 'h' ||
			(date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC')%60)::VARCHAR || 'm'
		ELSE
			(date_diff('seconds', x, current_timestamp AT TIME ZONE 'UTC')//60)::VARCHAR || 'm' ||
			(date_diff('seconds', x, current_timestamp AT TIME ZONE 'UTC')%60)::VARCHAR || 's'
	END;`,

		"CREATE OR REPLACE MACRO k8s.agem(x) AS k8s.age(x.metadata.creationTimestamp);",

		"CREATE OR REPLACE MACRO k8s.agema(x_metadata) AS k8s.age(x_metadata.creationTimestamp);",

		`CREATE OR REPLACE MACRO k8s.cpu_m(cpu) AS
	CASE
		WHEN trim(cpu) LIKE '%m' THEN
			CAST(CEILING(CAST(RTRIM(trim(cpu), 'm') AS NUMERIC)) AS BIGINT)
		ELSE
			CAST(CAST(trim(cpu) AS NUMERIC) * 1000 AS BIGINT)
	END;`,

		`CREATE OR REPLACE MACRO k8s.memory_bytes(memory) AS
	CAST(CASE
		WHEN trim(memory) LIKE '%Ki' THEN
			CAST(RTRIM(trim(memory), 'Ki') AS NUMERIC) * 1024
		WHEN trim(memory) LIKE '%Mi' THEN
			CAST(RTRIM(trim(memory), 'Mi') AS NUMERIC) * 1048576
		WHEN trim(memory) LIKE '%Gi' THEN
			CAST(RTRIM(trim(memory), 'Gi') AS NUMERIC) * 1073741824
		WHEN trim(memory) LIKE '%Ti' THEN
			CAST(RTRIM(trim(memory), 'Ti') AS NUMERIC) * 1099511627776
		WHEN trim(memory) LIKE '%Pi' THEN
			CAST(RTRIM(trim(memory), 'Pi') AS NUMERIC) * 1125899906842624
		WHEN trim(memory) LIKE '%Ei' THEN
			CAST(RTRIM(trim(memory), 'Ei') AS NUMERIC) * 1152921504606846976
		WHEN trim(memory) LIKE '%m' THEN
			CEILING(CAST(RTRIM(trim(memory), 'm') AS NUMERIC) / 1000)
		WHEN trim(memory) LIKE '%k' THEN
			CAST(RTRIM(trim(memory), 'k') AS NUMERIC) * 1000
		WHEN trim(memory) LIKE '%M' THEN
			CAST(RTRIM(trim(memory), 'M') AS NUMERIC) * 1000000
		WHEN trim(memory) LIKE '%G' THEN
			CAST(RTRIM(trim(memory), 'G') AS NUMERIC) * 1000000000
		WHEN trim(memory) LIKE '%T' THEN
			CAST(RTRIM(trim(memory), 'T') AS NUMERIC) * 1000000000000
		WHEN trim(memory) LIKE '%P' THEN
			CAST(RTRIM(trim(memory), 'P') AS NUMERIC) * 1000000000000000
		WHEN trim(memory) LIKE '%E' THEN
			CAST(RTRIM(trim(memory), 'E') AS NUMERIC) * 1000000000000000000
		ELSE
			CAST(trim(memory) AS NUMERIC)
	END AS BIGINT);`,

		`CREATE OR REPLACE MACRO k8s.memory_human(bytes) AS
	CASE
		WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 6) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 6)), '.') || 'Ei'
		WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 5) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 5)), '.') || 'Pi'
		WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 4) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 4)), '.') || 'Ti'
		WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 3) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 3)), '.') || 'Gi'
		WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 2) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 2)), '.') || 'Mi'
		WHEN CAST(bytes AS DOUBLE) >= 1024               THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / 1024), '.') || 'Ki'
		ELSE CAST(bytes AS VARCHAR)
	END;`,

		`CREATE OR REPLACE VIEW drspn AS
	SELECT
		d_kind:d.kind,
		d_apiVersion:d.apiVersion,
		d_name:d.name,
		d_namespace:d.namespace,
		d_metadata:d.metadata,
		d_spec:d.spec,
		d_status:d.status,
		rs_kind:rs.kind,
		rs_apiVersion:rs.apiVersion,
		rs_name:rs.name,
		rs_namespace:rs.namespace,
		rs_metadata:rs.metadata,
		rs_spec:rs.spec,
		rs_status:rs.status,
		p_kind:p.kind,
		p_apiVersion:p.apiVersion,
		p_name:p.name,
		p_namespace:p.namespace,
		p_metadata:p.metadata,
		p_spec:p.spec,
		p_status:p.status,
		n_kind:n.kind,
		n_apiVersion:n.apiVersion,
		n_name:n.name,
		n_namespace:n.namespace,
		n_metadata:n.metadata,
		n_spec:n.spec,
		n_status:n.status
	FROM deploy AS d
	JOIN rs ON rs.namespace = d.namespace AND rs.metadata.ownerReferences[1].uid = d.metadata.uid
	JOIN pods AS p ON p.namespace = rs.namespace AND rs.metadata.uid = p.metadata.ownerReferences[1].uid
	LEFT OUTER JOIN nodes as n ON n.name = p.spec.nodeName
;`,

		`CREATE OR REPLACE VIEW spn AS
	SELECT
		s_kind:s.kind,
		s_apiVersion:s.apiVersion,
		s_name:s.name,
		s_namespace:s.namespace,
		s_metadata:s.metadata,
		s_spec:s.spec,
		s_status:s.status,
		p_kind:p.kind,
		p_apiVersion:p.apiVersion,
		p_name:p.name,
		p_namespace:p.namespace,
		p_metadata:p.metadata,
		p_spec:p.spec,
		p_status:p.status,
		n_kind:n.kind,
		n_apiVersion:n.apiVersion,
		n_name:n.name,
		n_namespace:n.namespace,
		n_metadata:n.metadata,
		n_spec:n.spec,
		n_status:n.status
	FROM sts AS s
	JOIN pods AS p ON p.namespace = s.namespace AND s.metadata.uid = p.metadata.ownerReferences[1].uid
	LEFT OUTER JOIN nodes as n ON n.name = p.spec.nodeName
;`,

		`CREATE OR REPLACE VIEW pn AS
	SELECT
		p_kind:p.kind,
		p_apiVersion:p.apiVersion,
		p_name:p.name,
		p_namespace:p.namespace,
		p_metadata:p.metadata,
		p_spec:p.spec,
		p_status:p.status,
		n_kind:n.kind,
		n_apiVersion:n.apiVersion,
		n_name:n.name,
		n_namespace:n.namespace,
		n_metadata:n.metadata,
		n_spec:n.spec,
		n_status:n.status
	FROM nodes AS n
	JOIN pods as p ON n.name = p.spec.nodeName
;`,
	}

	for _, e := range cmds {
		cmd := []string{"duckdb", DBNAME, "-s", e}
		err = run.CMD(cmd...).Log().Run()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}

	return nil
}
