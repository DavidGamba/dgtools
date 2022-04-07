// This file is part of clitable.
//
// Copyright (C) 2017-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package clitable provides a tool to view data as a table on the cmdline.

		┌──┬──┐
		│  │  │
		├──┼──┤
		└──┴──┘

*/
package clitable

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

// Logger - Default *log.Logger variable.
// Set output to os.Stderr or override.
var Logger = log.New(ioutil.Discard, "clitable DEBUG ", log.LstdFlags)

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

// Row - Data and Error struct
type Row struct {
	Fields []string
	Error  error
}

// Table - interface used to print a table
type Table interface {
	RowIterator() <-chan Row
}

// TableConfig -
type TableConfig struct {
	HeaderStart     string
	HeaderEnd       string
	HeaderJuncture  string
	Body            string
	Column          string
	LineStart       string
	LineEnd         string
	LineJuncture    string
	LineBetweenRows bool
	ColumnEdges     bool
}

// NewDefaultTableConfig -
func NewDefaultTableConfig(columnEdges, lineBetweenRows bool) TableConfig {
	return TableConfig{
		HeaderStart:     "┌",
		HeaderEnd:       "┐",
		HeaderJuncture:  "┬",
		Body:            "─",
		Column:          "│",
		LineStart:       "├",
		LineEnd:         "┤",
		LineJuncture:    "┼",
		LineBetweenRows: lineBetweenRows,
		ColumnEdges:     columnEdges,
	}
}

func printHorizontalLine(start, juncture, end, body string, tableInfo *TableInfo, config TableConfig) {
	for i := 0; i < tableInfo.Columns; i++ {
		if i == 0 {
			if config.ColumnEdges {
				fmt.Printf(start)
			} else {
				fmt.Printf("")
			}
		} else {
			fmt.Printf(juncture)
		}
		// Column width with space padding on each side
		fmt.Printf(strings.Repeat(body, tableInfo.ColumnWidths[i]+2))
		if i+1 == tableInfo.Columns {
			if config.ColumnEdges {
				fmt.Println(end)
			} else {
				fmt.Println("")
			}
		}
	}
}

// PrintCSVTable - Given an io.Reader that points to CSV content, it prints the CSV as a table.
func PrintCSVTable(r io.Reader) error {
	var readerCopy bytes.Buffer
	reader := io.TeeReader(r, &readerCopy)
	t := CSVTable{reader}
	tableInfo, err := GetTableInfo(t)
	if err != nil {
		return err
	}
	Logger.Printf("tableInfo: %s\n", tableInfo)
	t = CSVTable{&readerCopy}
	return FprintfTable(os.Stdout, NewDefaultTableConfig(true, true), t, tableInfo)
}

// PrintSimpleTable -
func PrintSimpleTable(data [][]string) error {
	t := SimpleTable{data}
	tableInfo, err := GetTableInfo(t)
	if err != nil {
		return err
	}
	Logger.Printf("tableInfo: %s\n", tableInfo)
	return FprintfTable(os.Stdout, NewDefaultTableConfig(true, true), t, tableInfo)
}

// FprintfTable -
func FprintfTable(w io.Writer, config TableConfig, t Table, tableInfo *TableInfo) error {
	var err error
	if tableInfo == nil {
		tableInfo, err = GetTableInfo(t)
		if err != nil {
			return err
		}
	}
	header := true
	rowCounter := 0
	for row := range t.RowIterator() {
		if row.Error != nil {
			return row.Error
		}
		if header {
			printHorizontalLine("┌", "┬", "┐", "─", tableInfo, config)
			header = false
		}
		for i := 0; i < tableInfo.RowHeights[rowCounter]; i++ {
			if config.ColumnEdges {
				fmt.Printf("│")
			} else {
				fmt.Printf("")
			}
			for j := 0; j < tableInfo.Columns; j++ {
				if j > 0 {
					fmt.Printf("│")
				}
				if len(row.Fields) <= j {
					fmt.Printf(" %-"+strconv.Itoa(tableInfo.ColumnWidths[j])+"s ", " ")
					continue
				}
				multiLine := strings.Split(row.Fields[j], "\n")
				if len(multiLine) > i {
					fmt.Printf(" %-"+strconv.Itoa(tableInfo.ColumnWidths[j])+"s ", multiLine[i])
				} else {
					fmt.Printf(" %-"+strconv.Itoa(tableInfo.ColumnWidths[j])+"s ", " ")
				}
			}
			if config.ColumnEdges {
				fmt.Println("│")
			} else {
				fmt.Println("")
			}
		}
		if rowCounter+1 < tableInfo.Rows {
			if config.LineBetweenRows {
				printHorizontalLine("├", "┼", "┤", "─", tableInfo, config)
			}
		}
		rowCounter++
	}
	printHorizontalLine("└", "┴", "┘", "─", tableInfo, config)
	return nil
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

// CSVTable -
type CSVTable struct {
	Reader io.Reader
}

// RowIterator - Implements the Table interface.
func (t CSVTable) RowIterator() <-chan Row {
	return CSVRowIterator(t.Reader)
}

// CSVRowIterator -
func CSVRowIterator(reader io.Reader) <-chan Row {
	c := make(chan Row)
	go func() {
		r := csv.NewReader(reader)
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
			if len(multiLine) > 1 {
				l := 0
				for _, d := range multiLine {
					if len(d) > l {
						l = len(d)
					}
				}
				rowColumnWidths[i] = l
				rowRows[i] = len(multiLine)
			} else {
				rowColumnWidths[i] = len(cData)
				rowRows[i] = 1
			}
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
