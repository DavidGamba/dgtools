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
	"encoding/json"
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

// MapTable - Implements the table interface on a map
// Nested map entries will be json marshalled.
type MapTable struct {
	MapList      []map[string]any // list of maps to traverse
	keys         []string         // internal representation of keys
	Keys         []string         // User Key order override, doesn't not have to be exhaustive
	IgnoreHeader bool             // Don't include table header
	// NestedFormat will initially just be json but could also be YAML, though that requires an extra dependency
}

func (t MapTable) RowIterator() <-chan Row {
	c := make(chan Row)
	go func() {
		if len(t.MapList) <= 0 {
			close(c)
			return
		}
		t.keys = mapListKeys(t.MapList)
		t.keys = prefixSlice(t.Keys, t.keys)
		if !t.IgnoreHeader {
			c <- Row{Fields: t.keys}
		}
		for _, m := range t.MapList {
			row := make([]string, len(t.keys))
			for i, k := range t.keys {
				if v, ok := m[k]; ok {
					switch v := v.(type) {
					case string:
						row[i] = v
					default:
						b, err := json.Marshal(v)
						if err != nil {
							c <- Row{Error: err}
							close(c)
							return
						}
						row[i] = string(b)
					}
				}
			}
			c <- Row{Fields: row}
		}
		close(c)
	}()
	return c
}
