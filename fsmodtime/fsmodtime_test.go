// This file is part of fsmodtime.
//
// Copyright (C) 2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package fsmodtime

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
	"time"
)

func setupLogging() *bytes.Buffer {
	s := ""
	buf := bytes.NewBufferString(s)
	Logger.SetOutput(buf)
	return buf
}

type fswrap struct {
	fs.ReadDirFS
}

var (
	errorReadDir = fmt.Errorf("read dir error")
	errorInfo    = fmt.Errorf("info error")
)

func (fsw fswrap) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, errorReadDir
}

type deWrap struct{}

func TestFirstLast(t *testing.T) {
	m := make(fstest.MapFS)
	m["a/a.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(3, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["a/aa/aa.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(4, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["a/aa/ab.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(6, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["a/aa/ac.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(5, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["b/b.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(2, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["b/bb/bb.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["c"] = &fstest.MapFile{
		Mode: 0o777 | fs.ModeDir,
	}
	m["d/dd"] = &fstest.MapFile{
		Mode: 0o777 | fs.ModeDir,
	}

	t.Run("empty path", func(t *testing.T) {
		buf := setupLogging()
		_, _, err := Last(m, []string{"", "b", "c", "d"})
		if err == nil {
			t.Log(buf.String())
			t.Fatalf("expected error, nothing found")
		}
		if !errors.Is(err, ErrorInvalidPath) {
			t.Errorf("unexpected error: %v", err)
		}
		t.Log(buf.String())
	})

	t.Run("invalid path", func(t *testing.T) {
		buf := setupLogging()
		_, _, err := Last(m, []string{"x", "b", "c", "d"})
		if err == nil {
			t.Log(buf.String())
			t.Fatalf("expected error, nothing found")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("unexpected error: %v", err)
		}
		t.Log(buf.String())
	})

	t.Run("nil fs", func(t *testing.T) {
		buf := setupLogging()
		_, _, err := Last(nil, []string{"a", "b", "c", "d"})
		if err == nil {
			t.Log(buf.String())
			t.Fatalf("expected error, nothing found")
		}
		if !errors.Is(err, ErrorInvalidFS) {
			t.Errorf("unexpected error: %v", err)
		}
		_, _, err = First(nil, []string{"a", "b", "c", "d"})
		if err == nil {
			t.Log(buf.String())
			t.Fatalf("expected error, nothing found")
		}
		if !errors.Is(err, ErrorInvalidFS) {
			t.Errorf("unexpected error: %v", err)
		}
		t.Log(buf.String())
	})

	t.Run("walkPaths error", func(t *testing.T) {
		buf := setupLogging()
		err := walkPaths(m, []string{"a", "b", "c", "d"}, func(root string, fi fs.FileInfo) error {
			return fmt.Errorf("oops")
		})
		if err == nil {
			t.Log(buf.String())
			t.Fatalf("expected error, nothing found")
		}
		t.Log(buf.String())
	})

	t.Run("last", func(t *testing.T) {
		buf := setupLogging()
		path, fi, err := Last(m, []string{"a", "b", "c", "d"})
		if err != nil {
			t.Log(buf.String())
			t.Fatalf(err.Error())
		}
		if fi.Name() != "ab.txt" {
			t.Errorf("wrong result: %s, expected: %s", fi.Name(), "ab.txt")
		}
		if path != "a/aa" {
			t.Errorf("wrong result: %s, expected: %s", path, "a/aa")
		}
		t.Log(buf.String())
	})

	t.Run("first", func(t *testing.T) {
		buf := setupLogging()
		path, fi, err := First(m, []string{"a", "b", "c", "d"})
		if err != nil {
			t.Log(buf.String())
			t.Fatalf(err.Error())
		}
		if fi.Name() != "bb.txt" {
			t.Errorf("wrong result: %s, expected: %s", fi.Name(), "bb.txt")
		}
		if path != "b/bb" {
			t.Errorf("wrong result: %s, expected: %s", path, "b/bb")
		}
		t.Log(buf.String())
	})

	t.Run("fs error", func(t *testing.T) {
		buf := setupLogging()
		fsw := fswrap{m}
		_, _, err := Last(fsw, []string{"a", "b", "c", "d"})
		if err == nil {
			t.Log(buf.String())
			t.Fatalf("expected error, nothing found")
		}
		if !errors.Is(err, errorReadDir) {
			t.Errorf("unexpected error: %v", err)
		}
		t.Log(buf.String())
	})
}
