// This file is part of go-utils.
//
// Copyright (C) 2019  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/DavidGamba/dgtools/yamlutils"

	"github.com/DavidGamba/go-getoptions"
)

// BuildMetadata - Provides the metadata part of the version information.
var BuildMetadata = "dev"

const semVersion = "0.5.0"

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

func main() {
	var file string
	var include bool
	var add string
	var keys []string
	opt := getoptions.New()
	opt.Self("", `Parses YAML input passed from file or piped to STDIN and filters it by key or index.

    Source: https://github.com/DavidGamba/go-utils`)
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false)
	opt.Bool("version", false, opt.Alias("V"))
	opt.Bool("n", false, opt.Description("Remove trailing spaces."))
	opt.Bool("silent", false, opt.Description("Don't print full context errors."))
	opt.BoolVar(&include, "include", false, opt.Description("Include parent key if it is a map key."))
	opt.StringVar(&file, "file", "", opt.Alias("f"), opt.ArgName("file"), opt.Description("YAML file to read."))
	opt.StringVar(&add, "add", "", opt.ArgName("yaml/json input"), opt.Description("Child input to add at the current location."))
	opt.StringSliceVar(&keys, "key", 1, 99, opt.Alias("k"), opt.ArgName("key/index"),
		opt.Description(`Key or index to descend to.
Multiple keys allow to descend further.
Indexes are positive integers.`))
	_, err := opt.Parse(os.Args[1:])
	if opt.Called("help") {
		fmt.Fprintln(os.Stderr, opt.Help())
		os.Exit(1)
	}
	if opt.Called("version") {
		fmt.Printf("Version: %s+%s\n", semVersion, BuildMetadata)
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if opt.Called("debug") {
		logger.SetOutput(os.Stderr)
		yamlutils.Logger.SetOutput(os.Stderr)
	}
	var xpath []string
	for _, k := range keys {
		xpath = append(xpath, strings.Split(k, "/")...)
	}
	logger.Printf("path: '%s'\n", strings.Join(xpath, ","))

	// Check if stdin is pipe p or device D
	statStdin, _ := os.Stdin.Stat()
	stdinIsDevice := (statStdin.Mode() & os.ModeDevice) != 0

	var yml *yamlutils.YML
	if !stdinIsDevice && !opt.Called("file") {
		logger.Printf("Reading from stdin\n")
		reader := os.Stdin
		yml, err = yamlutils.NewFromReader(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: reading yaml from STDIN: %s\n", err)
			os.Exit(1)
		}
	} else {
		logger.Printf("Reading from file: %s\n", file)
		if !opt.Called("file") {
			fmt.Fprintf(os.Stderr, "ERROR: missing argument '--file <file>'\n")
			os.Exit(1)
		}
		yml, err = yamlutils.NewFromFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: reading yaml file: %s\n", err)
			os.Exit(1)
		}
	}

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
		return
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
}
