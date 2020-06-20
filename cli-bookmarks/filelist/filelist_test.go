// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package filelist

import (
	"reflect"
	"testing"
)

func Test_listFiles(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"test_files/test_tree listing",
			args{"../test_files/test_tree"},
			[]string{
				"../test_files/test_tree/.hiddenA",
				"../test_files/test_tree/.hiddenB",
				"../test_files/test_tree/DirA",
				"../test_files/test_tree/DirB",
				"../test_files/test_tree/dirA",
				"../test_files/test_tree/dirB",
				"../test_files/test_tree/pathA",
				"../test_files/test_tree/pathB"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listFiles(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("listFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listFilesWithFilter(t *testing.T) {
	type args struct {
		dir          string
		prefixFilter string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"test_files/test_tree listing",
			args{"../test_files/test_tree", ""},
			[]string{
				"../test_files/test_tree/.hiddenA",
				"../test_files/test_tree/.hiddenB",
				"../test_files/test_tree/DirA",
				"../test_files/test_tree/DirB",
				"../test_files/test_tree/dirA",
				"../test_files/test_tree/dirB",
				"../test_files/test_tree/pathA",
				"../test_files/test_tree/pathB"},
			false,
		},
		{
			"test_files/test_tree listing filtered",
			args{"../test_files/test_tree", "d"},
			[]string{
				"../test_files/test_tree/DirA",
				"../test_files/test_tree/DirB",
				"../test_files/test_tree/dirA",
				"../test_files/test_tree/dirB"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListFilesWithFilter(tt.args.dir, tt.args.prefixFilter)
			if (err != nil) != tt.wantErr {
				t.Errorf("listFilesWithFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listFilesWithFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
