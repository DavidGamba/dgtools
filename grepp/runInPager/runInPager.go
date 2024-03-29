// This file is part of grepp.
//
// Copyright (C) 2012-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package runInPager

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Runner interface {
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	Run(context.Context) error
}

/* func runInPager - runs a function passed as an argument and sends the output
* over a pager.
*
* The function could be of any kind as long as it takes an io.Writer as a
* parameter where to print the output.
* However, due to Go's lack of generic the function type in this script is set.
*
* If PAGER is 'less' it will use the -R option to force less to process ansi
* colors properly.
* Otherwise it uses whathever PAGER is set.
 */
func Command(ctx context.Context, caller Runner) error {
	pager := strings.Split(os.Getenv("PAGER"), " ")
	var cmd *exec.Cmd
	// Make sure to use -R to show colors when using less
	if pager[0] == "less" {
		pager[0] = "-R"
		cmd = exec.Command("less", pager...)
	} else {
		cmd = exec.Command(pager[0], pager[1:]...)
	}
	var pr *io.PipeReader
	var pw *io.PipeWriter
	// create a pipe (blocking)
	pr, pw = io.Pipe()
	defer pw.Close()
	cmd.Stdin = pr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cPager := make(chan struct{})
	// Create a blocking chan, Run the pager and unblock once it is finished
	go func() {
		_ = cmd.Run()
		close(cPager)
	}()

	caller.SetStdout(pw)
	caller.SetStderr(pw)
	err := caller.Run(ctx)

	// Close pipe
	pw.Close()

	// Wait for the pager to be finished
	<-cPager
	return err
}
