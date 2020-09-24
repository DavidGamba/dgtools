// This file is part of go-utils.
//
// Copyright (C) 2016-2019  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package yamlutils - Utilities to read yml files like if using xpath
*/
package yamlutils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	// "reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Logger - Custom lib logger
var Logger = log.New(ioutil.Discard, "yamlutils ", log.LstdFlags)

// YML object
type YML struct {
	Tree interface{}
}

// NewFromFile returns a pointer to a YML object from a file.
func NewFromFile(filename string) (*YML, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var tree interface{}
	err = yaml.Unmarshal(data, &tree)
	if err != nil {
		return nil, err
	}
	return &YML{Tree: tree}, nil
}

// NewFromReader returns a pointer to a YML object from an io.Reader.
func NewFromReader(reader io.Reader) (*YML, error) {
	var tree interface{}
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(buf.Bytes(), &tree)
	if err != nil {
		return nil, err
	}
	return &YML{Tree: tree}, nil
}

// NewFromString - returns a pointer to a YML object from a string.
func NewFromString(str string) (*YML, error) {
	var tree interface{}
	err := yaml.Unmarshal([]byte(str), &tree)
	if err != nil {
		return nil, err
	}
	return &YML{Tree: tree}, nil
}

// GetString returns a string designated by path.
// Path is a string with elements separated by /.
// Array indexes are given as a number.
// For example: "level1/level2/3/level4"
func (y *YML) GetString(include bool, keys []string) (string, error) {
	path := strings.Join(keys, ",")
	target, _, errPath := NavigateTree(include, y.Tree, keys)
	// Check if response is a single element
	switch o := target.(type) {
	case string, int, uint, float32, float64, bool:
		if errPath != nil {
			return fmt.Sprintf("%v", o), fmt.Errorf("yaml path '%s' didn't return a valid string: %w", path, errPath)
		}
		return fmt.Sprintf("%v", o), nil
	}
	// Marshal complex response
	out, err := yaml.Marshal(target)
	if errPath != nil {
		return string(out), fmt.Errorf("yaml path '%s' didn't return a valid string: %w", path, errPath)
	}
	if err != nil {
		return string(out), fmt.Errorf("failed to Marshal output: %w", err)
	}
	Logger.Printf("%s", out)
	return string(out), nil
}

func (y *YML) AddString(keys []string, input string) (string, error) {
	path := strings.Join(keys, ",")
	errPath := AddChildToTree(&y.Tree, &y.Tree, keys, input)
	// Check if response is a single element
	switch o := y.Tree.(type) {
	case string, int, uint, float32, float64, bool:
		if errPath != nil {
			return fmt.Sprintf("%v", o), fmt.Errorf("yaml path '%s' didn't return a valid string: %w", path, errPath)
		}
		return fmt.Sprintf("%v", o), nil
	}
	// Marshal complex response
	out, err := yaml.Marshal(y.Tree)
	if errPath != nil {
		if errors.Is(errPath, ErrInvalidChildTypeKeyValue) {
			return string(out), errPath
		}
		if errors.Is(errPath, ErrInvalidParentType) {
			return string(out), fmt.Errorf("yaml path '%s': %w", path, errPath)
		}
		return string(out), fmt.Errorf("yaml path '%s' didn't return a valid string: %w", path, errPath)
	}
	if err != nil {
		return string(out), fmt.Errorf("failed to Marshal output: %w", err)
	}
	Logger.Printf("%s", out)
	return string(out), nil
}

// ErrExtraElementsInPath - Indicates when there is a final match and there are remaining path elements.
var ErrExtraElementsInPath = fmt.Errorf("extra elements in path")

// ErrMapKeyNotFound - Key not in config.
var ErrMapKeyNotFound = fmt.Errorf("map key not found")

// ErrNotAnIndex - The given path is not a numerical index and the element is of type slice/array.
var ErrNotAnIndex = fmt.Errorf("not an index")

// ErrInvalidIndex - The given index is invalid.
var ErrInvalidIndex = fmt.Errorf("invalid index")

