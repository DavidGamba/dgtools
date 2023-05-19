// This file is part of clitable.
//
// Copyright (C) 2017-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package clitable

import (
	"encoding/csv"
	"io"
)

// Table - interface used to walk through a table one row at a time
type Table interface {
	RowIterator() <-chan Row
}

// Row - Data and Error struct
type Row struct {
	Fields []string
	Error  error
}

// SimpleTable - A basic structure that implements the Table interface.
type SimpleTable struct {
	Data [][]string
}

// RowIterator - Implements the Table interface.
func (t SimpleTable) RowIterator() <-chan Row {
	return SimpleRowIterator(t.Data)
}

// SimpleRowIterator -
func SimpleRowIterator(data [][]string) <-chan Row {
	c := make(chan Row)
	go func() {
		for _, row := range data {
			c <- Row{Fields: row}
		}
		close(c)
	}()
	return c
}

// CSVTable - Implements the table interface from an io.Reader to CSV data.
type CSVTable struct {
	Reader    io.Reader
	Separator rune
}

// RowIterator - Implements the Table interface.
func (t CSVTable) RowIterator() <-chan Row {
	return CSVRowIterator(t.Reader, t.Separator)
}

// CSVRowIterator -
func CSVRowIterator(reader io.Reader, separator rune) <-chan Row {
	c := make(chan Row)
	go func() {
		r := csv.NewReader(reader)
		if separator != 0 {
			r.Comma = separator
		}
		r.FieldsPerRecord = -1
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				c <- Row{Error: err}
				close(c)
				return
			}
			c <- Row{Fields: record}
		}
		close(c)
	}()
	return c
}
