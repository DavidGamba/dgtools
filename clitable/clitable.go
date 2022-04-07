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

type TablePrinter struct {
	tableConfig tableConfig
}

type tableConfig struct {
	HasHeader bool

	TopLine     bool
	TopStart    string
	TopEnd      string
	TopJuncture string
	TopBody     string

	HeaderDividerLine     bool
	HeaderDividerStart    string
	HeaderDividerEnd      string
	HeaderDividerJuncture string
	HeaderDividerBody     string

	BottomLine     bool
	BottomStart    string
	BottomEnd      string
	BottomJuncture string
	BottomBody     string

	DividerLine     bool
	DividerStart    string
	DividerEnd      string
	DividerJuncture string
	DividerBody     string

	Column      string
	ColumnEdges bool
}

func NewTablePrinter() *TablePrinter {
	tp := &TablePrinter{}
	tp.SetStyle(Full)
	tp.tableConfig.HasHeader = true
	return tp
}

type Style int

const (
	Full Style = iota
	Ascii
	Compact
	Space
)

func (tp *TablePrinter) HasHeader(b bool) *TablePrinter {
	tp.tableConfig.HasHeader = b
	return tp
}

func (tp *TablePrinter) SetStyle(s Style) *TablePrinter {
	switch s {
	case Full:
		tp.tableConfig.TopLine = true
		tp.tableConfig.TopStart = "┌"
		tp.tableConfig.TopEnd = "┐"
		tp.tableConfig.TopJuncture = "┬"
		tp.tableConfig.TopBody = "─"

		tp.tableConfig.HeaderDividerLine = true
		tp.tableConfig.HeaderDividerStart = "╞"
		tp.tableConfig.HeaderDividerEnd = "╡"
		tp.tableConfig.HeaderDividerJuncture = "╪"
		tp.tableConfig.HeaderDividerBody = "═"

		tp.tableConfig.BottomLine = true
		tp.tableConfig.BottomStart = "└"
		tp.tableConfig.BottomEnd = "┘"
		tp.tableConfig.BottomJuncture = "┴"
		tp.tableConfig.BottomBody = "─"

		tp.tableConfig.DividerLine = true
		tp.tableConfig.DividerStart = "├"
		tp.tableConfig.DividerEnd = "┤"
		tp.tableConfig.DividerJuncture = "┼"
		tp.tableConfig.DividerBody = "─"

		tp.tableConfig.Column = "│"
		tp.tableConfig.ColumnEdges = true
	case Compact:
		tp.tableConfig.TopLine = true
		tp.tableConfig.TopStart = "┌"
		tp.tableConfig.TopEnd = "┐"
		tp.tableConfig.TopJuncture = "┬"
		tp.tableConfig.TopBody = "─"

		tp.tableConfig.HeaderDividerLine = true
		tp.tableConfig.HeaderDividerStart = "╞"
		tp.tableConfig.HeaderDividerEnd = "╡"
		tp.tableConfig.HeaderDividerJuncture = "╪"
		tp.tableConfig.HeaderDividerBody = "═"

		tp.tableConfig.BottomLine = true
		tp.tableConfig.BottomStart = "└"
		tp.tableConfig.BottomEnd = "┘"
		tp.tableConfig.BottomJuncture = "┴"
		tp.tableConfig.BottomBody = "─"

		tp.tableConfig.DividerLine = false
		tp.tableConfig.DividerStart = "├"
		tp.tableConfig.DividerEnd = "┤"
		tp.tableConfig.DividerJuncture = "┼"
		tp.tableConfig.DividerBody = "─"

		tp.tableConfig.Column = "│"
		tp.tableConfig.ColumnEdges = false
	case Ascii:
		tp.tableConfig.TopLine = true
		tp.tableConfig.TopStart = "+"
		tp.tableConfig.TopEnd = "+"
		tp.tableConfig.TopJuncture = "+"
		tp.tableConfig.TopBody = "-"

		tp.tableConfig.HeaderDividerLine = true
		tp.tableConfig.HeaderDividerStart = "+"
		tp.tableConfig.HeaderDividerEnd = "+"
		tp.tableConfig.HeaderDividerJuncture = "+"
		tp.tableConfig.HeaderDividerBody = "="

		tp.tableConfig.BottomLine = true
		tp.tableConfig.BottomStart = "+"
		tp.tableConfig.BottomEnd = "+"
		tp.tableConfig.BottomJuncture = "+"
		tp.tableConfig.BottomBody = "-"

		tp.tableConfig.DividerLine = true
		tp.tableConfig.DividerStart = "+"
		tp.tableConfig.DividerEnd = "+"
		tp.tableConfig.DividerJuncture = "+"
		tp.tableConfig.DividerBody = "-"

		tp.tableConfig.Column = "|"
		tp.tableConfig.ColumnEdges = true
	case Space:
		tp.tableConfig.TopLine = false
		tp.tableConfig.BottomLine = false
		tp.tableConfig.DividerLine = false
		tp.tableConfig.HeaderDividerLine = false

		tp.tableConfig.Column = " "
		tp.tableConfig.ColumnEdges = false
	}
	return tp
}

