// This file is part of ffind.
//
// Copyright (C) 2017-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffind

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func TestListOneLevelDarwin(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	// log.SetOutput(os.Stderr)
	goToRootDir()
	cases := []struct {
		file     string
		follow   bool
		expected []string
	}{
		{"./test_files/test_tree", false, []string{
			"test_files/test_tree/.A",
			"test_files/test_tree/.hg",
			"test_files/test_tree/.svn",
			"test_files/test_tree/A",
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
		{"./test_files/test_tree/.A", false, []string{
			"test_files/test_tree/.A/b",
		},
		},
		{"./test_files/test_tree/.A/b/C/d/E", false, []string{
			"test_files/test_tree/.A/b/C/d/E",
		},
		},
		{"./test_files/test_tree/slnA", false, []string{
			"test_files/test_tree/slnA",
		},
		},
		{"./test_files/test_tree/slnA", true, []string{
			"test_files/test_tree/slnA/b",
		},
		},
		{"./test_files/test_tree/slnB", false, []string{
			"test_files/test_tree/slnB",
		},
		},
		{"./test_files/test_tree/slnB", true, []string{
			"test_files/test_tree/slnB/b",
		},
		},
	}
	for _, c := range cases {
		ch := ListOneLevel(c.file, c.follow, SortFnByName)
		tree := []string{}
		for e := range ch {
			if e.Error != nil {
				t.Fatalf("Unexpected error: %s\n", e.Error)
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
			t.Fatalf("Expected error, none received! \n")
		}
	}
	file = "tmp-no-permissions"
	err := os.Mkdir(file, 0000)
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
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
		t.Fatalf("Unexpected error: %s\n", err)
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

var listRecursiveCasesDarwin = []struct {
	file              string
	ignoreDirResults  bool
	ignoreFileResults bool
	ignoreCVSDirs     bool
	expected          []string
}{
	{"./test_files/test_tree/.A", false, false, false, []string{
		"test_files/test_tree/.A/b",
		"test_files/test_tree/.A/b/C",
		"test_files/test_tree/.A/b/C/d",
		"test_files/test_tree/.A/b/C/d/E",
	},
	},
	{"./test_files/test_tree/.A/b/C/d/E", false, false, false, []string{
		"test_files/test_tree/.A/b/C/d/E",
	},
	},
	{"./test_files/test_tree", false, false, false, []string{
		"test_files/test_tree/.A",
		"test_files/test_tree/.A/b",
		"test_files/test_tree/.A/b/C",
		"test_files/test_tree/.A/b/C/d",
		"test_files/test_tree/.A/b/C/d/E",
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
	{"./test_files/test_tree", true, false, false, []string{
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
	{"./test_files/test_tree", false, true, false, []string{
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
	{"./test_files/test_tree", false, false, true, []string{
		"test_files/test_tree/.A",
		"test_files/test_tree/.A/b",
		"test_files/test_tree/.A/b/C",
		"test_files/test_tree/.A/b/C/d",
		"test_files/test_tree/.A/b/C/d/E",
		"test_files/test_tree/A",
		"test_files/test_tree/A/b",
		"test_files/test_tree/A/b/C",
		"test_files/test_tree/A/b/C/d",
		"test_files/test_tree/A/b/C/d/E",
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

func TestListRecursiveDarwin(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	// log.SetOutput(os.Stderr)
	goToRootDir()
	for _, c := range listRecursiveCasesDarwin {
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
					t.Fatalf("Unexpected error: %s\n", e.Error)
				}
			}
			tree = append(tree, e.Path)
		}
		compareTestStringSlices(t, c.expected, tree)
	}
	file := "tmp-no-permissions"
	err := os.Mkdir(file, 0000)
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
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
		t.Fatalf("Unexpected error: %s\n", err)
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
			t.Fatalf("Unexpected error: %s\n", err)
		}
		tree = append(tree, e.Path)
	}
	compareTestStringSlices(t, []string{"test_files/test_files/file.rb"}, tree)
}
