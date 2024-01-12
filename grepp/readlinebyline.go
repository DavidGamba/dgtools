// This file is part of grepp.
//
// Copyright (C) 2012-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	l "github.com/DavidGamba/dgtools/grepp/logging"
)

var errorBufferSizeTooSmall = fmt.Errorf("buffer size too small")

type LineError struct {
	Line       string
	LineNumber int
	Error      error
}

func ReadLineByLine(filename string, bufferSize int) <-chan LineError {
	l.Debug.Printf("[readLineByLine] %s : %d\n", filename, bufferSize)
	c := make(chan LineError)
	go func() {
		fh, err := os.Open(filename)
		if err != nil {
			c <- LineError{Error: err}
			close(c)
			return
		}
		defer fh.Close()
		reader := bufio.NewReaderSize(fh, bufferSize)
		// line number
		n := 0
		for {
			n++
			line, isPrefix, err := reader.ReadLine()
			if isPrefix {
				err := fmt.Errorf("%s: %w", filename, errorBufferSizeTooSmall)
				c <- LineError{Error: err}
				break
			}
			// stop reading file
			if err != nil {
				if err != io.EOF {
					c <- LineError{Error: err}
				}
				break
			}
			c <- LineError{Line: string(line), LineNumber: n}
		}
		close(c)
	}()
	return c
}
