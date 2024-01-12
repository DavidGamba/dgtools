// This file is part of grepp.
//
// Copyright (C) 2012-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package main provides an improved version of the most common combinations of grep, find and sed in a single script.
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	greppLib "github.com/DavidGamba/dgtools/grepp/lib/grepp"
	l "github.com/DavidGamba/dgtools/grepp/logging"
	"github.com/DavidGamba/ffind/lib/ffind"
	"github.com/mgutz/ansi"
)

func checkPatternInFile(filename string, pattern string, ignoreCase bool) (bool, error) {
	re, _ := getRegex(pattern, ignoreCase)
	for le := range ReadLineByLine(filename, bufferSize) {
		if le.Error != nil {
			return false, le.Error
		}
		match := re.MatchString(string(le.Line))
		if match {
			return true, nil
		}
	}
	return false, nil
}

type lineMatch struct {
	filename string
	n        int
	match    [][]string
	end      []string
	line     string
}

func getRegex(pattern string, ignoreCase bool) (re, reEnd *regexp.Regexp) {
	if ignoreCase {
		re = regexp.MustCompile(`(?i)(.*?)(?P<pattern>` + pattern + `)`)
		reEnd = regexp.MustCompile(`(?i).*` + pattern + `(.*?)$`)
	} else {
		re = regexp.MustCompile(`(.*?)(?P<pattern>` + pattern + `)`)
		reEnd = regexp.MustCompile(`.*` + pattern + `(.*?)$`)
	}
	return
}

// TODO: Handle error properly here
func searchInFile(filename, pattern string, ignoreCase bool) <-chan lineMatch {
	c := make(chan lineMatch)
	re, reEnd := getRegex(pattern, ignoreCase)
	go func() {
		for le := range ReadLineByLine(filename, bufferSize) {
			if le.Error != nil {
				l.Error.Fatal(le.Error)
			}
			match := re.FindAllStringSubmatch(string(le.Line), -1)
			remainder := reEnd.FindStringSubmatch(string(le.Line))
			c <- lineMatch{filename: filename, n: le.LineNumber, line: string(le.Line), match: match, end: remainder}
		}
		close(c)
	}()
	return c
}

func color(color string, line string, useColor bool) string {
	if useColor {
		return fmt.Sprintf("%s%s", color, line)
	}
	return line
}

func colorReset(useColor bool) string {
	if useColor {
		return ansi.Reset
	}
	return ""
}

func (g grepp) writeLineMatch(file *os.File, lm lineMatch) {
	for _, m := range lm.match {
		replace := g.replace
		if strings.Contains(g.replace, `\1`) {
			if len(m) >= 4 {
				replace = strings.ReplaceAll(replace, `\1`, m[3])
			}
			if len(m) >= 5 {
				replace = strings.ReplaceAll(replace, `\2`, m[4])
			}
			if len(m) >= 6 {
				replace = strings.ReplaceAll(replace, `\3`, m[5])
			}
		}
		file.WriteString(m[1] + replace)
	}
	file.WriteString(lm.end[len(lm.end)-1] + "\n")
}

// Each section is in charge of starting with the color or reset.
func (g grepp) printLineMatch(lm lineMatch) {
	stringLine := func() string {
		if g.useColor {
			result := ansi.Reset
			l.Debug.Printf("[printLineMatch] %#v\n", lm)
			for _, m := range lm.match {
				replace := g.replace
				if strings.Contains(g.replace, `\1`) {
					if len(m) >= 4 {
						replace = strings.ReplaceAll(replace, `\1`, m[3])
					}
					if len(m) >= 5 {
						replace = strings.ReplaceAll(replace, `\2`, m[4])
					}
					if len(m) >= 6 {
						replace = strings.ReplaceAll(replace, `\3`, m[5])
					}
					l.Debug.Printf("[printLineMatch] replace: %s\n", replace)
				}
				result += fmt.Sprintf("%s%s%s%s%s%s",
					stripCtlFromUTF8(m[1]),
					ansi.Red,
					stripCtlFromUTF8(m[2]),
					ansi.Green,
					stripCtlFromUTF8(replace),
					ansi.Reset)
			}
			result += stripCtlFromUTF8(lm.end[len(lm.end)-1])
			return result
		}
		return stripCtlFromUTF8(lm.line)
	}

	result := ""
	if g.showFile {
		result += color(ansi.Magenta, lm.filename, g.useColor) + " " + color(ansi.Blue, ":", g.useColor)
	}
	if g.useNumber {
		result += color(ansi.Green, strconv.Itoa(lm.n), g.useColor) + color(ansi.Blue, ":", g.useColor)
	}
	result += colorReset(g.useColor) + " " + stringLine()
	fmt.Fprintln(g.Stdout, result)
}

// Each section is in charge of starting with the color or reset.
func (g grepp) printMinorWarning(line string) {
	result := color(ansi.LightBlack, line, g.useColor)
	fmt.Fprintln(g.Stderr, result)
}

