// This file is part of run.
//
// Copyright (C) 2020-2024  David Gamba Rios
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
	Cmd      []string // exposed for mocking purposes only
	debug    bool
	env      []string
	dir      string
	Stdout   io.Writer // exposed for mocking purposes only
	Stderr   io.Writer // exposed for mocking purposes only
	stdin    io.Reader
	saveErr  bool
	printErr bool
	ctx      context.Context
	mockFn   MockFn
}

type runInfoContextKey string

func ContextWithRunInfo(ctx context.Context, value *RunInfo) context.Context {
	return context.WithValue(ctx, runInfoContextKey("runInfo"), value)
}

// CMD - Normal constructor.
func CMD(cmd ...string) *RunInfo {
	r := &RunInfo{Cmd: cmd}
	r.env = os.Environ()
	r.Stdout = nil
	r.Stderr = nil
	r.ctx = context.Background()
	r.printErr = true
	return r
}

// CMD - Pulls RunInfo from context if it exists and if not it initializes a new one.
// Useful when loading a RunInfo from context to ease testing.
func CMDCtx(ctx context.Context, cmd ...string) *RunInfo {
	v, ok := ctx.Value(runInfoContextKey("runInfo")).(*RunInfo)
	if ok {
		v.Cmd = cmd
		return v
	}
	r := CMD(cmd...)
	r.ctx = ctx
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

// GetEnv - used for testing
func (r *RunInfo) GetEnv() []string {
	return r.env
}

// Dir - specifies the working directory of the command.
func (r *RunInfo) Dir(dir string) *RunInfo {
	r.dir = dir
	return r
}

// GetDir - used for testing
func (r *RunInfo) GetDir() string {
	return r.dir
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
//	err := run.CMD("./command", "arg").SaveErr().Run() // or .STDOutOutput() or .CombinedOutput()
//	if err != nil {
//	  var exitErr *exec.ExitError
//	  if errors.As(err, &exitErr) {
//	    errOutput := exitErr.Stderr
func (r *RunInfo) SaveErr() *RunInfo {
	r.saveErr = true
	return r
}

// PrintErr - Regardless of operation mode, print errors to os.Stderr as they occur.
//
// Deprecated: This is the default behaviour, setting this is unnecessary and will be removed in the future.
func (r *RunInfo) PrintErr() *RunInfo {
	r.printErr = true
	return r
}

// DiscardErr - Don't print command error to stderr by default.
func (r *RunInfo) DiscardErr() *RunInfo {
	r.printErr = false
	return r
}

// CombinedOutput - Runs given CMD and returns STDOut and STDErr combined.
func (r *RunInfo) CombinedOutput() ([]byte, error) {
	var b bytes.Buffer
	r.Stdout = &b
	r.Stderr = &b
	err := r.Run()
	return b.Bytes(), err
}

// STDOutOutput - Runs given CMD and returns STDOut only.
//
// Stderr output is discarded unless a call to SaveErr() or PrintErr() was made.
func (r *RunInfo) STDOutOutput() ([]byte, error) {
	var b bytes.Buffer
	r.Stdout = &b
	err := r.Run()
	return b.Bytes(), err
}

type MockFn func(*RunInfo) error

func (r *RunInfo) Mock(fn MockFn) *RunInfo {
	r.mockFn = fn
	return r
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
//	Run()            // Output goes to os.Stdout and os.Stderr
//	Run(out)         // Sets the command's os.Stdout and os.Stderr to out.
//	Run(out, outErr) // Sets the command's os.Stdout to out and os.Stderr to outErr.
func (r *RunInfo) Run(w ...io.Writer) error {
	if r.debug {
		msg := fmt.Sprintf("run %v", r.Cmd)
		if r.dir != "" {
			msg += fmt.Sprintf(" on %s", r.dir)
		}
		Logger.Println(msg)
	}
	if len(w) == 0 {
		if r.Stdout == nil {
			r.Stdout = osStdout
		}
	} else if len(w) == 1 {
		r.Stdout = w[0]
		r.Stderr = w[0]
	} else if len(w) > 1 {
		r.Stdout = w[0]
		r.Stderr = w[1]
	}
	if r.printErr {
		if r.Stderr == nil {
			r.Stderr = osStderr
		} else if r.Stderr != osStderr {
			r.Stderr = io.MultiWriter(r.Stderr, osStderr)
		}
	}
	var b bytes.Buffer
	if r.saveErr {
		if r.Stderr == nil {
			r.Stderr = &b
		} else {
			r.Stderr = io.MultiWriter(r.Stderr, &b)
		}
	}

	if r.mockFn != nil {
		err := r.mockFn(r)
		if err != nil && r.saveErr {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitErr.Stderr = b.Bytes()
			}
		}
		return err
	}

	c := exec.CommandContext(r.ctx, r.Cmd[0], r.Cmd[1:]...)
	c.Dir = r.dir
	c.Env = r.env
	c.Stdout = r.Stdout
	c.Stderr = r.Stderr
	c.Stdin = r.stdin

	err := c.Run()
	if err != nil && r.saveErr {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitErr.Stderr = b.Bytes()
		}
	}
	return err
}
