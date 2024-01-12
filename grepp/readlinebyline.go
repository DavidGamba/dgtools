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
