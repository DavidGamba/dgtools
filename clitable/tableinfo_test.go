// This file is part of clitable.
//
// Copyright (C) 2017-2022  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package clitable_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/DavidGamba/dgtools/clitable"
)

func Test_GetTableInfo(t *testing.T) {
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *clitable.TableInfo
		wantErr bool
	}{
		{"single column, single row",
			args{bytes.NewBufferString("hello")},
			&clitable.TableInfo{
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
			&clitable.TableInfo{
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
			&clitable.TableInfo{
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
			&clitable.TableInfo{
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
			&clitable.TableInfo{
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
			&clitable.TableInfo{
				Columns: 0,
				Rows:    0,
			},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := clitable.GetTableInfo(clitable.CSVTable{tt.args.reader})
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

func TestGetTableInfoSimpleTable(t *testing.T) {
	type args struct {
		t clitable.Table
	}
	tests := []struct {
		name    string
		args    args
		want    *clitable.TableInfo
		wantErr bool
	}{
		{
			"Basic table 1x1",
			args{clitable.SimpleTable{[][]string{{"a"}}}},
			&clitable.TableInfo{
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
			args{clitable.SimpleTable{[][]string{{"a", "b"}, {"a", "b"}}}},
			&clitable.TableInfo{
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
			got, err := clitable.GetTableInfo(tt.args.t)
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
