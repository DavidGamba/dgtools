// This file is part of ffind.
//
// Copyright (C) 2017-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffind

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func makeFS() fstest.MapFS {
	m := make(fstest.MapFS)

	m["test_tree"] = &fstest.MapFile{Mode: 0777 | fs.ModeDir}

	m["test_tree/.A"] = &fstest.MapFile{Mode: 0777 | fs.ModeDir}
	m["test_tree/.a"] = &fstest.MapFile{Mode: 0777 | fs.ModeDir}
	m["test_tree/.hg"] = &fstest.MapFile{Mode: 0777 | fs.ModeDir}
	m["test_tree/.svn"] = &fstest.MapFile{Mode: 0777 | fs.ModeDir}
	m["test_tree/.git"] = &fstest.MapFile{Mode: 0777 | fs.ModeDir}
	m["test_tree/A"] = &fstest.MapFile{Mode: 0777 | fs.ModeDir}
	m["test_tree/a"] = &fstest.MapFile{Mode: 0777 | fs.ModeDir}

	m["test_tree/slnA"] = &fstest.MapFile{Mode: 0777 | fs.ModeSymlink, Data: []byte("A")}
	m["test_tree/slnB"] = &fstest.MapFile{Mode: 0777 | fs.ModeSymlink, Data: []byte("slnA")}
	m["test_tree/slnC"] = &fstest.MapFile{Mode: 0777 | fs.ModeSymlink, Data: []byte("A/b/C/d/E")}
	m["test_tree/slnD"] = &fstest.MapFile{Mode: 0777 | fs.ModeSymlink, Data: []byte("a/B/c/D/e")}
	m["test_tree/slnE"] = &fstest.MapFile{Mode: 0777 | fs.ModeSymlink, Data: []byte("snlF")}
	m["test_tree/slnF"] = &fstest.MapFile{Mode: 0777 | fs.ModeSymlink, Data: []byte("snlG")}
	m["test_tree/slnG"] = &fstest.MapFile{Mode: 0777 | fs.ModeSymlink, Data: []byte("broken")}

	m["test_tree/.A/b/C/d/E"] = &fstest.MapFile{Mode: 0666}
	m["test_tree/.a/B/c/D/e"] = &fstest.MapFile{Mode: 0666}
	m["test_tree/.hg/E"] = &fstest.MapFile{Mode: 0666}
	m["test_tree/.hg/e"] = &fstest.MapFile{Mode: 0666}
	m["test_tree/.svn/E"] = &fstest.MapFile{Mode: 0666}
	m["test_tree/.svn/e"] = &fstest.MapFile{Mode: 0666}
	m["test_tree/A/b/C/d/E"] = &fstest.MapFile{Mode: 0666}
	m["test_tree/a/B/c/D/e"] = &fstest.MapFile{Mode: 0666}

	return m
}

// func TestListRecursiveFS(t *testing.T) {
// 	goToRootDir()
// 	// m := makeFS()
// 	m := os.DirFS("test_files")
// 	ee := NewEntryError(m, "test_tree")
// 	ch := listRecursiveFS(ee, 0, true, nil)
// 	for e := range ch {
// 		if e.Error != nil {
// 			t.Errorf("Error: %v\n", e.Error)
// 		}
// 		t.Errorf("%v\n", e.Path)
// 	}
// }

func goToRootDir() {
	cwd, _ := os.Getwd()
	if strings.HasSuffix(cwd, "lib/ffind") {
		// Go to the root of the repository
		_ = os.Chdir("../..")
	}
	cwd, _ = os.Getwd()
}

func compareTestStringSlices(t *testing.T, expected []string, received []string) {
	t.Helper()
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
	}
}

func TestNewFileError(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	// log.SetOutput(os.Stderr)
	goToRootDir()
	file := "./test_files/test_tree"
	fe, err := NewFileError(file)
	if err != nil {
		t.Fatalf("Unexpected error '%s': %s\n", file, err)
	}
	if fe.Path != "test_files/test_tree" {
		t.Fatalf("Paths don't match '%s': %s\n", file, "test_files/test_tree")
	}
	file = "non-existent"
	fe, err = NewFileError(file)
	if err == nil {
		t.Fatalf("Expected error, none received!\n")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("Expected IsNotExist error, received: %s\n", err)
	}
	log.Printf("%#v\n", fe)
}

// Make sure filepath.EvalSymlinks behaves as expected.
func TestEvalSymlink(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	goToRootDir()

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
			t.Fatalf("Unexpected error '%s': %s\n", c.file, err)
		}
		if read != c.expected {
			t.Fatalf("Got %s != Expected %s\n", read, c.expected)
		}
	}
	_, err := filepath.EvalSymlinks("./test_files/test_tree/slnE")
	if err == nil {
		t.Fatalf("Expected error, none received! \n")
	}
	if !strings.Contains(err.Error(), "too many links") {
		t.Fatalf("Unexpected error: %s\n", err)
	}
}
