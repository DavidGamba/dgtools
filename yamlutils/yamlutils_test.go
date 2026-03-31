// This file is part of dgtools.
//
// # Copyright (C) 2019-2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package yamlutils

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
		{"simple", false, []string{}, "hello", "hello", nil},
		{"extra elements in path", false, []string{"hello"}, "hello", "hello", trees.ErrExtraElementsInPath},
		{"simple", false, []string{"hello"}, "hello: world", "world", nil},
		{"simple", false, []string{"x"}, "hello: world", "hello: world\n", trees.ErrMapKeyNotFound},
		{"simple", false, []string{"hello", "1"}, `hello:
  - one
  - two
  - three`, "two", nil},
		{"simple", false, []string{"hello", "3"}, `hello:
  - one
  - two
  - three`, "- one\n- two\n- three\n", trees.ErrInvalidIndex},
		{"simple", false, []string{"hello", "-1"}, `hello:
  - one
  - two
  - three`, "- one\n- two\n- three\n", trees.ErrInvalidIndex},
		{"simple", false, []string{"hello", "1", "world"}, `hello:
  - one
  - world: hola
  - three`, "hola", nil},
		{"simple", false, []string{"hello", "1", "world"}, `hello:
  - one
  - world: true
  - three`, "true", nil},
		{"simple", false, []string{"hello", "1", "world"}, `hello:
  - one
  - world: 123
  - three`, "123", nil},
		{"simple", false, []string{"hello", "1", "world"}, `hello:
  - one
  - world: 123.123
  - three`, "123.123", nil},
		{"simple", true, []string{"hello", "1", "world"}, `hello:
  - one
  - world: 123.123
  - three`, "world: 123.123\n", nil},
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

func TestAddChild(t *testing.T) {
	tests := []struct {
		name        string
		structure   any
		childString string
		expected    any
		err         error
	}{
		{"simple", nil, "hello", nil, ErrInvalidParentType},
		{"simple", "str", "hello", "str", ErrInvalidParentType},
		{"simple", 123, "hello", 123, ErrInvalidParentType},
		{"array", []any{1, 2, 3}, "hello", []any{1, 2, 3, "hello"}, nil},
		{"array", []any{1, 2, 3}, "hello: world", []any{1, 2, 3, map[string]any{"hello": "world"}}, nil},
		{"array", []any{1, 2, 3}, "hello: world", []any{1, 2, 3, map[string]any{"hello": "world"}}, nil},
		{"map error",
			map[any]any{"map": []string{"one"}, "another": []string{"two"}},
			"hello",
			map[any]any{"map": []string{"one"}, "another": []string{"two"}}, ErrInvalidChildTypeKeyValue},
		{"map key value",
			map[any]any{"map": []string{"one"}, "another": []string{"two"}},
			"hello: world",
			map[any]any{"map": []string{"one"}, "another": []string{"two"}, "hello": "world"}, nil},
		{"map",
			map[any]any{"map": []string{"one"}, "another": []string{"two"}},
			"map: world",
			map[any]any{"map": "world", "another": []string{"two"}}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := ""
			buf := bytes.NewBufferString(s)
			Logger.SetOutput(buf)
			err := AddChild(&test.structure, test.childString)
			if !errors.Is(err, test.err) {
				t.Errorf("Unexpected error: %s\n", err)
			}
			if !reflect.DeepEqual(test.structure, test.expected) {
				t.Errorf("Expected:\n%#v\nGot:\n%#v\n", test.expected, test.structure)
			}
			t.Log(buf.String())
		})
	}
}

func TestAddChildToTree(t *testing.T) {
	tests := []struct {
		name        string
		path        []string
		structure   any
		expected    any
		childString string
		err         error
	}{
		{"simple", []string{}, "hola", "hola", "hello", ErrInvalidParentType},
		{"array", []string{}, []any{"hola"}, []any{"hola", "hello"}, "hello", nil},
		{"map", []string{}, map[any]any{"map": "hola"}, map[any]any{"map": "hola", "hello": "world"}, "hello: world", nil},
		{"map", []string{"map"}, map[any]any{"map": []any{"one", "two", "three"}},
			map[any]any{"map": []any{"one", "two", "three", "four"}}, "four", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := ""
			buf := bytes.NewBufferString(s)
			Logger.SetOutput(buf)
			err := AddChildToTree(&test.structure, &test.structure, test.path, test.childString)
			if !errors.Is(err, test.err) {
				t.Errorf("Unexpected error: %s\n", err)
			}
			if !reflect.DeepEqual(test.structure, test.expected) {
				t.Errorf("Expected:\n%#v\nGot:\n%#v\n", test.expected, test.structure)
			}
			t.Log(buf.String())
		})
	}
}
