// This file is part of diffdir.
//
// Copyright (C) 2020-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package main provides a wrapper around git diff --no-index to diff 2 directories.
*/
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(io.Discard, "", log.LstdFlags)

var version = "0.1.0"

var showHiddenDir, showHiddenFile bool

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Self("diffdir", "wrapper around 'git diff --no-index' to diff 2 directories.")
	opt.Bool("help", false, opt.Alias("?", "h"))
	opt.Bool("debug", false)
	opt.Bool("version", false, opt.Alias("V"))
	opt.BoolVar(&showHiddenDir, "hidden-dir", false, opt.Description("Descend into hidden dirs."))
	opt.BoolVar(&showHiddenFile, "hidden-file", false, opt.Alias("hidden"), opt.Description("Diff hidden files."))
	opt.HelpSynopsisArgs("<dir_a> <dir_b>")
	remaining, err := opt.Parse(args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
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
		Logger.SetOutput(os.Stderr)
	}
	Logger.Println(remaining)

	if len(remaining) < 2 {
		fmt.Fprintf(os.Stderr, "ERROR: Missing required arguments!\n")
		fmt.Println(opt.Help())
		return 1
	}
	seen := map[string]struct{}{}
	onlyA := []string{}

	err = diffDir(seen, &onlyA, remaining[0], remaining[0], remaining[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}

	for _, name := range onlyA {
		fmt.Printf("ONLY: %s\n", name)
	}
	checkBFiles(seen, remaining[1], remaining[1])

	return 0
}

func diffDir(seen map[string]struct{}, onlyA *[]string, aRoot, a, b string) error {
	// TODO: Ensure a and b are dirs

	Logger.Printf("Dir: %s\n", a)
	aFiles, err := ioutil.ReadDir(a)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", a, err)
	}
	for _, f := range aFiles {
		name := filepath.Join(a, f.Name())

		if f.IsDir() {
			if !showHiddenDir {
				// Ignore hidden dirs
				if strings.HasPrefix(f.Name(), ".") {
					continue
				}
			}
			diffDir(seen, onlyA, aRoot, name, b)
			continue
		}

		if !showHiddenFile {
			// Ignore hidden files
			if strings.HasPrefix(f.Name(), ".") {
				continue
			}
		}
		rel, err := filepath.Rel(aRoot, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to parse file '%s': %s\n", name, err)
		}
		seen[rel] = struct{}{}
		// fmt.Println(name)
		// fmt.Println(rel)
		bName := filepath.Join(b, rel)
		if _, err := os.Stat(bName); err != nil {
			*onlyA = append(*onlyA, name)
			continue
		}
		diffExternal(name, bName)
	}

	return nil
}

func checkBFiles(seen map[string]struct{}, bRoot, b string) error {
	bFiles, err := ioutil.ReadDir(b)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", b, err)
	}

	for _, f := range bFiles {
		name := filepath.Join(b, f.Name())

		if f.IsDir() {
			if !showHiddenDir {
				// Ignore hidden dirs
				if strings.HasPrefix(f.Name(), ".") {
					continue
				}
			}
			checkBFiles(seen, bRoot, name)
			continue
		}

		if !showHiddenFile {
			// Ignore hidden files
			if strings.HasPrefix(f.Name(), ".") {
				continue
			}
		}
		rel, err := filepath.Rel(bRoot, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to parse file '%s': %s\n", name, err)
		}
		if _, ok := seen[rel]; ok {
			continue
		}
		fmt.Printf("ONLY: %s\n", name)
	}
	return nil
}

func diffExternal(a, b string) error {
	Logger.Printf("Diff %s %s\n", a, b)
	gitOptions := []string{
		"--no-pager",
		"diff",
		"--no-index",
		"--color=always",
		a, b,
	}
	cmd := exec.Command("git", gitOptions...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}
