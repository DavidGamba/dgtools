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
	"testing"
)

func TestListOneLevelLinux(t *testing.T) {
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
