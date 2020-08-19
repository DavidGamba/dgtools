// This file is part of joinlines.
//
// Copyright (C) 2017-2020  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package main provides an utility to join lines in the command line.
*/
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/DavidGamba/go-getoptions"
)

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

var version = "0.2.0"

func main() {
	os.Exit(program())
}

func program() int {
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false)
	opt.Bool("version", false, opt.Alias("V"))
	opt.Self("", "Simple utility to join lines from a command output.")
	remaining, err := opt.Parse(os.Args[1:])
	if opt.Called("help") {
		help(opt)
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("version") {
		fmt.Println(version)
		return 0
	}
	if opt.Called("debug") {
		logger.SetOutput(os.Stderr)
	}
	logger.Println(remaining)

	separator := " "
	if len(remaining) >= 1 {
		separator = remaining[0]
	}

	// Check if stdin is pipe p or device D
	statStdin, _ := os.Stdin.Stat()
	stdinIsDevice := (statStdin.Mode() & os.ModeDevice) != 0

	if !stdinIsDevice {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return 1
		}
		str := string(bytes)
		str = strings.TrimSuffix(str, "\n")
		logger.Println(str)
		fmt.Printf("%s\n", strings.Join(strings.Split(str, "\n"), separator))
	} else {
		help(opt)
		return 1
	}

	return 0
}

func help(opt *getoptions.GetOpt) {
	fmt.Fprintln(os.Stderr, opt.Help(getoptions.HelpName))
	fmt.Fprintf(os.Stderr, `SYNOPSIS:
	# Pipe output from another command
	<command_output> | joinlines [<separator>]

	joinlines [--help]

`)
	fmt.Fprintln(os.Stderr, opt.Help(getoptions.HelpOptionList))
}
