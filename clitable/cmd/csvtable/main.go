// This file is part of clitable.
//
// Copyright (C) 2017-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package csvtable provides a tool to view csv files on the cmdline.

	┌──┬──┐
	│  │  │
	├──┼──┤
	└──┴──┘
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/DavidGamba/dgtools/clitable"
	"github.com/DavidGamba/go-getoptions"
)

const semVersion = "0.3.0"

var Logger = log.New(ioutil.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func examples() {
	fmt.Fprintf(os.Stderr, `EXAMPLES:
    # Read CSV file
    csvtable <csv_filename>

    # Pipe CSV file to csvtable
    cat <csv_filename> | csvtable

    # Read TSV file
    csvtable <tsv_filename> --tsv
`)
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Bool("debug", false)
	opt.Bool("version", false, opt.Alias("V"))
	opt.Bool("tsv", false)
	opt.Bool("no-header", true)
	opt.HelpSynopsisArg("<filename>", "CSV|TSV file to read")

	opt.SetCommandFn(Run)
	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if opt.Called("version") {
		v, err := version(semVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Version: %s\n", v)
		os.Exit(0)
	}
	if opt.Called("debug") {
		Logger.SetOutput(os.Stderr)
		clitable.Logger.SetOutput(os.Stderr)
	}
	Logger.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			examples()
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	header := opt.Value("no-header").(bool)

	var reader io.Reader
	if len(args) < 1 {
		// Check if stdin is pipe p or device D
		statStdin, _ := os.Stdin.Stat()
		stdinIsDevice := (statStdin.Mode() & os.ModeDevice) != 0

		if stdinIsDevice {
			fmt.Fprint(os.Stderr, opt.Help())
			examples()
			os.Exit(1)
		}
		Logger.Printf("Reading from stdin\n")
		reader = os.Stdin
	} else {
		filename := args[0]
		Logger.Printf("Reading from file %s\n", filename)
		fh, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer fh.Close()
		reader = fh
	}
	tp := clitable.NewTablePrinter()
	if opt.Called("tsv") {
		tp.Separator('\t')
	}
	err := tp.HasHeader(header).FprintCSVReader(os.Stdout, reader)
	if err != nil {
		return err
	}
	return nil
}

func version(semVersion string) (string, error) {
	var revision, timeStr, modified string
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
			case "vcs.time":
				vcsTime := s.Value
				date, err := time.Parse("2006-01-02T15:04:05Z", vcsTime)
				if err != nil {
					return "", fmt.Errorf("failed to parse time: %w", err)
				}
				timeStr = date.Format("20060102_150405")
			case "vcs.modified":
				if s.Value == "true" {
					modified = "modified"
				}
			}
		}
	}
	if revision != "" && timeStr != "" {
		semVersion += fmt.Sprintf("+%s.%s", revision, timeStr)
		if modified != "" {
			semVersion += fmt.Sprintf(".%s", modified)
		}
	}
	return semVersion, nil
}
