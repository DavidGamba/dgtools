// This file is part of watchmake.
//
// Copyright (C) 2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package lib provides a utility to watch the filesystem for changes and run tasks when that happens.

example:

version: 1

watch "file" {
	file      = "x"
	on_delete = "retrack"  // untrack
	on_error  = "warn"     // ignore, fail
}

watch "dir" {
	dir       = "y"
	on_delete = "retrack"  // untrack
	on_error  = "fail"     // ignore, fail
	ignore    = ["*.js", "vendor/"]
}

task "build" {
	files = [watch.file.files, watch.dir.files]
	task per_file {
		actions = [["echo", "modified", "{}"]]
	}
	task once {
		actions = [["go", "build"]]
	}
	task per_file {
		actions = [["echo", "completed", "{}"]]
	}
}

*/
package lib

import (
	"io"
	"io/ioutil"
	"log"

	"github.com/DavidGamba/dgtools/private/hclutils"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Logger - default logger instance
var Logger = log.New(ioutil.Discard, "", log.LstdFlags)

var schema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "task", LabelNames: []string{"id"}},
		{Type: "watch", LabelNames: []string{"id"}},
	},
}

const configVersion = 1

// Watchmake -
type Watchmake struct {
	Version  int        `hcl:"version"`
	Watchers []*Watcher `hcl:"watch,block"`
	Tasks    []*Task    `hcl:"task,block"`
}

// Watcher -
type Watcher struct {
	ID          string `hcl:"id,label"`
	Description string `hcl:"description,optional"`
	File        string `hcl:"file,optional"`
	Dir         string `hcl:"dir,optional"`
	Output      []string
}

// Task -
type Task struct {
	ID          string           `hcl:"id,label"`
	Description string           `hcl:"description,optional"`
	Files       []string         `hcl:"files"`
	Tasks       []*TaskExecution `hcl:"task,block"`
}

type TaskExecution struct {
	ID      string     `hcl:"id,label"`
	Actions [][]string `hcl:"actions"`
}

func DecodeConfig(w io.Writer, parser *hclparse.Parser, f *hcl.File) (*Watchmake, error) {
	config := &Watchmake{}

	content, diags := f.Body.Content(schema)
	err := hclutils.HandleDiags(w, parser, diags)
	if err != nil {
		return nil, err
	}

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}

	// Evaluate watch block on the first pass since task depends on it
	for _, block := range content.Blocks {
		switch block.Type {
		case "watch":
			watch := Watcher{}
			diags := gohcl.DecodeBody(block.Body, ctx, &watch)
			err = hclutils.HandleDiags(w, parser, diags)
			if err != nil {
				return config, err
			}
			watch.ID = block.Labels[0]

			// TODO: Evaluate file globs and generate list of files
			watch.Output = []string{watch.File}

			watcherCty, err := gocty.ToCtyValue(map[string][]string{"output": watch.Output}, cty.Object(map[string]cty.Type{
				"output": cty.List(cty.String),
			}))
			if err != nil {
				return config, err
			}

			var m map[string]cty.Value
			watchContextMap, ok := ctx.Variables["watch"]
			if !ok {
				m = map[string]cty.Value{
					block.Labels[0]: watcherCty,
				}
			} else {
				m = watchContextMap.AsValueMap()
				m[block.Labels[0]] = watcherCty
			}
			ctx.Variables["watch"] = cty.MapVal(m)

			Logger.Printf("Watch: %v\n", &watch)
			config.Watchers = append(config.Watchers, &watch)
		}
	}
	Logger.Printf("Ctx: %#v\n", ctx)

	for _, block := range content.Blocks {
		switch block.Type {
		case "task":
			task := Task{}
			diags := gohcl.DecodeBody(block.Body, ctx, &task)
			err = hclutils.HandleDiags(w, parser, diags)
			if err != nil {
				return config, err
			}
			task.ID = block.Labels[0]
			// Logger.Printf("Task: %s\n", &task)
			config.Tasks = append(config.Tasks, &task)
		}
	}
	return config, nil
}
