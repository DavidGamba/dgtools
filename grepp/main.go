// This file is part of grepp.
//
// Copyright (C) 2012-2024  David Gamba Rios
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
	"os"

	l "github.com/DavidGamba/dgtools/grepp/logging"
	"github.com/DavidGamba/dgtools/grepp/runInPager"
	"github.com/DavidGamba/dgtools/grepp/semver"
	"github.com/DavidGamba/go-getoptions"
)

// Buffer Size used to read files when searching through them.
// Default value should cover most cases.
var bufferSize int

var g grepp

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	l.LogInit(io.Discard, io.Discard, os.Stdout, os.Stderr, os.Stderr)

	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.SetCommandFn(Run)

	opt.Bool("version", false) // version info
	opt.BoolVar(&g.ignoreBinary, "ignore-binary", true, opt.Alias("I"))
	opt.BoolVar(&g.caseSensitive, "case-sensitive", false, opt.Alias("c"))

	// TODO:
	// opt.StringVar(&g.useColor, "color", "auto")
	opt.BoolVar(&g.useColor, "color", true)
	opt.BoolVar(&g.useNumber, "line-number", true, opt.Alias("n"))
	opt.BoolVar(&g.filenameOnly, "files-with-matches", false, opt.Alias("l"))
	opt.StringVar(&g.replace, "replace", "", opt.Alias("r"), opt.Description(`Replace matches with the given text. Use \1, \2, \3 to replace captures`))
	opt.BoolVar(&g.force, "force", false, opt.Alias("f"))
	opt.IntVar(&g.context, "context", 0, opt.Alias("C"), opt.Description("Number of lines of context to show"))
	opt.IntVar(&bufferSize, "buffer", 16384)
	opt.BoolVar(&g.showBufferSizeErrors, "show-buffer-errors", false, opt.Alias("sbe"))
	opt.Bool("no-pager", false)
	opt.StringSlice("ignore-extension", 1, 1, opt.Alias("ie"))
	// "fp"      // fullPath - Used to show the file full path instead of the relative to the current dir.
	// "name"    // filePattern - Use to further filter the search to files matching that pattern.
	// "ignore"  // ignoreFilePattern - Use to further filter the search to files not matching that pattern.
	// "spacing" // keepSpacing - Do not remove initial spacing.
	// "no-page" // Don't use pager for output

	opt.Bool("debug", false, opt.Description("Enable debug logging"))
	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("version") {
		version := semver.Version{Major: 0, Minor: 2, Patch: 0, PreReleaseLabel: "dev"}
		fmt.Println(version)
		return 0
	}
	if opt.Called("debug") {
		l.Debug.SetOutput(os.Stderr)
		l.Trace.SetOutput(os.Stderr)
	}
	l.Debug.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		if errors.Is(err, getoptions.ErrorParsing) {
			fmt.Fprintf(os.Stderr, "\n"+opt.Help())
		}
		return 1
	}
	return 0
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	noPager := opt.Value("no-pager").(bool)
	ie := opt.Value("ignore-extension").([]string)

	// TODO: Read from ~/.grepprc
	g.ignoreExtensionList = []string{
		".un~",     // vim
		".swp",     // vim
		".a",       // c
		".so",      // c
		".db",      // Database file
		".base64",  // encoded_data
		".svg",     // image
		".png",     // image
		".PNG",     // image
		".jpg",     // image
		".ttf",     // font
		".pdf",     // pdf
		".tfstate", // terraform state
	}
	g.ignoreExtensionList = append(g.ignoreExtensionList, ie...)

	pattern, args, err := opt.GetRequiredArg(args)
	if err != nil {
		return err
	}
	g.pattern = pattern
	g.searchBase = "."
	if len(args) > 0 {
		g.searchBase = args[0]
	}
	searchBaseInfo, err := os.Stat(g.searchBase)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", g.searchBase, err)
	}
	if searchBaseInfo.IsDir() {
		g.showFile = true
	} else {
		g.showFile = false
		// If filename provided, don't skip it
		g.ignoreBinary = false
	}

	l.Debug.Printf("pattern: %s, searchBase: %s, replace: %s", g.pattern, g.searchBase, g.replace)
	l.Debug.Printf(fmt.Sprintln(g))

	isDevice := isDevice()
	if !noPager && isDevice {
		l.Debug.Println("runInPager")
		err = runInPager.Command(ctx, &g)
	} else if noPager && isDevice {
		g.Stdout = os.Stdout
		g.Stderr = os.Stderr
		err = g.Run(ctx)
	} else {
		g.useColor = false
		g.Stdout = os.Stdout
		g.Stderr = os.Stderr
		err = g.Run(ctx)
	}
	return err
}

func isDevice() bool {
	statStdout, _ := os.Stdout.Stat()
	return (statStdout.Mode() & os.ModeDevice) != 0
}
