// This file is part of ffind.
//
// Copyright (C) 2017-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffind

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func goToRootDir() {
	log.SetOutput(os.Stderr)
	cwd, _ := os.Getwd()
	log.Printf("CWD: %s", cwd)
	if strings.HasSuffix(cwd, "lib/ffind") {
		// Go to the root of the repository
		_ = os.Chdir("../..")
	}
	cwd, _ = os.Getwd()
	log.Printf("CWD: %s", cwd)
	log.SetOutput(io.Discard)
}

func setupLogging() *bytes.Buffer {
	s := ""
	buf := bytes.NewBufferString(s)
	Logger.SetOutput(buf)
	return buf
}

func compareTestStringSlices(t *testing.T, expected []string, received []string) {
	if !reflect.DeepEqual(received, expected) {
		str := ""
		for i := range expected {
			switch {
			case len(received) <= i:
				str += fmt.Sprintf("%s\n", expected[i])
			case expected[i] != received[i]:
				str += fmt.Sprintf("%s != %s\n", expected[i], received[i])
			default:
				str += fmt.Sprintf("%s == %s\n", expected[i], received[i])
			}
		}
		if len(received) > len(expected) {
			for i := len(expected); i < len(received); i++ {
				str += fmt.Sprintf(" == %s\n", received[i])
			}
		}
		t.Errorf("trees differ: expected (%d) received (%d)\n%s\n", len(expected), len(received), str)
		return
	}
}

func TestNewFileError(t *testing.T) {
	buf := setupLogging()
	file := "test_files/test_tree"

	fsys := fstest.MapFS{
		"test_files/test_tree/a": {},
		"test_files/test_tree/b": {},
		"test_files/test_tree/c": {},
	}

	fe, err := NewFileError(fsys, file)
	if err != nil {
		t.Errorf("Unexpected error '%s': %s\n", file, err)
		t.Log(buf.String())
		return
	}
	if fe.Path != "test_files/test_tree" {
		t.Errorf("Paths don't match '%s': %s\n", file, "test_files/test_tree")
		t.Log(buf.String())
		return
	}
	file = "non-existent"
	fe, err = NewFileError(fsys, file)
	if err == nil {
		t.Errorf("Expected error, none received!\n")
		t.Log(buf.String())
		return
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected IsNotExist error, received: %s\n", err)
		t.Log(buf.String())
		return
	}
	Logger.Printf("%#v\n", fe)

	t.Log(buf.String())
}

// Make sure filepath.EvalSymlinks behaves as expected.
func TestEvalSymlink(t *testing.T) {
	log.SetOutput(io.Discard)
	goToRootDir()
	buf := setupLogging()

	cases := []struct {
		file     string
		expected string
	}{
		{"./test_files/test_tree", "test_files/test_tree"},
		{"./test_files/test_tree/.A", "test_files/test_tree/.A"},
		{"./test_files/test_tree/.A/b/C/d/E", "test_files/test_tree/.A/b/C/d/E"},
		{"./test_files/test_tree/slnA", "test_files/test_tree/A"},
		{"./test_files/test_tree/slnB", "test_files/test_tree/A"},
	}
	for _, c := range cases {
		// Read given file information
		read, err := filepath.EvalSymlinks(c.file)
		if err != nil {
			t.Errorf("Unexpected error '%s': %s\n", c.file, err)
			t.Log(buf.String())
			return
		}
		if read != c.expected {
			t.Errorf("Got %s != Expected %s\n", read, c.expected)
			t.Log(buf.String())
			return
		}
	}
	_, err := filepath.EvalSymlinks("./test_files/test_tree/slnE")
	if err == nil {
		t.Errorf("Expected error, none received! \n")
		t.Log(buf.String())
		return
	}
	if !strings.Contains(err.Error(), "too many links") {
		t.Errorf("Unexpected error: %s\n", err)
		t.Log(buf.String())
		return
	}
}

