// This file is part of bake.
//
// Copyright (C) 2023-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

// type TaskDefinitionFn func(ctx context.Context, opt *getoptions.GetOpt) error
// type TaskFn func(*getoptions.GetOpt) getoptions.CommandFn

type OptTree struct {
	Root *OptNode
}

type OptNode struct {
	Name        string
	Opt         *getoptions.GetOpt
	Children    map[string]*OptNode
	Parent      string
	DescName    string
	Description string
}

func NewOptTree(opt *getoptions.GetOpt) *OptTree {
	return &OptTree{
		Root: &OptNode{
			Name:        "",
			Parent:      "",
			Opt:         opt,
			DescName:    "",
			Description: "",
			Children:    make(map[string]*OptNode),
		},
	}
}

// Regex for description: fn-name - description
var descriptionRe = regexp.MustCompile(`^\w\S+ -`)

func (ot *OptTree) AddCommand(name, descName, description string) (*getoptions.GetOpt, error) {
	Logger.Printf("Adding command %s with function %s\n", descName, name)
	keys := strings.Split(descName, ":")
	node := ot.Root
	var cmd *getoptions.GetOpt
	for i, key := range keys {
		n, ok := node.Children[key]
		if ok {
			// Logger.Printf("key: %v already defined, parent: %s\n", key, node.Name)
			node = n
			cmd = n.Opt
			if len(keys) == i+1 {
				cmd.Self(key, description)
			}
			continue
		}
		// Logger.Printf("key: %v not defined, parent: %s\n", key, node.Name)
		desc := ""
		if len(keys) == i+1 {
			desc = description
		}
		err := validateCmdName(key, descName)
		if err != nil {
			return nil, err
		}
		cmd = node.Opt.NewCommand(key, desc)
		node.Children[key] = &OptNode{
			Name:        "",
			Parent:      node.DescName,
			Opt:         cmd,
			Children:    make(map[string]*OptNode),
			Description: desc,
			DescName:    key,
		}
		if len(keys) == i+1 {
			node.Children[key].Name = name
			cmd.SetCommandFn(func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
				Logger.Printf("Running %v from %s\n", InputArgs, Dir)
				// filepath.Join removes the ./ if Dir is .
				// Need to ensure that it is running the local binary, not the one in the PATH
				cmd := "./bake"
				if Dir != "." {
					cmd = filepath.Join(Dir, "bake")
				}
				c := []string{cmd}
				run.CMD(append(c, InputArgs...)...).Log().Run()
				return nil
			})
		}
		node = node.Children[key]
	}
	return cmd, nil
}

var golangKeywords = map[string]struct{}{
	"break":       {},
	"default":     {},
	"func":        {},
	"interface":   {},
	"go":          {},
	"select":      {},
	"case":        {},
	"defer":       {},
	"goto":        {},
	"map":         {},
	"struct":      {},
	"chan":        {},
	"else":        {},
	"if":          {},
	"package":     {},
	"switch":      {},
	"const":       {},
	"fallthrough": {},
	"import":      {},
	"range":       {},
	"type":        {},
	"continue":    {},
	"for":         {},
	"return":      {},
	"var":         {},
}

func validateCmdName(name, descName string) error {
	// if command name matches a golang keyword, return an error
	if _, ok := golangKeywords[name]; ok {
		return fmt.Errorf("command name '%s' in '%s' is a golang keyword", name, descName)
	}
	return nil
}

func (ot *OptTree) String() string {
	return ot.Root.String()
}

func (on *OptNode) String() string {
	out := ""
	parent := on.Parent
	if parent == "" {
		parent = "opt"
	}

	if on.DescName != "" {
		out += fmt.Sprintf("%s := %s.NewCommand(\"%s\", `%s`)\n", on.DescName, parent, on.DescName, on.Description)
	}

	if on.Name != "" {
		out += fmt.Sprintf("%sFn := %s(%s)\n", on.DescName, on.Name, on.DescName)
		out += fmt.Sprintf("%s.SetCommandFn(%sFn)\n", on.DescName, on.DescName)

		// TODO: This is not considering more than two levels of commands
		if parent == "opt" {
			out += fmt.Sprintf("TM.Add(\"%s\", %sFn)\n\n", on.DescName, on.DescName)
		} else {
			out += fmt.Sprintf("TM.Add(\"%s:%s\", %sFn)\n\n", parent, on.DescName, on.DescName)
		}
	}
	for _, child := range on.Children {
		out += child.String()
	}
	return out
}
