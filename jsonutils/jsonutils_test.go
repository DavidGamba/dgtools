// This file is part of dgtools.
//
// Copyright (C) 2019-2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package jsonutils

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/DavidGamba/dgtools/trees"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		include  bool
		path     []string
		input    string
		expected string
		err      error
	}{
		{"simple 0", false, []string{}, `"hello"`, "hello", nil},
		{"extra elements in path", false, []string{"hello"}, `"hello"`, "hello", trees.ErrExtraElementsInPath},
		{"simple 1", false, []string{"hello"}, `{"hello": "world"}`, "world", nil},
		{"simple 2", false, []string{"x"}, `{"hello": "world"}`, `{
  "hello": "world"
}`, trees.ErrMapKeyNotFound},
		{"simple 3", false, []string{"hello", "1"}, `{"hello": ["one", "two", "three"]}`, "two", nil},
		{"simple 4", false, []string{"hello", "3"}, `{"hello": [ "one", "two", "three"]}`, `[
  "one",
  "two",
  "three"
]`, trees.ErrInvalidIndex},
		{"simple 5", false, []string{"hello", "-1"}, `{"hello": ["one", "two", "three"]}`, `[
  "one",
  "two",
  "three"
]`, trees.ErrInvalidIndex},
		{"simple 6", false, []string{"hello", "1", "world"}, `{"hello": ["one", {"world": "hola"}, "three"]}`, `hola`, nil},
		{"simple 7", false, []string{"hello", "1", "world"}, `{"hello": ["one", {"world": true}, "three"]}`, "true", nil},
		{"simple 8", false, []string{"hello", "1", "world"}, `{"hello": ["one", {"world": 123}, "three"]}`, "123", nil},
		{"simple 9", false, []string{"hello", "1", "world"}, `{"hello": ["one", {"world": 123.123}, "three"]}`, "123.123", nil},
		{"simple 10", true, []string{"hello", "1", "world"}, `{"hello": ["one", {"world": 123.123}, "three"]}`, `{
  "world": 123.123
}`, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := ""
			buf := bytes.NewBufferString(s)
			Logger.SetOutput(buf)
			yml, err := NewFromString(test.input)
			if err != nil {
				t.Fatalf("Unexpected error: %s\n", err)
			}
			output, err := yml.GetString(test.include, test.path)
			if !errors.Is(err, test.err) {
				t.Errorf("Unexpected error: %s\n", err)
			}
			if !reflect.DeepEqual(output, test.expected) {
				t.Errorf("Expected:\n%#v\nGot:\n%#v\n", test.expected, output)
			}
			t.Log(buf.String())
		})
	}

}
