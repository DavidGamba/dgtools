// This file is part of run.
//
// Copyright (C) 2020-2021  David Gamba Rios
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
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

var osStdout io.Writer = os.Stdout
var osStderr io.Writer = os.Stderr

type RunInfo struct {
	cmd      []string
	debug    bool
	env      []string
	dir      string
	stdout   io.Writer
	stderr   io.Writer
	stdin    io.Reader
	saveErr  bool
	printErr bool
	ctx      context.Context
}

func CMD(cmd ...string) *RunInfo {
	r := &RunInfo{cmd: cmd}
	r.env = os.Environ()
	r.stdout = nil
	r.stderr = nil
	r.ctx = context.Background()
	return r
}

func (r *RunInfo) Log() *RunInfo {
	r.debug = true
	return r
}

// Stdin - connect caller's os.Stdin to command stdin.
func (r *RunInfo) Stdin() *RunInfo {
	r.stdin = os.Stdin
	return r
}

// In - Pass input to stdin.
func (r *RunInfo) In(input []byte) *RunInfo {
	reader := bytes.NewReader(input)
	r.stdin = reader
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

// Ctx - specifies the context of the command to allow for timeouts.
func (r *RunInfo) Ctx(ctx context.Context) *RunInfo {
	r.ctx = ctx
	return r
}

// SaveErr - If the command starts but does not complete successfully, the error is of
// type *ExitError. In this case, save the error output into *ExitError.Stderr for retrieval.
//
// Retrieval can be done as shown below:
//
//   err := run.CMD("./command", "arg").SaveErr().Run() // or .STDOutOutput() or .CombinedOutput()
//   if err != nil {
//     var exitErr *exec.ExitError
//     if errors.As(err, &exitErr) {
//       errOutput := exitErr.Stderr
func (r *RunInfo) SaveErr() *RunInfo {
	r.saveErr = true
	return r
}

// PrintErr - Regardless of operation mode, print errors to os.Stderr as they occur.
func (r *RunInfo) PrintErr() *RunInfo {
	r.printErr = true
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
//
// Stderr output is discarded unless a call to SaveErr() or PrintErr() was made.
func (r *RunInfo) STDOutOutput() ([]byte, error) {
	var b bytes.Buffer
	r.stdout = &b
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
//
// Examples:
//
//   Run()            // Output goes to os.Stdout and os.Stderr
//   Run(out)         // Sets the command's os.Stdout and os.Stderr to out.
//   Run(out, outErr) // Sets the command's os.Stdout to out and os.Stderr to outErr.
func (r *RunInfo) Run(w ...io.Writer) error {
	if r.debug {
		msg := fmt.Sprintf("run %v", r.cmd)
		if r.dir != "" {
			msg += fmt.Sprintf(" on %s", r.dir)
		}
		Logger.Println(msg)
	}
	c := exec.CommandContext(r.ctx, r.cmd[0], r.cmd[1:]...)
	c.Dir = r.dir
	c.Env = r.env
	if len(w) == 0 {
		if r.stdout == nil {
			r.stdout = osStdout
		}
		c.Stdout = r.stdout
		c.Stderr = r.stderr
	} else if len(w) == 1 {
		c.Stdout = w[0]
		c.Stderr = w[0]
	} else if len(w) > 1 {
		c.Stdout = w[0]
		c.Stderr = w[1]
	}
	if r.printErr {
		if c.Stderr == nil {
			c.Stderr = osStderr
		} else if c.Stderr != osStderr {
			c.Stderr = io.MultiWriter(c.Stderr, osStderr)
		}
	}
	var b bytes.Buffer
	if r.saveErr {
		if c.Stderr == nil {
			c.Stderr = &b
		} else {
			c.Stderr = io.MultiWriter(c.Stderr, &b)
		}
	}
	c.Stdin = r.stdin
	err := c.Run()
	if err != nil && r.saveErr {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitErr.Stderr = b.Bytes()
		}
	}
	return err
}
