// This file is part of fsmodtime.
//
// Copyright (C) 2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package fsmodtime

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"
	"time"
)

func checkError(t *testing.T, got, expected error) {
	t.Helper()
	if (got == nil && expected != nil) || (got != nil && expected == nil) || (got != nil && expected != nil && !errors.Is(got, expected)) {
		t.Errorf("wrong error received: got = '%#v', want '%#v'", got, expected)
	}
}

func TestExpandEnv(t *testing.T) {
	setup := func(env map[string]string) {
		for k, v := range env {
			os.Setenv(k, v)
		}
	}
	cleanup := func(env map[string]string) {
		for k := range env {
			os.Unsetenv(k)
		}
	}
	tests := []struct {
		name     string
		input    []string
		env      map[string]string
		expected []string
		err      error
	}{
		{"empty", []string{}, nil, []string{}, nil},
		{"no env", []string{"hello world"}, nil, []string{"hello world"}, nil},
		{"env", []string{"home $HOME"}, map[string]string{"HOME": "test"}, []string{"home test"}, nil},
		{"env", []string{"home $HOME/hello"}, map[string]string{"HOME": "test"}, []string{"home test/hello"}, nil},
		{"env", []string{"home ~/hello"}, map[string]string{"HOME": "test"}, []string{"home test/hello"}, nil},
		{"env", []string{"~/hello"}, map[string]string{"HOME": "test"}, []string{"test/hello"}, nil},
		{"env", []string{"hello $_world"}, map[string]string{"_world": "mundo"}, []string{"hello mundo"}, nil},
		{"env", []string{"hello ${_world}"}, map[string]string{"_world": "mundo"}, []string{"hello mundo"}, nil},
		{
			"env",
			[]string{"hello ${_world}", "/$_world$_home"},
			map[string]string{"_world": "mundo", "_home": "/home/david"},
			[]string{"hello mundo", "/mundo/home/david"},
			nil,
		},
		{"error", []string{"hello $_world"}, nil, []string{"hello "}, ErrNotFound},
		{"error", []string{"hello $_world$_home"}, nil, []string{"hello "}, ErrNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setup(test.env)
			out, err := ExpandEnv(test.input, nil)
			checkError(t, err, test.err)
			if !reflect.DeepEqual(out, test.expected) {
				t.Errorf("got: %#v, expected: %#v", out, test.expected)
			}
			cleanup(test.env)
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := ExpandEnv(test.input, test.env)
			checkError(t, err, test.err)
			if !reflect.DeepEqual(out, test.expected) {
				t.Errorf("got: %#v, expected: %#v", out, test.expected)
			}
		})
	}
}

