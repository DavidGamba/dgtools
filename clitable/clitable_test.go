// This file is part of clitable.
//
// Copyright (C) 2017-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package clitable_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidGamba/dgtools/clitable"
)

type Data struct {
	Info []struct {
		Name string
		ID   int
	}
}

func (d *Data) RowIterator() <-chan clitable.Row {
	c := make(chan clitable.Row)
	go func() {
		c <- clitable.Row{Fields: []string{"", "Name", "ID"}}
		for i, row := range d.Info {
			c <- clitable.Row{Fields: []string{strconv.Itoa(i + 1), row.Name, strconv.Itoa(row.ID)}}
		}
		close(c)
	}()
	return c
}

func TestTableStructData(t *testing.T) {
	data := &Data{[]struct {
		Name string
		ID   int
	}{{"Hello", 1}, {"World ⚽⛪⚽⛪Å®", 2}}}

	clitable.NewTablePrinter().Fprint(os.Stdout, data)
	simpleData := [][]string{{"Hello", "1"}, {"World", "2"}}
	clitable.NewTablePrinter().HasHeader(false).Print(clitable.SimpleTable{simpleData})
	r := strings.NewReader("Hello,1\nWorld,2\n")
	clitable.NewTablePrinter().HasHeader(false).FprintCSVReader(os.Stdout, r)

	clitable.NewTablePrinter().SetStyle(clitable.Full).Print(data)
	clitable.NewTablePrinter().SetStyle(clitable.Full).HasHeader(false).Print(data)
	clitable.NewTablePrinter().SetStyle(clitable.Compact).Print(data)
	clitable.NewTablePrinter().SetStyle(clitable.Ascii).Print(data)
	clitable.NewTablePrinter().SetStyle(clitable.Space).Print(data)
	clitable.NewTablePrinter().SetStyle(clitable.CSV).Print(data)
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		lenght int
		count  int
	}{
		{"A", "A", 1, 0},
		{"Å", "Å", 1, 0},
		{"⚽", "⚽", 2, 1},
		{"⛪", "⛪", 2, 1},
		{"®", "®", 1, 0},
		{"World Å®⚽⛪⚽⛪", "World Å®⚽⛪⚽⛪", 16, 4},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l, c := clitable.StringWidth(test.input)
			if test.lenght != l || test.count != c {
				t.Errorf("expected: %d, got: %d, %s - %#U\n", test.lenght, l, test.input, []byte(test.input))
			}
		})
	}
}
