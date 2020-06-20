// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this

package completion

import (
	"reflect"
	"testing"
)

func Test_fileListCompletion(t *testing.T) {
	type args struct {
		alias    string
		filePath string
		entry    string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"t alias",
			args{"t", "../test_files/test_tree", "t"},
			[]string{
				"t/.hiddenA",
				"t/.hiddenB",
				"t/DirA",
				"t/DirB",
				"t/dirA",
				"t/dirB",
				"t/pathA",
				"t/pathB"},
			false,
		},
		{
			"t alias with filepath separator",
			args{"t", "../test_files/test_tree", "t/"},
			[]string{
				"t/.hiddenA",
				"t/.hiddenB",
				"t/DirA",
				"t/DirB",
				"t/dirA",
				"t/dirB",
				"t/pathA",
				"t/pathB"},
			false,
		},
		{
			"t alias filtered",
			args{"t", "../test_files/test_tree", "t/d"},
			[]string{
				"t/DirA",
				"t/DirB",
				"t/dirA",
				"t/dirB"},
			false,
		},
		{
			"test alias",
			args{"test", "../test_files/test_tree", "test"},
			[]string{
				"test/.hiddenA",
				"test/.hiddenB",
				"test/DirA",
				"test/DirB",
				"test/dirA",
				"test/dirB",
				"test/pathA",
				"test/pathB"},
			false,
		},
		{
			"t alias filtered",
			args{"t", "../test_files/test_tree", "t/pathA/"},
			[]string{"t/pathA/pathAdirA", "t/pathA/pathAdirB"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fileListCompletion(tt.args.alias, tt.args.filePath, tt.args.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("fileListCompletion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fileListCompletion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_matchAlias(t *testing.T) {
	type args struct {
		bookmarks map[string]string
		entry     string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"match aliases",
			args{map[string]string{"key": "value", "Key": "Value", "key2": "value2", "other": "otherValue"}, ""},
			[]string{"Key", "key", "key2", "other"},
		},
		{
			"match aliases filtered",
			args{map[string]string{"key": "value", "Key": "Value", "key2": "value2", "other": "otherValue"}, "k"},
			[]string{"key", "key2"},
		},
		{
			"match aliases filtered full match",
			args{map[string]string{"key": "value", "Key": "Value", "key2": "value2", "other": "otherValue"}, "key"},
			[]string{"key/", "key2"},
		},
		{
			"match aliases filtered single match",
			args{map[string]string{"key": "value", "Key": "Value", "key2": "value2", "other": "otherValue"}, "o"},
			[]string{"other/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchAlias(tt.args.bookmarks, tt.args.entry); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchAlias() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_completionResults(t *testing.T) {
	type args struct {
		bookmarks map[string]string
		entry     string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"match aliases",
			args{map[string]string{"key": "value", "Key": "Value", "key2": "value2", "other": "otherValue"}, ""},
			[]string{"Key", "key", "key2", "other"},
			false,
		},
		{
			"match aliases filtered",
			args{map[string]string{"key": "value", "Key": "Value", "key2": "value2", "other": "otherValue"}, "k"},
			[]string{"key", "key2"},
			false,
		},
		{
			"match aliases filtered full match",
			args{map[string]string{"key": "value", "Key": "Value", "key2": "value2", "other": "otherValue"}, "key"},
			[]string{"key/", "key2"},
			false,
		},
		{
			"match aliases filtered single match",
			args{map[string]string{"key": "value", "Key": "Value", "key2": "value2", "other": "otherValue"}, "o"},
			[]string{"other/"},
			false,
		},
		{
			"select alias",
			args{map[string]string{"key": "../test_files/test_tree", "Key": "Value", "key2": "value2", "other": "otherValue"}, "key/"},
			[]string{
				"key/.hiddenA",
				"key/.hiddenB",
				"key/DirA",
				"key/DirB",
				"key/dirA",
				"key/dirB",
				"key/pathA",
				"key/pathB"},
			false,
		},
		{
			"wrong alias",
			args{map[string]string{"key": "../test_files/test_tree", "Key": "Value", "key2": "value2", "other": "otherValue"}, "test/"},
			[]string{},
			true,
		},
		{
			"select alias and filter",
			args{map[string]string{"key": "../test_files/test_tree", "Key": "Value", "key2": "value2", "other": "otherValue"}, "key/d"},
			[]string{
				"key/DirA",
				"key/DirB",
				"key/dirA",
				"key/dirB"},
			false,
		},
		{
			"select alias with single dir match",
			args{map[string]string{"key": "../test_files/test_tree", "Key": "Value", "key2": "value2", "other": "otherValue"}, "key/pathA"},
			[]string{"key/pathA/"},
			false,
		},
		{
			"select alias and filter dir",
			args{map[string]string{"key": "../test_files/test_tree", "Key": "Value", "key2": "value2", "other": "otherValue"}, "key/pathA/"},
			[]string{"key/pathA/pathAdirA", "key/pathA/pathAdirB"},
			false,
		},
		{
			"select alias and filter dir",
			args{map[string]string{"key": "../test_files/test_tree", "Key": "Value", "key2": "value2", "other": "otherValue"}, "key/pathA/pathAdirA"},
			[]string{"key/pathA/pathAdirA/"},
			false,
		},
		{
			"select alias and filter dir",
			args{map[string]string{"key": "../test_files/test_tree", "Key": "Value", "key2": "value2", "other": "otherValue"}, "key/pathA/pathAdirA/"},
			[]string{"key/pathA/pathAdirA/pathAdirAsubdirA", "key/pathA/pathAdirA/pathAdirAsubdirB"},
			// []string{"key/pathA/pathAdirA/"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompletionResults(tt.args.bookmarks, tt.args.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("completionResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("completionResults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPrefixFilter(t *testing.T) {
	type args struct {
		alias string
		entry string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"prefix", args{"ansible", "ansible/lib"}, "lib"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPrefixFilter(tt.args.alias, tt.args.entry); got != tt.want {
				t.Errorf("getPrefixFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