func TestListOneLevel(t *testing.T) {
	log.SetOutput(io.Discard)
	// log.SetOutput(os.Stderr)
	goToRootDir()
	buf := setupLogging()
	cases := []struct {
		file     string
		follow   bool
		expected []string
	}{
		{"test_files/test_tree", false, []string{
			"test_files/test_tree/.A",
			"test_files/test_tree/.a",
			"test_files/test_tree/.hg",
			"test_files/test_tree/.svn",
			"test_files/test_tree/A",
			"test_files/test_tree/a",
			"test_files/test_tree/slnA",
			"test_files/test_tree/slnB",
			"test_files/test_tree/slnC",
			"test_files/test_tree/slnD",
			"test_files/test_tree/slnE",
			"test_files/test_tree/slnF",
			"test_files/test_tree/slnG",
			"test_files/test_tree/z",
		},
		},
		{"test_files/test_tree/.A", false, []string{
			"test_files/test_tree/.A/b",
		},
		},
		{"test_files/test_tree/.A/b/C/d/E", false, []string{
			"test_files/test_tree/.A/b/C/d/E",
		},
		},
		{"test_files/test_tree/slnA", false, []string{
			"test_files/test_tree/slnA",
		},
		},
		{"test_files/test_tree/slnA", true, []string{
			"test_files/test_tree/slnA/b",
		},
		},
		{"test_files/test_tree/slnB", false, []string{
			"test_files/test_tree/slnB",
		},
		},
		{"test_files/test_tree/slnB", true, []string{
			"test_files/test_tree/slnB/b",
		},
		},
	}
	for _, c := range cases {
		ch := ListOneLevel(c.file, c.follow, SortFnByName)
		tree := []string{}
		for e := range ch {
			if e.Error != nil {
				t.Errorf("Unexpected error: %s\n", e.Error)
				t.Log(buf.String())
				return
			}
			tree = append(tree, e.Path)
		}
		compareTestStringSlices(t, c.expected, tree)
	}
	file := "./test_files/test_tree/slnE"
	ch := ListOneLevel(file, true, SortFnByName)
	log.Printf("%d\n", len(ch))
	for e := range ch {
		if e.Error == nil {
			t.Errorf("Expected error, none received! \n")
			t.Log(buf.String())
			return
		}
	}
	file = "tmp-no-permissions"
	err := os.Mkdir(file, 0000)
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
		t.Log(buf.String())
		return
	}
	ch = ListOneLevel(file, false, SortFnByName)
	tree := []string{}
	for e := range ch {
		if e.Error == nil {
			t.Errorf("Expected error, none received! \n")
		}
		if !os.IsPermission(e.Error) {
			t.Errorf("Expected IsPermission error, received: %s\n", e.Error.Error())
		}
		tree = append(tree, e.Path)
	}
	err = os.RemoveAll(file)
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
		t.Log(buf.String())
		return
	}
	compareTestStringSlices(t, []string{file}, tree)
	file = "non-existent"
	ch = ListOneLevel(file, false, SortFnByName)
	for e := range ch {
		if e.Error == nil {
			t.Errorf("Expected error, none received! \n")
		}
		if !os.IsNotExist(e.Error) {
			t.Errorf("Expected IsNotExist error, received: %s\n", e.Error)
		}
	}
}