func TestTarget(t *testing.T) {
	// Given two input dirs, src and images, we want to validate that our outputs are newer than any of the inputs.
	// We only care about *.adoc and *.jpg files, so the metadata.yaml files should be ignored.
	m := make(fstest.MapFS)
	m["src/a.adoc"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(3, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["src/b.adoc"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(4, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["src/c/d.adoc"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(5, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["src/metadata.yaml"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(99, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["images/a.jpg"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(6, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["images/b.jpg"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(7, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["images/c/d.jpg"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(8, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["images/metadata.yaml"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(99, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	targets := []string{"$_outputs_dir/doc.pdf", "$_outputs_dir/*.html"}
	sources := []string{"$_src_dir/*.adoc", "$_images_dir/*.jpg", "$_images_dir/*.png", "$_templates_dir/*"}
	os.Setenv("_outputs_dir", "outputs")
	os.Setenv("_src_dir", "src")
	os.Setenv("_images_dir", "images")
	os.Setenv("_templates_dir", "templates")

	t.Run("missing targets", func(t *testing.T) {
		buf := setupLogging()
		paths, modified, err := Target(m, targets, sources, Recursive(true))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(paths, []string{}) {
			t.Errorf("unexpected paths: %v", paths)
		}
		if !modified {
			t.Errorf("unexpected modified: %v", modified)
		}
		t.Log(buf.String())
	})

	t.Run("old targets", func(t *testing.T) {
		buf := setupLogging()
		m["outputs/doc.pdf"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(2, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/index.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/b.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		paths, modified, err := Target(m, targets, sources)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(paths, []string{"images/b.jpg"}) {
			t.Errorf("unexpected paths: %v", paths)
		}
		if !modified {
			t.Errorf("unexpected modified: %v", modified)
		}
		t.Log(buf.String())
	})

	t.Run("old targets recursive", func(t *testing.T) {
		buf := setupLogging()
		m["outputs/doc.pdf"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(2, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/index.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/b.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(9, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/b.yaml"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(10, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/c/d.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(11, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/e/f.yaml"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(12, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		paths, modified, err := Target(m, targets, sources, Recursive(true))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(paths, []string{"templates/e/f.yaml"}) {
			t.Errorf("unexpected paths: %v", paths)
		}
		if !modified {
			t.Errorf("unexpected modified: %v", modified)
		}
		t.Log(buf.String())
	})

	t.Run("old targets not recursive", func(t *testing.T) {
		buf := setupLogging()
		m["outputs/doc.pdf"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(2, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/index.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/b.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(9, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/b.yaml"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(10, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/c/d.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(11, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/e/f.yaml"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(12, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		paths, modified, err := Target(m, targets, sources, Recursive(false))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(paths, []string{"templates/b.yaml"}) {
			t.Errorf("unexpected paths: %v", paths)
		}
		if !modified {
			t.Errorf("unexpected modified: %v", modified)
		}
		t.Log(buf.String())
	})

	t.Run("current targets", func(t *testing.T) {
		buf := setupLogging()
		m["outputs/doc.pdf"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(51, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/index.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/b.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		paths, modified, err := Target(m, targets, sources)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(paths, []string{}) {
			t.Errorf("unexpected paths: %v", paths)
		}
		if modified {
			t.Errorf("unexpected modified: %v", modified)
		}
		t.Log(buf.String())
	})

	t.Run("current targets recursive", func(t *testing.T) {
		buf := setupLogging()
		m["outputs/doc.pdf"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(51, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/index.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/b.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(9, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/b.yaml"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(10, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/c/d.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(11, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/e/f.yaml"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(12, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		paths, modified, err := Target(m, targets, sources, Recursive(true))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(paths, []string{}) {
			t.Errorf("unexpected paths: %v", paths)
		}
		if modified {
			t.Errorf("unexpected modified: %v", modified)
		}
		t.Log(buf.String())
	})

	t.Run("current targets not recursive", func(t *testing.T) {
		buf := setupLogging()
		m["outputs/doc.pdf"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(51, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/index.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/b.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/a.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(9, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/b.yaml"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(10, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/c/d.html"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(51, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["templates/e/f.yaml"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(52, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		paths, modified, err := Target(m, targets, sources, Recursive(false))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(paths, []string{}) {
			t.Errorf("unexpected paths: %v", paths)
		}
		if modified {
			t.Errorf("unexpected modified: %v", modified)
		}
		t.Log(buf.String())
	})

	t.Run("bad sources", func(t *testing.T) {
		buf := setupLogging()
		m["target"] = &fstest.MapFile{
			Mode:    0o666,
			ModTime: time.Date(51, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		paths, modified, err := Target(m, []string{"target"}, []string{"source"})
		if err == nil {
			t.Log(buf.String())
			t.Fatalf("expected error, nothing found")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("unexpected error: %v", err)
		}
		if paths != nil {
			t.Errorf("unexpected paths: %v", paths)
		}
		if modified {
			t.Errorf("unexpected modified: %v", modified)
		}
		t.Log(buf.String())
	})
}

func TestParentDir(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		buf := setupLogging()
		dir, err := ParentDir([]string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if dir != "" {
			t.Errorf("unexpected dir: %s", dir)
		}
		t.Log(buf.String())
	})
	t.Run("one", func(t *testing.T) {
		buf := setupLogging()
		dir, err := ParentDir([]string{"/a/b/c/d"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if dir != "/a/b/c/d" {
			t.Errorf("unexpected dir: %s", dir)
		}
		t.Log(buf.String())
	})
	t.Run("simple", func(t *testing.T) {
		buf := setupLogging()
		dir, err := ParentDir([]string{"/a/b/c/d", "/a/b/c/e"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if dir != "/a/b/c" {
			t.Errorf("unexpected dir: %s", dir)
		}
		t.Log(buf.String())
	})
	t.Run("relative", func(t *testing.T) {
		buf := setupLogging()
		dir, err := ParentDir([]string{"../a/b/c/d", "../a/b/c/e"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		cwd, _ := os.Getwd()
		dir, _ = filepath.Rel(cwd, dir)
		if dir != "../a/b/c" {
			t.Errorf("unexpected dir: %s", dir)
		}
		t.Log(buf.String())
	})
	t.Run("mixed", func(t *testing.T) {
		buf := setupLogging()
		dir, err := ParentDir([]string{"a/b/c/d", "../../e"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		cwd, _ := os.Getwd()
		dir, _ = filepath.Rel(cwd, dir)
		if dir != "../.." {
			t.Errorf("unexpected dir: %s", dir)
		}
		t.Log(buf.String())
	})
}
