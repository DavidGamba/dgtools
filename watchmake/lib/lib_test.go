// This file is part of watchmake.
//
// Copyright (C) 2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package lib

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/DavidGamba/dgtools/private/hclutils"
	"github.com/davecgh/go-spew/spew"
)

func setupLogging() *bytes.Buffer {
	s := ""
	buf := bytes.NewBufferString(s)
	Logger.SetOutput(buf)
	return buf
}

func TestDecodeConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Watchmake
	}{
		{"simple",
			`
# version = 1

watch file {
	file = "x"
}

task build {
	files = watch.file.output
	task per_file {
		actions = [["echo", "modified", "{}"]]
	}
	task once {
		actions = [["go", "build"]]
	}
	task per_file {
		actions = [["echo", "completed", "{}"]]
	}
}
		`,
			&Watchmake{
				Watchers: []*Watcher{
					{ID: "file", File: "x", output: []string{"x"}},
				},
				Tasks: []*Task{
					{
						ID: "build", Files: []string{"x"}, Tasks: []*TaskExecution{
							{ID: "per_file", Actions: [][]string{{"echo", "modified", "{}"}}},
							{ID: "once", Actions: [][]string{{"go", "build"}}},
							{ID: "per_file", Actions: [][]string{{"echo", "completed", "{}"}}},
						},
					},
				},
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logging := setupLogging()
			buf := new(bytes.Buffer)
			parser, f, err := hclutils.ParseHCL(buf, []byte(test.input), "test.hcl")
			if err != nil {
				t.Fatalf("%s\n%s\n", err, buf.String())
			}
			config, err := DecodeConfig(buf, parser, f)
			if buf.String() != "" {
				t.Errorf("output didn't match expected value: %s", buf.String())
			}
			if err != nil {
				out := new(bytes.Buffer)
				spew.Fdump(out, config)
				t.Log(logging.String())
				t.Fatalf("Error: %s, %s", err, out.String())
			}
			if !reflect.DeepEqual(config, test.expected) {
				out := new(bytes.Buffer)
				exp := new(bytes.Buffer)
				spew.Fdump(out, config)
				spew.Fdump(exp, test.expected)
				t.Log(logging.String())
				t.Fatalf("unexpected value:\n%s!=\n%s", out.String(), exp.String())
			}
			t.Log(logging.String())
		})
	}
}
