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
	err := ReadConfig(d, nil)
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
	var out []byte
	if len(commandParts) > 0 {
		Logger.Printf("Running command: %v\n", commandParts)
		out, err = run.CMD(commandParts...).Log().STDOutOutput()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}

	// Run command data filter
	filterParts := commandData.Filter
	for i, e := range filterParts {
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
		filterParts[i] = e
	}
	if len(filterParts) > 0 {
		Logger.Printf("Running command: %v\n", filterParts)
		out, err = run.CMD(filterParts...).In(out).Log().STDOutOutput()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}

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

	cmds = d.Provider[provider].Macros
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
}