func (tp *TablePrinter) Print(t Table) error {
	return tp.Fprint(os.Stdout, t)
}

func (tp *TablePrinter) Fprint(w io.Writer, t Table) error {
	tableInfo, err := GetTableInfo(t)
	if err != nil {
		return err
	}
	return tp.fprint(w, t, tableInfo)
}

func (tp *TablePrinter) FprintCSVReader(w io.Writer, r io.Reader) error {
	// TODO: I am sure I can make this simpler and maybe even not necessary
	var readerCopy bytes.Buffer
	reader := io.TeeReader(r, &readerCopy)
	t := CSVTable{reader}
	tableInfo, err := GetTableInfo(t)
	if err != nil {
		return err
	}
	Logger.Printf("tableInfo: %s\n", tableInfo)
	t = CSVTable{&readerCopy}
	return tp.fprint(os.Stdout, t, tableInfo)
}

func (tp *TablePrinter) fprint(w io.Writer, t Table, tableInfo *TableInfo) error {
	header := tp.tableConfig.TopLine
	rowCounter := 0
	for row := range t.RowIterator() {
		if row.Error != nil {
			return row.Error
		}
		if header {
			printTopLine(tp.tableConfig, tableInfo)
			header = false
		}
		for i := 0; i < tableInfo.RowHeights[rowCounter]; i++ {
			if tp.tableConfig.ColumnEdges {
				fmt.Printf(tp.tableConfig.Column)
			} else {
				fmt.Printf("")
			}
			for j := 0; j < tableInfo.Columns; j++ {
				if j > 0 {
					fmt.Printf(tp.tableConfig.Column)
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
			if tp.tableConfig.ColumnEdges {
				fmt.Println(tp.tableConfig.Column)
			} else {
				fmt.Println("")
			}
		}
		if rowCounter+1 < tableInfo.Rows {
			if rowCounter+1 == 1 && tp.tableConfig.HeaderDividerLine {
				if tp.tableConfig.HasHeader {
					printHeaderDividerLine(tp.tableConfig, tableInfo)
				} else {
					printDividerLine(tp.tableConfig, tableInfo)
				}
			} else if tp.tableConfig.DividerLine {
				printDividerLine(tp.tableConfig, tableInfo)
			}
		}
		rowCounter++
	}
	if tp.tableConfig.BottomLine {
		printBottomLine(tp.tableConfig, tableInfo)
	}
	return nil
}

func printTopLine(tableConfig tableConfig, tableInfo *TableInfo) {
	printHorizontalLine(tableConfig.TopStart, tableConfig.TopJuncture, tableConfig.TopEnd, tableConfig.TopBody, tableConfig.ColumnEdges, tableInfo)
}

func printBottomLine(tableConfig tableConfig, tableInfo *TableInfo) {
	printHorizontalLine(tableConfig.BottomStart, tableConfig.BottomJuncture, tableConfig.BottomEnd, tableConfig.BottomBody, tableConfig.ColumnEdges, tableInfo)
}

func printDividerLine(tableConfig tableConfig, tableInfo *TableInfo) {
	printHorizontalLine(tableConfig.DividerStart, tableConfig.DividerJuncture, tableConfig.DividerEnd, tableConfig.DividerBody, tableConfig.ColumnEdges, tableInfo)
}

func printHeaderDividerLine(tableConfig tableConfig, tableInfo *TableInfo) {
	printHorizontalLine(tableConfig.HeaderDividerStart, tableConfig.HeaderDividerJuncture, tableConfig.HeaderDividerEnd, tableConfig.HeaderDividerBody, tableConfig.ColumnEdges, tableInfo)
}

func printHorizontalLine(start, juncture, end, body string, columnEdges bool, tableInfo *TableInfo) {
	for i := 0; i < tableInfo.Columns; i++ {
		if i == 0 {
			if columnEdges {
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
			if columnEdges {
				fmt.Println(end)
			} else {
				fmt.Println("")
			}
		}
	}
}
