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

var logger = log.New(ioutil.Discard, "main DEBUG ", log.LstdFlags)

func examples() {
	fmt.Fprintf(os.Stderr, `EXAMPLES:
    # Read CSV file
    csvtable <csv_filename>

    # Pipe CSV file to csvtable
    cat <csv_filename> | csvtable
`)
}

func main() {
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false)
	opt.Bool("version", false, opt.Alias("V"))
	opt.Bool("tsv", false)
	header := opt.Bool("no-header", true)
	opt.HelpSynopsisArgs("<csv_filename>")
	remaining, err := opt.Parse(os.Args[1:])
	if opt.Called("help") {
		fmt.Fprint(os.Stderr, opt.Help())
		examples()
		os.Exit(1)
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
		logger.SetOutput(os.Stderr)
		clitable.Logger.SetOutput(os.Stderr)
	}
	logger.Println(remaining)

	var reader io.Reader
	if len(remaining) < 1 {
		// Check if stdin is pipe p or device D
		statStdin, _ := os.Stdin.Stat()
		stdinIsDevice := (statStdin.Mode() & os.ModeDevice) != 0

		if stdinIsDevice {
			fmt.Fprint(os.Stderr, opt.Help())
			examples()
			os.Exit(1)
		}
		logger.Printf("Reading from stdin\n")
		reader = os.Stdin
	} else {
		filename := remaining[0]
		logger.Printf("Reading from file %s\n", filename)
		fh, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
		defer fh.Close()
		reader = fh
	}
	tp := clitable.NewTablePrinter()
	if opt.Called("tsv") {
		tp.Separator('\t')
	}
	err = tp.HasHeader(*header).FprintCSVReader(os.Stdout, reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
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
