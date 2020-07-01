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

func (r *RunInfo) Env(env ...string) *RunInfo {
	r.env = append(r.env, env...)
	return r
}

func (r *RunInfo) Dir(dir string) *RunInfo {
	r.dir = dir
	return r
}

func (r *RunInfo) CombinedOutput() ([]byte, error) {
	var b bytes.Buffer
	r.stdout = &b
	r.stderr = &b
	err := r.Run()
	return b.Bytes(), err
}

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
