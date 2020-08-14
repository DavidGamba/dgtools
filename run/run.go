// This file is part of run.
//
// Copyright (C) 2020  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package run provides a wrapper around os/exec with method chaining for modifying behaviour.
*/
package run

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

type RunInfo struct {
	cmd    []string
	debug  bool
	env    []string
	dir    string
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader
}

func CMD(cmd ...string) *RunInfo {
	r := &RunInfo{cmd: cmd}
	r.env = os.Environ()
	r.stdout = os.Stdout
	r.stderr = os.Stderr
	return r
}

func (r *RunInfo) Log() *RunInfo {
	r.debug = true
	return r
}

func (r *RunInfo) Stdin() *RunInfo {
	r.stdin = os.Stdin
	return r
}

// Env - Add key=value pairs to the environment of the process.
func (r *RunInfo) Env(env ...string) *RunInfo {
	r.env = append(r.env, env...)
	return r
}

// Dir - specifies the working directory of the command.
func (r *RunInfo) Dir(dir string) *RunInfo {
	r.dir = dir
	return r
}

// CombinedOutput - Runs given CMD and returns STDOut and STDErr combined.
func (r *RunInfo) CombinedOutput() ([]byte, error) {
	var b bytes.Buffer
	r.stdout = &b
	r.stderr = &b
	err := r.Run()
	return b.Bytes(), err
}

// STDOutOutput - Runs given CMD and returns STDOut only.
func (r *RunInfo) STDOutOutput() ([]byte, error) {
	var b bytes.Buffer
	r.stdout = &b
	r.stderr = nil
	err := r.Run()
	return b.Bytes(), err
}

// Run - wrapper around os/exec CMD.Run()
//
// Run starts the specified command and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems copying
// stdin, stdout, and stderr, and exits with a zero exit status.
//
// If the command starts but does not complete successfully, the error is of
// type *ExitError. Other error types may be returned for other situations.
func (r *RunInfo) Run() error {
	if r.debug {
		msg := fmt.Sprintf("run %v", r.cmd)
		if r.dir != "" {
			msg += fmt.Sprintf(" on %s", r.dir)
		}
		Logger.Println(msg)
	}
	c := exec.Command(r.cmd[0], r.cmd[1:]...)
	c.Dir = r.dir
	c.Env = r.env
	c.Stdout = r.stdout
	c.Stderr = r.stderr
	c.Stdin = r.stdin
	return c.Run()
}
