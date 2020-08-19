// This file is part of joinlines.
//
// Copyright (C) 2017  David Gamba Rios
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

func synopsis() {
	synopsis := `<command_output> | joinlines <separator>

joinlines [--help]`
	fmt.Fprintln(os.Stderr, synopsis)
}

func main() {
	log.SetOutput(ioutil.Discard)
	opt := getoptions.New()
	opt.Bool("help", false)
	opt.Bool("debug", false)
	remaining, err := opt.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if opt.Called("help") {
		synopsis()
		os.Exit(1)
	}
	if opt.Called("debug") {
		log.SetOutput(os.Stderr)
	}
	log.Println(remaining)
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
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		str := string(bytes)
		str = strings.TrimSuffix(str, "\n")
		log.Println(str)
		fmt.Printf("%s\n", strings.Join(strings.Split(str, "\n"), separator))
	} else {
		synopsis()
		os.Exit(1)
	}
}
