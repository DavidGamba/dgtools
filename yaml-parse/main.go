// This file is part of dgtools.
//
// # Copyright (C) 2019-2023  David Gamba Rios
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

var Logger = log.New(ioutil.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Self("", `Parses YAML input passed from file or piped to STDIN and filters it by key or index.

         Source: https://github.com/DavidGamba/dgtools`)
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false)
	opt.Bool("version", false, opt.Alias("V"))
	opt.Bool("n", false, opt.Description("Remove trailing spaces."))
	opt.Bool("silent", false, opt.Description("Don't print full context errors."))
	opt.Bool("include", false, opt.Description("Include parent key if it is a map key."))
	opt.String("file", "", opt.Alias("f"), opt.ArgName("file"), opt.Description("YAML file to read."))
	opt.String("add", "", opt.ArgName("yaml/json input"), opt.Description("Child input to add at the current location."))
	opt.StringSlice("key", 1, 99, opt.Alias("k"), opt.ArgName("key/index"),
		opt.Description(`Key or index to descend to.
Multiple keys allow to descend further.
Indexes are positive integers.`))
	opt.IntOptional("document", 1, opt.Description("Document number"), opt.Alias("d"), opt.ArgName("number"))
	_, err := opt.Parse(os.Args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		return 1
	}
	if opt.Called("version") {
		fmt.Printf("Version: %s+%s\n", semVersion, BuildMetadata)
		return 0
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("debug") {
		Logger.SetOutput(os.Stderr)
		yamlutils.Logger.SetOutput(os.Stderr)
	}

	err = realMain(opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func realMain(opt *getoptions.GetOpt) error {
	file := opt.Value("file").(string)
	include := opt.Value("include").(bool)
	add := opt.Value("add").(string)
	keys := opt.Value("key").([]string)

	var xpath []string
	for _, k := range keys {
		k = strings.TrimLeft(k, "/")
		xpath = append(xpath, strings.Split(k, "/")...)
	}
	Logger.Printf("path: '%s'\n", strings.Join(xpath, ","))

	// Check if stdin is pipe p or device D
	statStdin, _ := os.Stdin.Stat()
	stdinIsDevice := (statStdin.Mode() & os.ModeDevice) != 0

	var err error
	var ymlList []*yamlutils.YML
	if !stdinIsDevice && !opt.Called("file") {
		Logger.Printf("Reading from stdin\n")
		reader := os.Stdin
		ymlList, err = yamlutils.NewFromReader(reader)
		if err != nil {
			return fmt.Errorf("reading yaml from STDIN: %w", err)
		}
	} else {
		Logger.Printf("Reading from file: %s\n", file)
		if !opt.Called("file") {
			return fmt.Errorf("missing argument '--file <file>'")
		}
		ymlList, err = yamlutils.NewFromFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: reading yaml file: %s\n", err)
			os.Exit(1)
		}
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
