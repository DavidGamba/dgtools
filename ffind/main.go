// This file is part of ffind.
//
// Copyright (C) 2017-2025  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package main provides a utility to find files on the command line.

Goals:

• Return a file list given a dir.

•  Return the given file given a file.

• Do case insensitive (by default) or sensitive file matching.

• Allow to return files or dirs only.
Maybe build a list of common extensions in the skip code to allow for groups.
For example: '.rb' and '.erb' for ruby files.

• Follow Symlinks.
  • Is there a case where you don't want to? Allow disabling the follow anyway.

• Ignore hidden files (configurable).

  • In windows?

  • In Linux, ignore files starting with .

• Ignore git, svn and mercurial files (configurable).

TODO: Look into adding option to ignore reporting broken symlinks.

TODO: Implement version sort.

TODO: Limit depth

*/

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/DavidGamba/dgtools/ffind/lib/ffind"
	"github.com/DavidGamba/dgtools/ffind/semver"
	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(io.Discard, "", log.LstdFlags)

func synopsis() {
	synopsis := `USAGE:
        # List all files under current dir
        ffind [OPTIONS...]

        # List file_pattern matching files under current dir
        # <file_pattern> can't end in /
        ffind <file_pattern> [OPTIONS...]

        # List file pattern matching files under given dir
        ffind <file_pattern> <dir> [OPTIONS...]

        # List all files under given dir
        ffind <dir>/ [OPTIONS...]

        # List file pattern matching files under given dir
        ffind <dir>/ <file_pattern> [OPTIONS...]

HELP:
        ffind --type-list|--typelist # Show type list
        ffind --version              # Show version
        ffind -h|-?|--help           # shows short help

TODO:
        [-D|--no-dir <dir_to_ignore>]
        [--color <never|auto|always>]
`
	fmt.Fprintln(os.Stderr, synopsis)
}

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	var vcs, hidden, caseSensitive, follow, abspath bool
	var sortNum, typeDir, typeFile bool
	var fileType []string
	opt := getoptions.New()
	opt.SetMode(getoptions.SingleDash)
	opt.Bool("help", false, opt.Alias("?"), opt.Alias("h"))
	opt.Bool("version", false)
	opt.Bool("debug", false)
	opt.Bool("verbose", false)
	opt.Bool("type-list", false, opt.Alias("typelist"))
	opt.BoolVar(&vcs, "vcs", true, opt.Description("Sets --hidden when set."))
	opt.BoolVar(&hidden, "hidden", true)
	opt.BoolVar(&caseSensitive, "case", false)
	opt.BoolVar(&follow, "no-follow", true)
	opt.BoolVar(&abspath, "abs-path", false)
	opt.BoolVar(&sortNum, "num-sort", false)
	fileTypeWithFileAndDir := opt.StringSlice("t", 1, 1, opt.Alias("type"))
	noFileType := opt.StringSlice("T", 1, 1, opt.Alias("no-type"))
	matchExtensionList := opt.StringSlice("e", 1, 1, opt.Alias("extension"))
	ignoreExtensionList := opt.StringSlice("E", 1, 1, opt.Alias("no-extension"))
	remaining, err := opt.Parse(args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		synopsis()
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("version") {
		version := semver.Version{Major: 0, Minor: 6, Patch: 2}
		fmt.Println(version)
		return 1
	}
	if opt.Called("type-list") {
		ffind.PrintTypeList()
		return 0
	}
	if opt.Called("debug") {
		Logger.SetOutput(os.Stderr)
	}
	Logger.Println(remaining)

	// ctx, cancel, done := getoptions.InterruptContext()
	// defer func() { cancel(); <-done }()

	if opt.Called("t") || opt.Called("T") {
		Logger.Printf("type: %v, no-type %v\n", *fileTypeWithFileAndDir, *noFileType)
		for _, t := range *fileTypeWithFileAndDir {
			switch t {
			case "f":
				typeFile = true
			case "d":
				typeDir = true
			default:
				if !ffind.KnownFileType(t) {
					fmt.Fprintf(os.Stderr, "ERROR: Provided --type is not valid '%s'\n", t)
					return 1
				}
				fileType = append(fileType, t)
			}
		}
		for _, t := range *noFileType {
			if !ffind.KnownFileType(t) {
				fmt.Fprintf(os.Stderr, "ERROR: Provided --type is not valid '%s'\n", t)
				return 1
			}
		}
	}

	var filePattern string
	dir := "."
	switch len(remaining) {
	case 0:
		filePattern = "."
	case 1:
		if strings.HasSuffix(remaining[0], string(os.PathSeparator)) {
			Logger.Println("Assume dir")
			filePattern = "."
			dir = remaining[0]
		} else {
			filePattern = remaining[0]
		}
	case 2:
		if strings.HasSuffix(remaining[0], string(os.PathSeparator)) {
			Logger.Println("Assume dir")
			dir = remaining[0]
			filePattern = remaining[1]
		} else {
			filePattern = remaining[0]
			dir = remaining[1]
		}
	}

	var absdir string
	dir = filepath.Clean(dir)
	filePattern = filepath.Clean(filePattern)
	if abspath {
		absdir, err = filepath.Abs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return 1
		}
		absdir = filepath.Dir(absdir)
	}

	Logger.Printf("dir: %s\n", dir)
	Logger.Printf("filePattern: %s\n", filePattern)
	Logger.Printf("Ext: %v\n", ignoreExtensionList)
	var r *regexp.Regexp
	if caseSensitive {
		r, err = regexp.Compile(filePattern)
	} else {
		r, err = regexp.Compile("(?i)" + filePattern)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: with provided file pattern %s\n", err)
		return 1
	}

	var sfn ffind.SortFn
	if sortNum {
		sfn = ffind.SortFnByNum
	} else {
		sfn = ffind.SortFnByName
	}

	ch := ffind.ListRecursive(
		dir,
		follow,
		&ffind.BasicFileMatch{
			IgnoreDirResults:        typeFile,
			IgnoreFileResults:       typeDir,
			IgnoreVCSDirs:           vcs,
			IgnoreHidden:            hidden,
			IgnoreFileExtensionList: *ignoreExtensionList,
			IgnoreFileTypeList:      *noFileType,
			MatchFileExtensionList:  *matchExtensionList,
			MatchFileTypeList:       fileType,
		},
		sfn)
	for e := range ch {
		if e.Error != nil {
			fmt.Fprintf(os.Stderr, "ERROR: '%s' %s\n", e.Path, e.Error)
			if os.IsNotExist(e.Error) {
				continue
			}
		}
		Logger.Printf("ffind: %s\n", e.Path)
		if r.MatchString(filepath.Base(e.Path)) {
			if abspath {
				fmt.Println(filepath.Join(absdir, e.Path))
			} else {
				fmt.Println(e.Path)
			}
		}
	}
	return 0
}
