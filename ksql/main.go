package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/chzyer/readline"
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

	type outputMode string
	const (
		outputModePretty     outputMode = "pretty"
		outputModeSingleLine outputMode = "single_line"
	)
	mode := outputModePretty

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

	historyFile := ""
	cacheDirBase, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get user cache dir: %w", err)
	}
	cacheDir := filepath.Join(cacheDirBase, "ksql")
	os.MkdirAll(cacheDir, 0755)
	historyFile = filepath.Join(cacheDir, "history")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 "> ",
		HistoryFile:            historyFile,
		DisableAutoSaveHistory: true,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	for {
		var buf strings.Builder
		prompt := "> "
		for {
			rl.SetPrompt(prompt)
			line, err := rl.Readline()
			if err != nil {
				if errors.Is(err, readline.ErrInterrupt) || errors.Is(err, io.EOF) {
					fmt.Println("")
					return nil
				}
				return err
			}

			trimmed := strings.TrimSpace(line)
			// Dot-commands are single-line and don't require a trailing ';'.
			if buf.Len() == 0 && strings.HasPrefix(trimmed, ".") {
				buf.WriteString(strings.TrimSuffix(trimmed, ";"))
				break
			}
			if buf.Len() == 0 && trimmed == "exit" {
				buf.WriteString(trimmed)
				break
			}
			buf.WriteString(line)
			buf.WriteString("\n")
			if strings.HasSuffix(strings.TrimSpace(buf.String()), ";") {
				break
			}
			prompt = "- "
		}
		text := strings.TrimSpace(buf.String())
		historyEntry := strings.Join(strings.Fields(text), " ")
		rl.SaveHistory(historyEntry)

		if text == "exit" {
			return nil
		}
		if strings.HasPrefix(text, ".") {
			fields := strings.Fields(text)
			if len(fields) >= 1 && fields[0] == ".mode" {
				if len(fields) != 2 {
					fmt.Println("usage: .mode pretty|single_line")
					continue
				}
				switch fields[1] {
				case string(outputModePretty):
					mode = outputModePretty
					fmt.Println("mode: pretty")
				case string(outputModeSingleLine):
					mode = outputModeSingleLine
					fmt.Println("mode: single_line")
				default:
					fmt.Println("usage: .mode pretty|single_line")
				}
				continue
			}
			fmt.Println("unknown command")
			continue
		}

		rows, err := conn.QueryContext(ctx, text)
		if err != nil {
			fmt.Printf("failed: %v\n", err)
			continue
		}
		cols, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			return fmt.Errorf("failed: %w", err)
		}

		var results []map[string]any
		for rows.Next() {
			values := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range values {
				ptrs[i] = &values[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				_ = rows.Close()
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
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return fmt.Errorf("failed: %w", err)
		}
		_ = rows.Close()

		switch mode {
		case outputModePretty:
			out, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(out))
		case outputModeSingleLine:
			for _, row := range results {
				out, err := json.Marshal(row)
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(out))
			}
		default:
			return fmt.Errorf("unknown output mode: %q", mode)
		}

	}
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
		WHEN date_diff('days', x, current_timestamp AT TIME ZONE 'UTC') >= 1 THEN
			date_diff('days', x, current_timestamp AT TIME ZONE 'UTC')::VARCHAR || 'd'
		WHEN date_diff('hours', x, current_timestamp AT TIME ZONE 'UTC') >= 1 THEN
			date_diff('hours', x, current_timestamp AT TIME ZONE 'UTC')::VARCHAR || 'h'
		ELSE
			date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC')::VARCHAR || 'm'
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
