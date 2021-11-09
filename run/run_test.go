// This file is part of run.
//
// Copyright (C) 2020-2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build linux || darwin
// +build linux darwin

package run

import (
	"bytes"
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

	t.Run("STDOutOutput", func(t *testing.T) {
		out, err := CMD("echo", "hello", "world").STDOutOutput()
		if err != nil {
			t.Errorf("Unexpected error: %s\n", err)
		}
		if string(out) != "hello world\n" {
			t.Errorf("wrong output: %s\n", out)
		}
	})

	t.Run("STDOutOutput without stderr", func(t *testing.T) {
		out, err := CMD("ls", "x").STDOutOutput()
		if err == nil {
			t.Errorf("Unexpected pass: %s\n", err)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			errOut := string(exitErr.Stderr)
			if errOut != "" {
				t.Errorf("wrong stderr output: %s\n", errOut)
			}
			if exitErr.ExitCode() == 0 {
				t.Errorf("wrong exit code: %d\n", exitErr.ExitCode())
			}
		}
		if string(out) != "" {
			t.Errorf("wrong output: %s\n", out)
		}
	})

	t.Run("STDOutOutput with stderr", func(t *testing.T) {
		out, err := CMD("ls", "x").SaveErr().STDOutOutput()
		if err == nil {
			t.Errorf("Unexpected pass: %s\n", err)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			errOut := string(exitErr.Stderr)
			if !strings.Contains(errOut, "No such file") {
				t.Errorf("wrong stderr output: %s\n", errOut)
			}
			if exitErr.ExitCode() == 0 {
				t.Errorf("wrong exit code: %d\n", exitErr.ExitCode())
			}
		}
		if string(out) != "" {
			t.Errorf("wrong output: %s\n", out)
		}
	})

	t.Run("STDOutOutput print stderr", func(t *testing.T) {
		var b bytes.Buffer
		osStderr = &b
		out, err := CMD("ls", "x").PrintErr().STDOutOutput()
		if err == nil {
			t.Errorf("Unexpected pass: %s\n", err)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			errOut := string(exitErr.Stderr)
			if errOut != "" {
				t.Errorf("wrong stderr output: %s\n", errOut)
			}
			if exitErr.ExitCode() == 0 {
				t.Errorf("wrong exit code: %d\n", exitErr.ExitCode())
			}
		}
		if string(out) != "" {
			t.Errorf("wrong output: %s\n", out)
		}
		if !strings.Contains(b.String(), "No such file") {
			t.Errorf("wrong osStderr output: %s\n", b.String())
		}
	})

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
