package main

import (
	"bytes"
	"io/fs"
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

func TestGetJumpDir(t *testing.T) {
	m := make(fstest.MapFS)
	m["a/dev/b.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(3, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["a/prod/b.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(4, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["a/staging/b.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(6, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["recurse/b.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(6, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["a/both/b.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(6, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["a/xoth/b.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(6, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["a/dev/both/b.txt"] = &fstest.MapFile{
		Mode: 0o777 | fs.ModeDir,
	}
	m["b/dev/c/b.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(3, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["b/prod/c/b.txt"] = &fstest.MapFile{
		Mode:    0o666,
		ModTime: time.Date(4, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	m["d/dd"] = &fstest.MapFile{
		Mode: 0o777 | fs.ModeDir,
	}

	tests := []struct {
		name     string
		cwd      string
		jump     string
		expected string
		err      bool
	}{
		{"rel", "a/dev", "prod", "a/prod", false},
		{"both", "a/dev", "both", "a/both", false},
		{"prefix", "a/dev", "rod", "a/prod", false},
		{"prefix", "a/dev", "y", "", true},
		{"suffix", "a/dev", "pro", "a/prod", false},
		{"suffix", "a/dev", "y", "", true},
		{"recurse", "a/dev", "recurse", "recurse", false},
		{"maintain", "b/dev/c", "prod", "b/prod", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := setupLogging()
			target, err := GetJumpDir(m, test.cwd, test.jump)
			if test.err && err == nil {
				t.Errorf("expected error, got none")
			}
			if !test.err && err != nil {
				t.Errorf("expected no error, got %s", err)
			}
			if target != test.expected {
				t.Errorf("expected %q, got %q", test.expected, target)
			}
			t.Log(buf.String())
		})
	}
	tests = []struct {
		name     string
		cwd      string
		jump     string
		expected string
		err      bool
	}{
		{"rel", "a/dev", "prod", "a/prod", false},
		{"both", "a/dev", "both", "a/both", false},
		{"prefix", "a/dev", "rod", "a/prod", false},
		{"prefix", "a/dev", "y", "", true},
		{"suffix", "a/dev", "pro", "a/prod", false},
		{"suffix", "a/dev", "y", "", true},
		{"recurse", "a/dev", "recurse", "recurse", false},
		{"maintain", "b/dev/c", "prod", "b/prod/c", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := setupLogging()
			target, err := GetRelDir(m, test.cwd, test.jump)
			if test.err && err == nil {
				t.Errorf("expected error, got none")
			}
			if !test.err && err != nil {
				t.Errorf("expected no error, got %s", err)
			}
			if target != test.expected {
				t.Errorf("expected %q, got %q", test.expected, target)
			}
			t.Log(buf.String())
		})
	}
}
