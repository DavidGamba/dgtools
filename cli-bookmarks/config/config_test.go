// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config

import (
	"reflect"
	"strings"
	"testing"
)

func Test_ParseFile(t *testing.T) {
	t.Run("Read proper confing file", func(t *testing.T) {
		c, err := ParseFile("../test_files/test_tree/cli-bookmarks.toml")
		if err != nil {
			t.Errorf("Failed to parse config file: %s", err)
		}
		if c.Bookmarks["test"] != "test" {
			t.Errorf("test bookmark didn't match test")
		}
		if c.Bookmarks["hello"] != "hello/world" {
			t.Errorf("hello bookmark didn't match hello/world")
		}
	})
	t.Run("Read missing config file", func(t *testing.T) {
		_, err := ParseFile("missing.file")
		if err == nil {
			t.Fatal("No errors returned!")
		}
		if !strings.Contains(err.Error(), "no such file or directory") {
			t.Fatalf("Wrong error: %s", err)
		}
	})
}

func Test_Parse(t *testing.T) {
	type args struct {
		configStr string
	}
	tests := []struct {
		name       string
		args       args
		want       Config
		wantErr    bool
		wantErrStr string
	}{
		{"Good config", args{`[bookmarks]
hello="world"`}, Config{Bookmarks: map[string]string{"hello": "world"}}, false, ""},
		{"Bad alias", args{`[bookmarks]
hello/world="world"`}, Config{Bookmarks: map[string]string{}}, true, "cannot contain '/'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.configStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.wantErrStr) {
					t.Errorf("Parse() error = %v, wantErrStr %v", err, tt.wantErrStr)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
