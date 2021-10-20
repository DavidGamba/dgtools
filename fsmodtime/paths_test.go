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
		{"env", []string{"hello $_world"}, map[string]string{"_world": "mundo"}, []string{"hello mundo"}, nil},
		{"env", []string{"hello ${_world}"}, map[string]string{"_world": "mundo"}, []string{"hello mundo"}, nil},
		{"env", []string{"hello ${_world}", "/$_world$_home"},
			map[string]string{"_world": "mundo", "_home": "/home/david"},
			[]string{"hello mundo", "/mundo/home/david"}, nil},
		{"error", []string{"hello $_world"}, nil, []string{"hello "}, ErrNotFound},
		{"error", []string{"hello $_world$_home"}, nil, []string{"hello "}, ErrNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setup(test.env)
			out, err := ExpandEnv(test.input)
			checkError(t, err, test.err)
			if !reflect.DeepEqual(out, test.expected) {
				t.Errorf("got: %#v, expected: %#v", out, test.expected)
			}
			cleanup(test.env)
		})
	}
}

func TestTarget(t *testing.T) {
	// Given two input dirs, src and images, we want to validate that our outputs are newer than any of the inputs.
	// We only care about *.adoc and *.jpg files, so the metadata.yml files should be ignored.
	m := make(fstest.MapFS)
	m["src/a.adoc"] = &fstest.MapFile{
		Mode:    0666,
		ModTime: time.Date(3, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["src/b.adoc"] = &fstest.MapFile{
		Mode:    0666,
		ModTime: time.Date(4, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["src/metadata.yml"] = &fstest.MapFile{
		Mode:    0666,
		ModTime: time.Date(99, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["images/a.jpg"] = &fstest.MapFile{
		Mode:    0666,
		ModTime: time.Date(5, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["images/b.jpg"] = &fstest.MapFile{
		Mode:    0666,
		ModTime: time.Date(6, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["images/metadata.yml"] = &fstest.MapFile{
		Mode:    0666,
		ModTime: time.Date(99, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	targets := []string{"$_outputs_dir/doc.pdf", "$_outputs_dir/*.html"}
	sources := []string{"$_src_dir/*.adoc", "$_images_dir/*.jpg", "$_images_dir/*.png"}
	os.Setenv("_outputs_dir", "outputs")
	os.Setenv("_src_dir", "src")
	os.Setenv("_images_dir", "images")

	t.Run("missing targets", func(t *testing.T) {
		buf := setupLogging()
		paths, modified, err := Target(m, targets, sources)
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
			Mode:    0666,
			ModTime: time.Date(2, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/index.html"] = &fstest.MapFile{
			Mode:    0666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/a.html"] = &fstest.MapFile{
			Mode:    0666,
			ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/b.html"] = &fstest.MapFile{
			Mode:    0666,
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

	t.Run("current targets", func(t *testing.T) {
		buf := setupLogging()
		m["outputs/doc.pdf"] = &fstest.MapFile{
			Mode:    0666,
			ModTime: time.Date(51, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/index.html"] = &fstest.MapFile{
			Mode:    0666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/a.html"] = &fstest.MapFile{
			Mode:    0666,
			ModTime: time.Date(50, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		m["outputs/b.html"] = &fstest.MapFile{
			Mode:    0666,
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
}
