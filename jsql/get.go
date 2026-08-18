package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func GetRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")
	return GetExec(ctx, args)
}

func GetExec(ctx context.Context, args []string) error {
	// Get config
	d := &Config{}
	err := ReadConfig(d)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// If no arguments are provided list all providers
	if len(args) < 1 {
		fmt.Printf("Valid provider list: \n")
		for k := range d.Provider {
			fmt.Printf("%s\n", k)
		}
		return nil
	}

	// get and validate provider
	provider, args := args[0], args[1:]
	if _, ok := d.Provider[provider]; !ok {
		return fmt.Errorf("invalid provider: %s", provider)
	}

	// If no arguments are provided list all commands for the provider
	if len(args) < 1 {
		fmt.Printf("Valid commands for provider %s: \n", provider)
		for _, cmd := range d.Provider[provider].GetCommands {
			fmt.Printf("%s\n", cmd.Name)
		}
		return nil
	}

	// get and validate command
	command, args := args[0], args[1:]
	if _, ok := d.Provider[provider].GetCommands[command]; !ok {
		return fmt.Errorf("invalid command: %s", command)
	}

	// validate arguments
	if len(d.Provider[provider].GetCommands[command].Args) != len(args) {
		fmt.Fprintf(os.Stderr, "Usage:\n    %s %s", provider, command)
		for _, arg := range d.Provider[provider].GetCommands[command].Args {
			fmt.Fprintf(os.Stderr, " <%s>", arg.Name)
		}
		fmt.Fprintf(os.Stderr, "\n\nARGUMENTS:\n")
		for _, arg := range d.Provider[provider].GetCommands[command].Args {
			fmt.Fprintf(os.Stderr, "    %-20s %s\n", arg.Name, arg.Description)
		}
		return fmt.Errorf("missing args")
	}

	// Replace placeholders in table name
	table := d.Provider[provider].GetCommands[command].Table
	table = strings.ReplaceAll(table, "$schemaName", d.Provider[provider].SchemaName)
	table = strings.ReplaceAll(table, "$provider", provider)
	if len(args) > 0 {
		table = strings.ReplaceAll(table, "$arg1", args[0])
	}
	if len(args) > 1 {
		table = strings.ReplaceAll(table, "$arg2", args[1])
	}

	Logger.Printf("Provider: %s, Command: %s, Args: %v", provider, command, args)

	// Create cache dir
	cacheDir, err := createCacheDir(provider)
	if err != nil {
		return fmt.Errorf("failed to get cache dir: %w", err)
	}
	Logger.Printf("Using cache dir: %s", cacheDir)

	filename := filepath.Join(cacheDir, command+".json")

	// Run get command
	commandData := d.Provider[provider].GetCommands[command]
	commandParts := commandData.Command
	Logger.Printf("data: %v\n", commandData)
	for i, e := range commandParts {
		if len(args) > 0 {
			e = strings.ReplaceAll(e, "$arg1", args[0])
		}
		if len(args) > 1 {
			e = strings.ReplaceAll(e, "$arg2", args[1])
		}
		e = strings.ReplaceAll(e, "$table", table)
		e = strings.ReplaceAll(e, "$schemaName", d.Provider[provider].SchemaName)
		e = strings.ReplaceAll(e, "$provider", provider)
		e = strings.ReplaceAll(e, "$command", command)
		e = strings.ReplaceAll(e, "$filename", filename)
		commandParts[i] = e
	}
	Logger.Printf("Running command: %v\n", commandParts)
	out, err := run.CMD(commandParts...).Log().STDOutOutput()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	// TODO: filter goes here

	// Save to file
	fh, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer fh.Close()
	_, err = fh.Write(out)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	cmds := d.Provider[provider].GetCommands[command].Create
	for _, e := range cmds {
		if len(args) > 0 {
			e = strings.ReplaceAll(e, "$arg1", args[0])
		}
		if len(args) > 1 {
			e = strings.ReplaceAll(e, "$arg2", args[1])
		}
		e = strings.ReplaceAll(e, "$table", table)
		e = strings.ReplaceAll(e, "$schemaName", d.Provider[provider].SchemaName)
		e = strings.ReplaceAll(e, "$provider", provider)
		e = strings.ReplaceAll(e, "$command", command)
		e = strings.ReplaceAll(e, "$filename", filename)
		cmd := []string{"duckdb", DBNAME, "-s", e}
		err = run.CMD(cmd...).Log().Run()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}

	return nil

	contextName, namespace, err := GetK8sContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get k8s context: %w", err)
	}
	Logger.Printf("Current context: %s, namespace: %s", contextName, namespace)

	// TODO: Don't donwload every time but use a flag to force cache invalidation, otherwise re-use cache and only invalidate after a given age.

	for _, resource := range args {
		err := GetK8sResource(ctx, cacheDir, resource)
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}

		cmds := UpdateK8sResourceQueries(cacheDir, resource)
		for _, e := range cmds {
			cmd := []string{"duckdb", DBNAME, "-s", e}
			err = run.CMD(cmd...).Log().Run()
			if err != nil {
				return fmt.Errorf("failed: %w", err)
			}
		}

	}

	cmds = []string{
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
