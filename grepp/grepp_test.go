// This file is part of grepp.
//
// Copyright (C) 2012-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"io"
	"os"
	"testing"

	l "github.com/DavidGamba/dgtools/grepp/logging"
)

func TestMain(m *testing.M) {
	// l.LogInit(os.Stderr, os.Stderr, os.Stdout, os.Stderr, os.Stderr)
	l.LogInit(io.Discard, io.Discard, os.Stdout, os.Stderr, os.Stderr)
	os.Exit(m.Run())
}

func TestCheckPatternInFile(t *testing.T) {
	bufferSize = 2048
	cases := []struct {
		file       string
		pattern    string
		ignoreCase bool
		result     bool
	}{
		{"test_tree/A/b/C/d/E", "loreM", true, true},
		{"test_tree/A/b/C/d/E", "loreM", false, false},
		{"test_tree/A/b/C/d/E", "test", false, false},
		{"test_tree/A/b/C/d/E", "test", true, false},
	}
	for _, c := range cases {
		r, _ := checkPatternInFile(c.file, c.pattern, c.ignoreCase)
		if r != c.result {
			t.Errorf("checkPatternInFile(%q, %q, %v) == (%v), want (%v)",
				c.file, c.pattern, c.ignoreCase, r, c.result)
		}
	}
}

func TestGetRegex(t *testing.T) {
	cases := []struct {
		pattern    string
		ignoreCase bool
		line       string
		before     string
		match      string
		after      string
	}{
		{"pattern", true, "before pattern after", "before ", "pattern", " after"},
		{"(pattern)+", true, "before patternpattern after", "before ", "patternpattern", " after"},
		{"(pattern)(capture)(groups)", true, "before patterncapturegroups after", "before ", "patterncapturegroups", " after"},
	}
	for _, c := range cases {
		re, reEnd := getRegex(c.pattern, c.ignoreCase)
		match := re.FindAllStringSubmatch(c.line, -1)
		remainder := reEnd.FindStringSubmatch(c.line)
		if match[0][1] != c.before || match[0][2] != c.match || remainder[len(remainder)-1] != c.after {
			t.Errorf("TestGetRegex: expected %q, %q, %q | result match: %q, remainder: %q", c.before, c.match, c.after, match, remainder)
		}
	}
}
