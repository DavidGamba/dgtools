// This file is part of run.
//
// Copyright (C) 2020  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package run

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	out, err := CMD("echo", "hello", "world").CombinedOutput()
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
	}
	if string(out) != "hello world\n" {
		t.Errorf("wrong output: %s\n", out)
	}

	out, err = CMD("echo", "-n", "hello", "world").CombinedOutput()
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
	}
	if string(out) != "hello world" {
		t.Errorf("wrong output: %s\n", out)
	}

	out, err = CMD("ls", "./run").Dir("..").CombinedOutput()
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
	}
	if !strings.Contains(string(out), "run.go") {
		t.Errorf("wrong output: %s\n", out)
	}
}
