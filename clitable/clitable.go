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
