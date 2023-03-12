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
	"strings"

	"github.com/DavidGamba/dgtools/yamlutils"
	"github.com/DavidGamba/go-getoptions"
)

// BuildMetadata - Provides the metadata part of the version information.
var BuildMetadata = "dev"

const semVersion = "0.6.0"

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Self("", `Parses YAML input passed from file or piped to STDIN and filters it by key or index.

         Source: https://github.com/DavidGamba/dgtools`)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.Bool("version", false, opt.Alias("V"))
	opt.Bool("-", false, opt.Description("Read from STDIN"))
	opt.Bool("n", false, opt.Description("Remove trailing spaces."))
	opt.Bool("silent", false, opt.Description("Don't print full context errors."))
	opt.Bool("include", false, opt.Description("Include parent key if it is a map key."))
	opt.String("add", "", opt.ArgName("yaml/json input"), opt.Description("Child input to add at the current location."))
	opt.StringSlice("key", 1, 99, opt.Alias("k"), opt.ArgName("key/index"),
		opt.Description(`Key or index to descend to.
Multiple keys allow to descend further.
Indexes are positive integers.`))
	opt.IntOptional("document", 1, opt.Description("Document number"), opt.Alias("d"), opt.ArgName("number"))
	opt.HelpSynopsisArg("<file>...", `File(s) containing YAML document(s).
If '-' is given, read from STDIN.`)
	opt.HelpCommand("help", opt.Alias("?"))
	opt.SetCommandFn(Run)
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

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	include := opt.Value("include").(bool)
	add := opt.Value("add").(string)
	keys := opt.Value("key").([]string)
	useStdIn := opt.Value("-").(bool)

	if len(args) < 1 && !useStdIn {
		fmt.Fprintf(os.Stderr, "ERROR: missing <file> or STDIN input '-'\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	file := "-"
	if len(args) > 0 && !useStdIn {
		file = args[0]
	}

	var xpath []string
	for _, k := range keys {
		k = strings.TrimLeft(k, "/")
		xpath = append(xpath, strings.Split(k, "/")...)
	}
	Logger.Printf("path: '%s'\n", strings.Join(xpath, ","))

	ymlList, err := readInput(ctx, useStdIn, file)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	if !opt.Called("document") && len(ymlList) > 1 {
		fmt.Fprintf(os.Stderr,
			`WARNING: provided input contains %d YAML documents
         specify '--document <number>' to remove this warning
`, len(ymlList))
	}

	n := 0
	if opt.Called("document") {
		n = opt.Value("document").(int) - 1
	}
	if len(ymlList) < n+1 {
		return fmt.Errorf("wrong document number: %d", n+1)
	}
	yml := ymlList[n]

	if opt.Called("add") {
		str, err := yml.AddString(xpath, add)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			if !opt.Called("silent") {
				fmt.Fprintf(os.Stderr, ">\t%s\n", strings.ReplaceAll(str, "\n", "\n>\t"))
			}
			os.Exit(1)
		}
		if opt.Called("n") {
			str = strings.TrimSpace(str)
		}
		fmt.Print(str)
		return nil
	}

	str, err := yml.GetString(include, xpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		if !opt.Called("silent") {
			fmt.Fprintf(os.Stderr, ">\t%s\n", strings.ReplaceAll(str, "\n", "\n>\t"))
		}
		os.Exit(1)
	}
	if opt.Called("n") {
		str = strings.TrimSpace(str)
	}
	fmt.Print(str)
	return nil
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
