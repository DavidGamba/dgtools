// This file is part of dgtools.
//
// Copyright (C) 2016-2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package yamlutils - Utilities to read yml files like if using xpath
*/
package yamlutils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/DavidGamba/dgtools/trees"
	"go.yaml.in/yaml/v4"
)

// Logger - Custom lib logger
var Logger = log.New(io.Discard, "yamlutils ", log.LstdFlags)

// ErrInvalidParentType - The parent type is invalid.
var ErrInvalidParentType = fmt.Errorf("invalid parent type, must be list or key/value")

// ErrInvalidChildTypeKeyValue - The child type is invalid.
var ErrInvalidChildTypeKeyValue = fmt.Errorf("invalid child type, must be 'key: value'")

// YML object
type YML struct {
	Tree any
}

// NewFromFile returns a list of pointers to a YML object from a file.
//
// Returns a list since YAML files can contain multiple documents:
// https://yaml.org/spec/1.2-old/spec.html#id2800401
func NewFromFile(filename string) ([]*YML, error) {
	list := []*YML{}
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(fh)
	for {
		var tree any
		err = decoder.Decode(&tree)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, err
			}
			break
		}
		list = append(list, &YML{Tree: tree})
	}
	return list, nil
}

// NewFromReader returns a list of pointers to a YML object from an io.Reader.
//
// Returns a list since YAML files can contain multiple documents:
// https://yaml.org/spec/1.2-old/spec.html#id2800401
func NewFromReader(reader io.Reader) ([]*YML, error) {
	list := []*YML{}
	decoder := yaml.NewDecoder(reader)
	for {
		var tree any
		err := decoder.Decode(&tree)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, err
			}
			break
		}
		list = append(list, &YML{Tree: tree})
	}
	return list, nil
}

// NewFromString - returns a pointer to a YML object from a string.
func NewFromString(str string) (*YML, error) {
	var tree any
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
	target, _, errPath := trees.NavigateTree(include, y.Tree, keys)
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

func AddChild(m *any, child string) error {
	Logger.Printf("AddChild: %v += %v", *m, child)
	// Logger.Printf("type: %v\n", reflect.TypeOf(*m))
	var tree any
	err := yaml.Unmarshal([]byte(child), &tree)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%w", ErrInvalidParentType)
	}
	switch (*m).(type) {
	case map[any]any:
		Logger.Printf("AddChild: map type")
		Logger.Printf("value %v\n", tree)
		if t, ok := tree.(map[any]any); ok {
			for k, v := range t {
				// not importing maps.Copy for a single line
				(*m).(map[any]any)[k] = v
			}
			return nil
		}
		// maps.Copy can't copy map[string]any to a map[any]any
		if t, ok := tree.(map[string]any); ok {
			for k, v := range t {
				(*m).(map[any]any)[k] = v
			}
			return nil
		}
		return fmt.Errorf("%w: %T", ErrInvalidChildTypeKeyValue, tree)
	case []any:
		Logger.Printf("AddChild: slice/array type")
		r := append((*m).([]any), tree)
		*m = r
		return nil
	default:
		Logger.Printf("AddChild: single element type")
		return fmt.Errorf("%w", ErrInvalidParentType)
	}
}

func AddChildToTree(parent *any, current *any, p []string, child string) error {
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
	case map[any]any:
		Logger.Printf("AddChildToTree: map type")
		e, ok := t[p[0]]
		if !ok {
			return fmt.Errorf("%w: %s", trees.ErrMapKeyNotFound, p[0])
		}
		err := AddChildToTree(current, &e, p[1:], child)
		if err != nil {
			return err
		}
		(*current).(map[any]any)[p[0]] = e
		return nil
	case []any:
		Logger.Printf("AddChildToTree: slice/array type")
		index, err := strconv.Atoi(p[0])
		if err != nil {
			return fmt.Errorf("%w: %s", trees.ErrNotAnIndex, p[0])
		}
		if index < 0 || len(t) <= index {
			return fmt.Errorf("%w: %s", trees.ErrInvalidIndex, p[0])
		}
		err = AddChildToTree(current, &t[index], p[1:], child)
		if err != nil {
			return err
		}
		parent = current
		return nil
	default:
		Logger.Printf("AddChildToTree: single element type")
		return fmt.Errorf("%w: %s", trees.ErrExtraElementsInPath, strings.Join(p, "/"))
	}
}