var listRecursiveCases = []struct {
	file              string
	ignoreDirResults  bool
	ignoreFileResults bool
	ignoreCVSDirs     bool
	expected          []string
}{
	{"test_files/test_tree/.A", false, false, false, []string{
		"test_files/test_tree/.A/b",
		"test_files/test_tree/.A/b/C",
		"test_files/test_tree/.A/b/C/d",
		"test_files/test_tree/.A/b/C/d/E",
	},
	},
	{"test_files/test_tree/.A/b/C/d/E", false, false, false, []string{
		"test_files/test_tree/.A/b/C/d/E",
	},
	},
	{"test_files/test_tree", false, false, false, []string{
		"test_files/test_tree/.A",
		"test_files/test_tree/.A/b",
		"test_files/test_tree/.A/b/C",
		"test_files/test_tree/.A/b/C/d",
		"test_files/test_tree/.A/b/C/d/E",
		"test_files/test_tree/.a",
		"test_files/test_tree/.a/B",
		"test_files/test_tree/.a/B/c",
		"test_files/test_tree/.a/B/c/D",
		"test_files/test_tree/.a/B/c/D/e",
		"test_files/test_tree/.hg",
		"test_files/test_tree/.hg/E",
		"test_files/test_tree/.hg/e",
		"test_files/test_tree/.svn",
		"test_files/test_tree/.svn/E",
		"test_files/test_tree/.svn/e",
		"test_files/test_tree/A",
		"test_files/test_tree/A/b",
		"test_files/test_tree/A/b/C",
		"test_files/test_tree/A/b/C/d",
		"test_files/test_tree/A/b/C/d/E",
		"test_files/test_tree/a",
		"test_files/test_tree/a/B",
		"test_files/test_tree/a/B/c",
		"test_files/test_tree/a/B/c/D",
		"test_files/test_tree/a/B/c/D/e",
		"test_files/test_tree/slnA",
		"test_files/test_tree/slnA/b",
		"test_files/test_tree/slnA/b/C",
		"test_files/test_tree/slnA/b/C/d",
		"test_files/test_tree/slnA/b/C/d/E",
		"test_files/test_tree/slnB",
		"test_files/test_tree/slnB/b",
		"test_files/test_tree/slnB/b/C",
		"test_files/test_tree/slnB/b/C/d",
		"test_files/test_tree/slnB/b/C/d/E",
		"test_files/test_tree/slnC",
		"test_files/test_tree/slnD",
		"test_files/test_tree/slnE",
		"test_files/test_tree/slnF",
		"test_files/test_tree/slnG",
		"test_files/test_tree/z",
	},
	},
	{"test_files/test_tree", true, false, false, []string{
		"test_files/test_tree/.A/b/C/d/E",
		"test_files/test_tree/.a/B/c/D/e",
		"test_files/test_tree/.hg/E",
		"test_files/test_tree/.hg/e",
		"test_files/test_tree/.svn/E",
		"test_files/test_tree/.svn/e",
		"test_files/test_tree/A/b/C/d/E",
		"test_files/test_tree/a/B/c/D/e",
		"test_files/test_tree/slnA/b/C/d/E",
		"test_files/test_tree/slnB/b/C/d/E",
		"test_files/test_tree/slnC",
		"test_files/test_tree/slnD",
		"test_files/test_tree/slnE",
		"test_files/test_tree/slnF",
		"test_files/test_tree/slnG",
		"test_files/test_tree/z",
	},
	},
	{"test_files/test_tree", false, true, false, []string{
		"test_files/test_tree/.A",
		"test_files/test_tree/.A/b",
		"test_files/test_tree/.A/b/C",
		"test_files/test_tree/.A/b/C/d",
		"test_files/test_tree/.a",
		"test_files/test_tree/.a/B",
		"test_files/test_tree/.a/B/c",
		"test_files/test_tree/.a/B/c/D",
		"test_files/test_tree/.hg",
		"test_files/test_tree/.svn",
		"test_files/test_tree/A",
		"test_files/test_tree/A/b",
		"test_files/test_tree/A/b/C",
		"test_files/test_tree/A/b/C/d",
		"test_files/test_tree/a",
		"test_files/test_tree/a/B",
		"test_files/test_tree/a/B/c",
		"test_files/test_tree/a/B/c/D",
		"test_files/test_tree/slnA",
		"test_files/test_tree/slnA/b",
		"test_files/test_tree/slnA/b/C",
		"test_files/test_tree/slnA/b/C/d",
		"test_files/test_tree/slnB",
		"test_files/test_tree/slnB/b",
		"test_files/test_tree/slnB/b/C",
		"test_files/test_tree/slnB/b/C/d",
	},
	},
	{"test_files/test_tree", false, false, true, []string{
		"test_files/test_tree/.A",
		"test_files/test_tree/.A/b",
		"test_files/test_tree/.A/b/C",
		"test_files/test_tree/.A/b/C/d",
		"test_files/test_tree/.A/b/C/d/E",
		"test_files/test_tree/.a",
		"test_files/test_tree/.a/B",
		"test_files/test_tree/.a/B/c",
		"test_files/test_tree/.a/B/c/D",
		"test_files/test_tree/.a/B/c/D/e",
		"test_files/test_tree/A",
		"test_files/test_tree/A/b",
		"test_files/test_tree/A/b/C",
		"test_files/test_tree/A/b/C/d",
		"test_files/test_tree/A/b/C/d/E",
		"test_files/test_tree/a",
		"test_files/test_tree/a/B",
		"test_files/test_tree/a/B/c",
		"test_files/test_tree/a/B/c/D",
		"test_files/test_tree/a/B/c/D/e",
		"test_files/test_tree/slnA",
		"test_files/test_tree/slnA/b",
		"test_files/test_tree/slnA/b/C",
		"test_files/test_tree/slnA/b/C/d",
		"test_files/test_tree/slnA/b/C/d/E",
		"test_files/test_tree/slnB",
		"test_files/test_tree/slnB/b",
		"test_files/test_tree/slnB/b/C",
		"test_files/test_tree/slnB/b/C/d",
		"test_files/test_tree/slnB/b/C/d/E",
		"test_files/test_tree/slnC",
		"test_files/test_tree/slnD",
		"test_files/test_tree/slnE",
		"test_files/test_tree/slnF",
		"test_files/test_tree/slnG",
		"test_files/test_tree/z",
	},
	},
}