// NavigateTree allows you to define a path string to traverse a tree composed of maps and arrays.
// To navigate through slices/arrays use a numerical index, for example: [path to array 1]
// When include is true, the returned map will have the key as part of it.
func NavigateTree(include bool, m interface{}, p []string) (interface{}, []string, error) {
	// Logger.Printf("type: %v, path: %v\n", reflect.TypeOf(m), p)
	path := strings.Join(p, "/")
	Logger.Printf("NavigateTree: Self: %v, Input path: '%s'", include, path)
	if len(p) <= 0 {
		return m, p, nil
	}
	switch m.(type) {
	case map[interface{}]interface{}:
		Logger.Printf("NavigateTree: map type")
		t, ok := m.(map[interface{}]interface{})[p[0]]
		if !ok {
			return m, p, fmt.Errorf("%w: %s", ErrMapKeyNotFound, p[0])
		}
		if include && len(p) == 1 {
			Logger.Printf("NavigateTree: self return")
			return map[interface{}]interface{}{p[0]: m.(map[interface{}]interface{})[p[0]]}, p[1:], nil
		}
		return NavigateTree(include, t, p[1:])
	case []interface{}:
		Logger.Printf("NavigateTree: slice/array type")

		index, err := strconv.Atoi(p[0])
		if err != nil {
			return m, p, fmt.Errorf("%w: %s", ErrNotAnIndex, p[0])
		}
		if index < 0 || len(m.([]interface{})) <= index {
			return m, p, fmt.Errorf("%w: %s", ErrInvalidIndex, p[0])
		}
		return NavigateTree(include, m.([]interface{})[index], p[1:])
	default:
		Logger.Printf("NavigateTree: single element type")
		return m, p, fmt.Errorf("%w: %s", ErrExtraElementsInPath, strings.Join(p, "/"))
	}
}

// ErrInvalidParentType - The parent type is invalid.
var ErrInvalidParentType = fmt.Errorf("invalid parent type, must be list or key/value")

// ErrInvalidChildTypeKeyValue - The child type is invalid.
var ErrInvalidChildTypeKeyValue = fmt.Errorf("invalid child type, must be 'key: value'")

func AddChild(m *interface{}, child string) error {
	Logger.Printf("AddChild: %v", *m)
	// Logger.Printf("type: %v\n", reflect.TypeOf(*m))
	var tree interface{}
	err := yaml.Unmarshal([]byte(child), &tree)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%w", ErrInvalidParentType)
	}
	switch (*m).(type) {
	case map[interface{}]interface{}:
		Logger.Printf("AddChild: map type")
		if t, ok := tree.(map[interface{}]interface{}); ok {
			for k, v := range t {
				(*m).(map[interface{}]interface{})[k] = v
			}
			return nil
		}
		return fmt.Errorf("%w", ErrInvalidChildTypeKeyValue)
	case []interface{}:
		Logger.Printf("AddChild: slice/array type")
		r := append((*m).([]interface{}), tree)
		*m = r
		return nil
	default:
		Logger.Printf("AddChild: single element type")
		return fmt.Errorf("%w", ErrInvalidParentType)
	}
	return nil
}

func AddChildToTree(parent *interface{}, current *interface{}, p []string, child string) error {
	path := strings.Join(p, "/")
	Logger.Printf("AddChildToTree: Input path: '%s'", path)
	if len(p) <= 0 {
		Logger.Printf("Before %v, %v\n", *parent, *current)
		err := AddChild(current, child)
		if err != nil {
			return err
		}
		Logger.Printf("After %v, %v\n", *parent, *current)
		return nil
	}
	switch t := (*current).(type) {
	case map[interface{}]interface{}:
		Logger.Printf("AddChildToTree: map type")
		e, ok := t[p[0]]
		if !ok {
			return fmt.Errorf("%w: %s", ErrMapKeyNotFound, p[0])
		}
		err := AddChildToTree(current, &e, p[1:], child)
		if err != nil {
			return err
		}
		(*current).(map[interface{}]interface{})[p[0]] = e
		return nil
	case []interface{}:
		Logger.Printf("AddChildToTree: slice/array type")
		index, err := strconv.Atoi(p[0])
		if err != nil {
			return fmt.Errorf("%w: %s", ErrNotAnIndex, p[0])
		}
		if index < 0 || len(t) <= index {
			return fmt.Errorf("%w: %s", ErrInvalidIndex, p[0])
		}
		err = AddChildToTree(current, &t[index], p[1:], child)
		if err != nil {
			return err
		}
		parent = current
		return nil
	default:
		Logger.Printf("AddChildToTree: single element type")
		return fmt.Errorf("%w: %s", ErrExtraElementsInPath, strings.Join(p, "/"))
	}
}
