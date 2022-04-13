// This file is part of clitable.
//
// Copyright (C) 2017-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package clitable

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// TableInfo - Table information
type TableInfo struct {
	Columns            int
	Rows               int
	PerRowColumnWidths [][]int
	PerRowRows         [][]int // Number of Lines in a Row due to multiline entries.
	ColumnWidths       []int
	RowHeights         []int
}

func (i *TableInfo) String() string {
	str := fmt.Sprintf("%d Rows x %d Columns\nColumn widths: %v, Row heights: %v\nPerRowColumnWidths: %v\nPerRowRows: %v\n",
		i.Rows, i.Columns, i.ColumnWidths, i.RowHeights, i.PerRowColumnWidths, i.PerRowRows)
	return str
}

var single_width_table = []*unicode.RangeTable{
	unicode.Letter,
	unicode.Digit,
	unicode.Latin,
	{R16: []unicode.Range16{
		{0x0080, 0x00FF, 1}, // latin-1 supplement
	}},
}

func StringWidth(s string) (width int, doubleWidthCount int) {
	normalized := string(norm.NFC.Bytes([]byte(s)))
	l := len(normalized)                    // string len check, will count 2 or 3 for a single emoji
	u := utf8.RuneCountInString(normalized) // rune count, will correctly count number of chars
	diff := l - u                           // diff
	c := 0                                  // track multibyte chars that should have a 2 char width
	if diff > 0 {
		for _, r := range normalized {
			// fmt.Printf("l: %d, u: %d, %#U\n", l, u, r)
			if unicode.IsOneOf(single_width_table, r) {
				continue
			}
			rl := utf8.RuneLen(r)
			if rl > 1 { // multibyte char
				c++
			}
		}
	}
	return u + c, c
}

// GetTableInfo - Iterates over all the elements of the table to get number of Colums, Colum widths, etc.
func GetTableInfo(t Table) (*TableInfo, error) {
	var rows int
	var columns int
	var perRowColumnWidths [][]int
	var perRowRows [][]int
	var columnWidths []int
	var rowHeights []int
	for row := range t.RowIterator() {
		if row.Error != nil {
			return &TableInfo{}, row.Error
		}
		rowColumns := len(row.Fields)
		// Update columns
		if rowColumns > columns {
			columns = rowColumns
		}
		// Get Column Widths for this row
		rowColumnWidths := make([]int, rowColumns)
		rowRows := make([]int, rowColumns)
		for i, cData := range row.Fields {
			// Some records might be multiline, split the record and get the biggest width
			multiLine := strings.Split(cData, "\n")
			ll := 0
			for _, d := range multiLine {
				l, _ := StringWidth(d)
				if l > ll {
					ll = l
				}
			}
			rowColumnWidths[i] = ll
			rowRows[i] = len(multiLine)
		}
		perRowColumnWidths = append(perRowColumnWidths, rowColumnWidths)
		perRowRows = append(perRowRows, rowRows)
		rows++
	}
	// Get Overall Column Widths
	columnWidths = make([]int, columns)
	for c := 0; c < columns; c++ {
		for r := 0; r < rows; r++ {
			if len(perRowColumnWidths[r]) <= c {
				continue
			}
			w := perRowColumnWidths[r][c]
			if w > columnWidths[c] {
				columnWidths[c] = w
			}
		}
	}
	rowHeights = make([]int, rows)
	for r := 0; r < rows; r++ {
		for c := 0; c < columns; c++ {
			if len(perRowRows[r]) <= c {
				continue
			}
			h := perRowRows[r][c]
			if h > rowHeights[r] {
				rowHeights[r] = h
			}
		}
	}
	return &TableInfo{
		Rows:               rows,
		Columns:            columns,
		PerRowColumnWidths: perRowColumnWidths,
		PerRowRows:         perRowRows,
		ColumnWidths:       columnWidths,
		RowHeights:         rowHeights,
	}, nil
}