// Each section is in charge of starting with the color or reset.
func (g grepp) printLineContext(lm lineMatch) {
	result := ""
	if g.showFile {
		result += color(ansi.Magenta, lm.filename, g.useColor) + " " + color(ansi.Blue, "-", g.useColor)
	}
	if g.useNumber {
		result += color(ansi.Green, strconv.Itoa(lm.n), g.useColor) + color(ansi.Blue, "-", g.useColor)
	}
	result += colorReset(g.useColor) + " " + lm.line
	fmt.Fprintln(g.Stdout, result)
}

type grepp struct {
	ignoreBinary  bool
	caseSensitive bool
	useColor      bool
	useNumber     bool
	filenameOnly  bool
	replace       string
	force         bool
	context       int
	searchBase    string
	// Controls whether or not to show the filename. If the given location is a
	// file then there is no need to show the filename
	showFile             bool
	showBufferSizeErrors bool
	bufferSizeErrorsC    int
	pattern              string
	filePattern          string
	ignoreFilePattern    string
	ignoreExtensionList  []string
	Stdout               io.Writer
	Stderr               io.Writer
}

func (g grepp) String() string {
	return fmt.Sprintf("ignoreBinary: %v, caseSensitive: %v, useColor %v, useNumber %v, filenameOnly %v, force %v",
		g.ignoreBinary, g.caseSensitive, g.useColor, g.useNumber, g.filenameOnly, g.force)
}

func (g grepp) getFileList() <-chan ffind.FileError {
	c := make(chan ffind.FileError)
	go func() {
		if g.showFile {
			ch := ffind.ListRecursive(
				g.searchBase,
				true,
				&ffind.BasicFileMatch{
					IgnoreDirResults:        true,
					IgnoreFileResults:       false,
					IgnoreVCSDirs:           true,
					IgnoreHidden:            true,
					IgnoreFileExtensionList: g.ignoreExtensionList,
				},
				ffind.SortFnByName)
			for e := range ch {
				if e.Error != nil {
					fmt.Fprintf(os.Stderr, "ERROR: '%s' %s\n", e.Path, e.Error)
					// Ignore broken symlinks
					if os.IsNotExist(e.Error) {
						continue
					}
				}
				c <- e
			}
		} else {
			// TODO: FileInfo is not generated here. Check if it is needed.
			c <- ffind.FileError{
				Path: g.searchBase,
			}
		}
		close(c)
	}()
	return c
}

func (g grepp) Run(ctx context.Context) error {
	for ch := range g.getFileList() {
		filename := ch.Path
		if g.ignoreBinary && !greppLib.IsTextMIME(filename) {
			continue
		}
		if g.filenameOnly {
			ok, err := checkPatternInFile(filename, g.pattern, !g.caseSensitive)
			if err != nil {
				if errors.Is(err, errorBufferSizeTooSmall) {
					if g.showBufferSizeErrors {
						g.printMinorWarning(fmt.Sprintf("%s : %s\n", filename, err.Error()))
					} else {
						g.bufferSizeErrorsC++
					}
				} else {
					fmt.Fprintf(g.Stderr, "%s\n", err)
				}
			} else if ok {
				fmt.Fprintf(g.Stdout, "%s%s\n", color(ansi.Magenta, filename, g.useColor), colorReset(g.useColor))
			}
		} else {
			ok, err := checkPatternInFile(filename, g.pattern, !g.caseSensitive)
			if err != nil {
				if errors.Is(err, errorBufferSizeTooSmall) {
					if g.showBufferSizeErrors {
						g.printMinorWarning(fmt.Sprintf("%s : %s\n", filename, err.Error()))
					} else {
						g.bufferSizeErrorsC++
					}
				} else {
					fmt.Fprintf(g.Stderr, "%s\n", err)
				}
			} else if ok {
				var tmpFile *os.File
				var err error
				if g.force {
					tmpFile, err = os.CreateTemp("", filepath.Base(filename)+"-")
					defer tmpFile.Close()
					if err != nil {
						l.Error.Println("cannot open ", tmpFile)
						l.Error.Fatal(err)
					}
					l.Debug.Printf("tmpFile: %v", tmpFile.Name())
				}
				for d := range searchInFile(filename, g.pattern, !g.caseSensitive) {
					if len(d.match) == 0 {
						if g.context > 0 {
							g.printLineContext(d)
						}
					} else {
						g.printLineMatch(d)
					}
					if g.force {
						if len(d.match) == 0 {
							tmpFile.WriteString(d.line + "\n")
						} else {
							g.writeLineMatch(tmpFile, d)
						}
					}
				}
				if g.force {
					tmpFile.Close()
					err = copyFileContents(tmpFile.Name(), filename)
					if err != nil {
						l.Warning.Printf("Couldn't update file: %s. '%s'\n", filename, err)
					}
				}
			}
		}
	}

	if g.bufferSizeErrorsC > 0 {
		fmt.Fprintf(g.Stderr, "WARNING: %s found %d times\n", errorBufferSizeTooSmall, g.bufferSizeErrorsC)
	}
	return nil
}

func (g *grepp) SetStderr(w io.Writer) {
	l.Warning.SetOutput(w)
	l.Error.SetOutput(w)
	g.Stderr = w
}

func (g *grepp) SetStdout(w io.Writer) {
	l.Info.SetOutput(w)
	g.Stdout = w
}
