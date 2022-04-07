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

	"github.com/DavidGamba/dgtools/clitable"
	"github.com/DavidGamba/go-getoptions"
)

// BuildMetadata - Provides the metadata part of the version information.
//
//   go build -ldflags="-X main.BuildMetadata=`date +'%Y%m%d%H%M%S'`.`git rev-parse --short HEAD`"
var BuildMetadata = "dev"

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
		fmt.Printf("Version: %s+%s\n", semVersion, BuildMetadata)
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
	err = clitable.NewTablePrinter().HasHeader(*header).FprintCSVReader(os.Stdout, reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
