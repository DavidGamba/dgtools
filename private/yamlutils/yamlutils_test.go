// This file is part of go-utils.
//
// Copyright (C) 2019  David Gamba Rios
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
		{"extra elements in path", false, []string{"hello"}, "hello", "hello", ErrExtraElementsInPath},
		{"simple", false, []string{"hello"}, "hello: world", "world", nil},
		{"simple", false, []string{"x"}, "hello: world", "hello: world\n", ErrMapKeyNotFound},
		{"simple", false, []string{"hello", "1"}, `hello:
  - one
  - two
  - three`, "two", nil},
		{"simple", false, []string{"hello", "3"}, `hello:
  - one
  - two
  - three`, "- one\n- two\n- three\n", ErrInvalidIndex},
		{"simple", false, []string{"hello", "-1"}, `hello:
  - one
  - two
  - three`, "- one\n- two\n- three\n", ErrInvalidIndex},
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

func TestNavigateTree(t *testing.T) {
	tests := []struct {
		name         string
		include      bool
		path         []string
		input        interface{}
		expected     interface{}
		expectedPath []string
		err          error
	}{
		{"string config, no path", false,
			[]string{}, "hola", "hola", []string{}, nil},
		{"string config, bad path", false,
			[]string{"extra", "elements"}, "hola", "hola", []string{"extra", "elements"}, ErrExtraElementsInPath},
		{"array config, no path", false,
			[]string{}, []interface{}{"one", "two", "three"}, []interface{}{"one", "two", "three"}, []string{}, nil},
		{"array config, path", false,
			[]string{"1"}, []interface{}{"one", "two", "three"}, "two", []string{}, nil},
		{"array config, bad path", false,
			[]string{"[1]"}, []interface{}{"one", "two", "three"}, []interface{}{"one", "two", "three"}, []string{"[1]"}, ErrNotAnIndex},
		{"map config, no path", false,
			[]string{},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}},
			[]string{}, nil},
		{"map config, path", false,
			[]string{"map"},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}, "another": []string{"four", "five", "six"}},
			[]string{"one", "two", "three"},
			[]string{}, nil},
		{"map config, path", true,
			[]string{"map"},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}, "another": []string{"four", "five", "six"}},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}},
			[]string{}, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := ""
			buf := bytes.NewBufferString(s)
			Logger.SetOutput(buf)
			output, path, err := NavigateTree(test.include, test.input, test.path)
			if !errors.Is(err, test.err) {
				t.Errorf("Unexpected error: %s\n", err)
			}
			if !reflect.DeepEqual(path, test.expectedPath) {
				t.Errorf("Expected:\n%#v\nGot:\n%#v\n", test.expectedPath, path)
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
		structure   interface{}
		childString string
		expected    interface{}
		err         error
	}{
		{"simple", nil, "hello", nil, ErrInvalidParentType},
		{"simple", "str", "hello", "str", ErrInvalidParentType},
		{"simple", 123, "hello", 123, ErrInvalidParentType},
		{"array", []interface{}{1, 2, 3}, "hello", []interface{}{1, 2, 3, "hello"}, nil},
		{"array", []interface{}{1, 2, 3}, "hello: world", []interface{}{1, 2, 3, map[interface{}]interface{}{"hello": "world"}}, nil},
		{"array", []interface{}{1, 2, 3}, "hello: world", []interface{}{1, 2, 3, map[interface{}]interface{}{"hello": "world"}}, nil},
		{"map",
			map[interface{}]interface{}{"map": []string{"one"}, "another": []string{"two"}},
			"hello",
			map[interface{}]interface{}{"map": []string{"one"}, "another": []string{"two"}}, ErrInvalidChildTypeKeyValue},
		{"map",
			map[interface{}]interface{}{"map": []string{"one"}, "another": []string{"two"}},
			"hello: world",
			map[interface{}]interface{}{"map": []string{"one"}, "another": []string{"two"}, "hello": "world"}, nil},
		{"map",
			map[interface{}]interface{}{"map": []string{"one"}, "another": []string{"two"}},
			"map: world",
			map[interface{}]interface{}{"map": "world", "another": []string{"two"}}, nil},
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
		structure   interface{}
		expected    interface{}
		childString string
		err         error
	}{
		{"simple", []string{}, "hola", "hola", "hello", ErrInvalidParentType},
		{"array", []string{}, []interface{}{"hola"}, []interface{}{"hola", "hello"}, "hello", nil},
		{"map", []string{}, map[interface{}]interface{}{"map": "hola"}, map[interface{}]interface{}{"map": "hola", "hello": "world"}, "hello: world", nil},
		{"map", []string{"map"}, map[interface{}]interface{}{"map": []interface{}{"one", "two", "three"}},
			map[interface{}]interface{}{"map": []interface{}{"one", "two", "three", "four"}}, "four", nil},
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
