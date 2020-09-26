// This file is part of csv-table.
//
// Copyright (C) 2017-2019  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package csvtable

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func Test_GetTableInfo(t *testing.T) {
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    TableInfo
		wantErr bool
	}{
		{"single column, single row",
			args{bytes.NewBufferString("hello")},
			TableInfo{
				Columns:            1,
				Rows:               1,
				PerRowColumnWidths: [][]int{{5}},
				PerRowRows:         [][]int{{1}},
				ColumnWidths:       []int{5},
				RowHeights:         []int{1},
			},
			false},
		{"multi column, single row",
			args{bytes.NewBufferString("a,bb")},
			TableInfo{
				Columns:            2,
				Rows:               1,
				PerRowColumnWidths: [][]int{{1, 2}},
				PerRowRows:         [][]int{{1, 1}},
				ColumnWidths:       []int{1, 2},
				RowHeights:         []int{1},
			},
			false},
		{"multi column, multi row",
			args{bytes.NewBufferString(`a,bb
ccc,dddd`)},
			TableInfo{
				Columns:            2,
				Rows:               2,
				PerRowColumnWidths: [][]int{{1, 2}, {3, 4}},
				PerRowRows:         [][]int{{1, 1}, {1, 1}},
				ColumnWidths:       []int{3, 4},
				RowHeights:         []int{1, 1},
			},
			false},
		{"multi column, multi row, uneven rows",
			args{bytes.NewBufferString(`a
ccc,dddd`)},
			TableInfo{
				Columns:            2,
				Rows:               2,
				PerRowColumnWidths: [][]int{{1}, {3, 4}},
				PerRowRows:         [][]int{{1}, {1, 1}},
				ColumnWidths:       []int{3, 4},
				RowHeights:         []int{1, 1},
			},
			false},
		{"multi column, multi row, multiline column",
			args{bytes.NewBufferString(`a,"bb
bbbbb
bb",ccc
dddd,ee,fff`)},
			TableInfo{
				Columns:            3,
				Rows:               2,
				PerRowColumnWidths: [][]int{{1, 5, 3}, {4, 2, 3}},
				PerRowRows:         [][]int{{1, 3, 1}, {1, 1, 1}},
				ColumnWidths:       []int{4, 5, 3},
				RowHeights:         []int{3, 1},
			},
			false},
		{"bad input",
			args{bytes.NewBufferString(`a,"bb`)},
			TableInfo{
				Columns: 0,
				Rows:    0,
			},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTableInfo(CSVTable{tt.args.reader})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTableInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTableInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTableInfo(t *testing.T) {
	type args struct {
		t Table
	}
	tests := []struct {
		name    string
		args    args
		want    TableInfo
		wantErr bool
	}{
		{
			"Basic table 1x1",
			args{SimpleTable{[][]string{{"a"}}}},
			TableInfo{
				Columns:            1,
				Rows:               1,
				PerRowColumnWidths: [][]int{{1}},
				PerRowRows:         [][]int{{1}},
				ColumnWidths:       []int{1},
				RowHeights:         []int{1},
			},
			false,
		},
		{
			"Basic table 2x2",
			args{SimpleTable{[][]string{{"a", "b"}, {"a", "b"}}}},
			TableInfo{
				Columns:            2,
				Rows:               2,
				PerRowColumnWidths: [][]int{{1, 1}, {1, 1}},
				PerRowRows:         [][]int{{1, 1}, {1, 1}},
				ColumnWidths:       []int{1, 1},
				RowHeights:         []int{1, 1},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTableInfo(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTableInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTableInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
