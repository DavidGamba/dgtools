package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	_ "github.com/duckdb/duckdb-go/v2"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

var DBNAME = "ksql.duckdb"

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

func QueryRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")

	db, err := sql.Open("duckdb", DBNAME)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	defer db.Close()

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	defer conn.Close()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("")
				return nil
			}
			return err
		}
		if text == "exit" {
			break
		}

		rows, err := conn.QueryContext(ctx, text)
		if err != nil {
			fmt.Printf("failed: %v\n", err)
			continue
		}

		defer rows.Close()
		cols, _ := rows.Columns()

		var results []map[string]any
		for rows.Next() {
			values := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range values {
				ptrs[i] = &values[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				return fmt.Errorf("failed: %w", err)
			}
			row := make(map[string]any, len(cols))
			for i, col := range cols {
				val := values[i]
				if b, ok := val.([]byte); ok {
					val = string(b)
				}
				row[col] = val
			}
			results = append(results, row)
		}

		out, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(out))

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
	cacheDir := filepath.Join(cacheDirBase, "ksql", contextName)
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
		cmd = []string{"duckdb", DBNAME, "-s", fmt.Sprintf("DROP TABLE IF EXISTS %s;", rt)}
		err = run.CMD(cmd...).Log().Run()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
		cmd = []string{"duckdb", DBNAME, "-s", fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM '%s';", rt, filename)}
		err = run.CMD(cmd...).Log().Run()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}

	cmds := []string{
		"CREATE SCHEMA IF NOT EXISTS k8s;",
		"CREATE FUNCTION IF NOT EXISTS k8s.age(x) AS date_diff('days', x.metadata.creationtimestamp, today());",
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
