// This file is part of run.
//
// Copyright (C) 2020  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package run

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	out, err := CMD("echo", "hello", "world").CombinedOutput()
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
	}
	if string(out) != "hello world\n" {
		t.Errorf("wrong output: %s\n", out)
	}

	out, err = CMD("echo", "hello", "world").STDOutOutput()
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
	}
	if string(out) != "hello world\n" {
		t.Errorf("wrong output: %s\n", out)
	}

	out, err = CMD("ls", "x").STDOutOutput()
	if err == nil {
		t.Errorf("Unexpected pass: %s\n", err)
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if string(exitErr.Stderr) != "" {
			t.Errorf("wrong output: %s\n", out)
		}
		if exitErr.ExitCode() != 2 {
			t.Errorf("wrong exit code: %d\n", exitErr.ExitCode())
		}
	}
	if string(out) != "" {
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

	out, err = CMD("cat").In([]byte("hello world")).CombinedOutput()
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
	}
	if !strings.Contains(string(out), "hello world") {
		t.Errorf("wrong output: %s\n", out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = CMD("sleep", "1").Ctx(ctx).CombinedOutput()
	if err == nil {
		t.Errorf("Unexpected pass: %s\n", err)
	}
}