func TestListRecursive(t *testing.T) {
	log.SetOutput(io.Discard)
	// log.SetOutput(os.Stderr)
	goToRootDir()
	buf := setupLogging()
	for _, c := range listRecursiveCases {
		ch := ListRecursive(
			c.file,
			true,
			&BasicFileMatch{
				IgnoreDirResults:  c.ignoreDirResults,
				IgnoreFileResults: c.ignoreFileResults,
				IgnoreVCSDirs:     c.ignoreCVSDirs,
			},
			SortFnByName)
		tree := []string{}
		for e := range ch {
			if e.Error != nil {
				if !strings.Contains(e.Error.Error(), "too many links") {
					t.Errorf("Unexpected error: %s\n", e.Error)
					t.Log(buf.String())
					return
				}
			}
			tree = append(tree, e.Path)
		}
		compareTestStringSlices(t, c.expected, tree)
	}
	file := "tmp-no-permissions"
	err := os.Mkdir(file, 0000)
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
		t.Log(buf.String())
		return
	}
	ch := ListRecursive(file, false, &BasicFileMatch{}, SortFnByName)
	tree := []string{}
	for e := range ch {
		if e.Error == nil {
			t.Errorf("Expected error, none received! \n")
		}
		if !os.IsPermission(e.Error) {
			t.Errorf("Expected IsPermission error, received: %s\n", e.Error.Error())
		}
		tree = append(tree, e.Path)
	}
	err = os.RemoveAll(file)
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
		t.Log(buf.String())
		return
	}
	compareTestStringSlices(t, []string{file}, tree)
	file = "non-existent"
	ch = ListRecursive(file, false, &BasicFileMatch{}, SortFnByName)
	for e := range ch {
		if e.Error == nil {
			t.Errorf("Expected error, none received! \n")
		}
		if !os.IsNotExist(e.Error) {
			t.Errorf("Expected IsNotExist error, received: %s\n", e.Error)
		}
	}
	file = "test_files/test_files"
	ch = ListRecursive(
		file,
		false,
		&BasicFileMatch{
			MatchFileTypeList: []string{"ruby"},
		},
		SortFnByName)
	tree = []string{}
	for e := range ch {
		if e.Error != nil {
			t.Errorf("Unexpected error: %s\n", err)
			t.Log(buf.String())
			return
		}
		tree = append(tree, e.Path)
	}
	compareTestStringSlices(t, []string{"test_files/test_files/file.rb"}, tree)
}
