// This file is part of dgtools.
//
// # Copyright (C) 2019-2023  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/dgtools/yamlutils"
	"github.com/DavidGamba/go-getoptions"
)

// BuildMetadata - Provides the metadata part of the version information.
var BuildMetadata = "dev"

const semVersion = "0.1.0"

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Self("", `Parses YAML input passed from file(s) or piped to STDIN and allows to split it or combine it.

        Source: https://github.com/DavidGamba/dgtools`)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.Bool("version", false, opt.Alias("V"))
	opt.Bool("silent", false, opt.Description("Don't print full context errors."))

	read := opt.NewCommand("read", "read a multi document yaml file")
	read.HelpSynopsisArgs("<file>...")
	read.Bool("-", false, opt.Description("Read from STDIN"))
	read.StringSlice("key", 1, 99, opt.Alias("k"), opt.ArgName("key/index"),
		opt.Description(`Key or index to descend to.
Multiple keys allow to descend further.
Indexes are positive integers.`))
	read.Bool("hide-keys", false, opt.Alias("hk"), opt.Description("don't show keys in output"))
	read.SetCommandFn(ReadRun)

	split := opt.NewCommand("split", "split a multi document YAML file")
	split.HelpSynopsisArgs("<file>...")
	split.Bool("-", false, opt.Description("Read from STDIN"))
	split.Bool("force", false, opt.Description("Apply split"))
	split.String("dir", "", opt.Description("Output directory to write files to. Defaults to same as source."))
	split.String("output-prefix", "", opt.Alias("prefix"), opt.Description("Output Filename prefix"))
	split.StringSlice("key", 1, 99, opt.Alias("k"), opt.ArgName("key/index"),
		opt.Description(`Make keys data part of the filename, for example: -k kind -k metadata/name
If not used, the default name is the filename-<document-number>.yaml`))
	split.SetCommandFn(SplitRun)

	join := opt.NewCommand("join", "join multiple YAML files into a single multi document one")
	join.HelpSynopsisArgs("<file>...")
	join.String("output", "", opt.Required(), opt.Description("Output file"))
	join.SetCommandFn(JoinRun)

	opt.HelpCommand("help", opt.Alias("h", "?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("version") {
		fmt.Printf("Version: %s+%s\n", semVersion, BuildMetadata)
		return 0
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

func readInput(ctx context.Context, useStdIn bool, file string) ([]*yamlutils.YML, error) {
	// Check if stdin is pipe p or device D
	statStdin, _ := os.Stdin.Stat()
	stdinIsDevice := (statStdin.Mode() & os.ModeDevice) != 0

	var err error
	var ymlList []*yamlutils.YML
	if !stdinIsDevice && useStdIn {
		Logger.Printf("Reading from STDIN\n")
		reader := os.Stdin
		ymlList, err = yamlutils.NewFromReader(reader)
		if err != nil {
			return ymlList, fmt.Errorf("reading yaml from STDIN: %w", err)
		}
	} else {
		Logger.Printf("Reading from file: %s\n", file)
		ymlList, err = yamlutils.NewFromFile(file)
		if err != nil {
			return ymlList, fmt.Errorf("reading yaml file: %w", err)
		}
	}
	return ymlList, nil
}

func ReadRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	keys := opt.Value("key").([]string)
	useStdIn := opt.Value("-").(bool)
	hideKeys := opt.Value("hide-keys").(bool)

	if len(args) < 1 && !useStdIn {
		fmt.Fprintf(os.Stderr, "ERROR: missing <file> or STDIN input '-'\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	var xpaths [][]string
	for _, k := range keys {
		var xpath []string
		k = strings.TrimLeft(k, "/")
		xpath = append(xpath, strings.Split(k, "/")...)
		Logger.Printf("path: '%s'\n", strings.Join(xpath, ","))
		xpaths = append(xpaths, xpath)
	}

	errorCount := 0
	files := args
	if useStdIn {
		files = append(files, "-")
	}
	for _, file := range files {
		sIn := false
		if file == "-" {
			sIn = true
		}
		ymlList, err := readInput(ctx, sIn, file)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		for i, yml := range ymlList {
			fmt.Printf("# %s %02d\n", file, i+1)
			for _, xpath := range xpaths {
				str, err := yml.GetString(!hideKeys, xpath)
				if err != nil {
					errorCount++
					fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
					if !opt.Called("silent") {
						fmt.Fprintf(os.Stderr, ">\t%s\n", strings.ReplaceAll(str, "\n", "\n>\t"))
					}
					continue
				}
				str = strings.TrimSpace(str)
				fmt.Println(str)
			}
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("found %d errors when reading documents", errorCount)
	}

	return nil
}

func SplitRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	dir := opt.Value("dir").(string)
	force := opt.Value("force").(bool)
	useStdIn := opt.Value("-").(bool)
	oPrefix := opt.Value("output-prefix").(string)
	keys := opt.Value("key").([]string)

	if len(args) < 1 && !useStdIn {
		fmt.Fprintf(os.Stderr, "ERROR: missing <file> or STDIN input '-'\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}
	if useStdIn && !opt.Called("output-prefix") {
		fmt.Fprintf(os.Stderr, "ERROR: --output-prefix <string> is required when reading from STDIN '-'\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	file := "-"
	if len(args) > 0 {
		file = args[0]
	}

	if oPrefix == "" {
		oPrefix = strings.TrimSuffix(file, ".yaml")
		oPrefix = strings.TrimSuffix(oPrefix, ".yml")
	}
	if oPrefix == "" {
		oPrefix = "stdin"
	}
	outputDir := filepath.Dir(oPrefix)
	if dir != "" {
		outputDir = dir
	}
	if force {
		_ = os.MkdirAll(outputDir, 0755)
	}

	oPrefix = filepath.Base(oPrefix)

	nameKeys := [][]string{}
	for _, e := range keys {
		nameKeys = append(nameKeys, strings.Split(e, "/"))
	}

	ymlList, err := readInput(ctx, useStdIn, file)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	filenameCounter := map[string]int{}
	for i, yml := range ymlList {
		filename := fmt.Sprintf(`%[1]s-%02[2]d.yaml`, oPrefix, i+1)
		nameParts := []string{}
		for _, e := range nameKeys {
			es, err := yml.GetString(false, e)
			if err == nil {
				nameParts = append(nameParts, strings.TrimSpace(es))
			}
		}
		if len(nameParts) != 0 {
			filename = strings.Join(nameParts, "-")
			v, ok := filenameCounter[filename]
			if ok {
				filenameCounter[filename] = v + 1
				filename += fmt.Sprintf(`-%02d`, v+1)
			} else {
				filenameCounter[filename] = 1
				filename += "-01"
			}
			filename += ".yaml"
		}
		filename = filepath.Join(outputDir, filename)
		fmt.Printf("%s\n", filename)
		if force {
			str, err := yml.GetString(false, []string{})
			if err != nil {
				return fmt.Errorf("failed to read document %d: %w", i+1, err)
			}
			os.WriteFile(filename, []byte(str), 0640)
		}
	}

	if !force {
		fmt.Fprintf(os.Stderr, "WARNING: Running in Dry Run mode, use --force to apply changes\n")
	}

	return nil
}

func JoinRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	output := opt.Value("output").(string)

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <file>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	fh, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	for i, file := range args {
		if i > 0 {
			fh.WriteString("\n---\n")
		}
		contents, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file '%s': %w", file, err)
		}
		fh.Write(contents)
	}
	return nil
}
