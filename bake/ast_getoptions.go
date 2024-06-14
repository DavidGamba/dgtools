// This file is part of bake.
//
// Copyright (C) 2023-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"context"
	"go/ast"
	"go/printer"
	"iter"
	"regexp"
	"strings"

	"github.com/DavidGamba/go-getoptions"
)

// The goal is to be able to find the getoptions.CommandFn calls.
// Also, we need to inspect the function and get the opt.<Type> calls to know what options are being used.
//
//	func Asciidoc(opt *getoptions.GetOpt) getoptions.CommandFn {
//		opt.String("lang", "en", opt.ValidValues("en", "es"))
//		opt.String("hello", "world")
//		opt.String("hola", "mundo")
//		return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
func LoadAst(ctx context.Context, opt *getoptions.GetOpt, dir string) (*OptTree, error) {
	ot := NewOptTree(opt)

	for getOptFn, err := range AstGetoptionFns(ctx, dir) {
		if err != nil {
			return ot, err
		}

		cmd, err := ot.AddCommand(getOptFn.Name, getOptFn.DescName, getOptFn.Description)
		if err != nil {
			return ot, err
		}
		err = addOptionsToCMD(getOptFn, cmd, getOptFn.DescName)
		if err != nil {
			return ot, err
		}
	}

	return ot, nil
}

type GetOptFn struct {
	FnDecl

	DescName     string
	OptFieldName string
}

// The goal is to be able to find the getoptions.CommandFn calls.
// Also, we need to inspect the function and get the opt.<Type> calls to know what options are being used.
//
//	func Asciidoc(opt *getoptions.GetOpt) getoptions.CommandFn {
//		opt.String("lang", "en", opt.ValidValues("en", "es"))
//		opt.String("hello", "world")
//		opt.String("hola", "mundo")
//		return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
func AstGetoptionFns(ctx context.Context, dir string) iter.Seq2[GetOptFn, error] {
	return func(yield func(GetOptFn, error) bool) {
		// Regex for description: fn-name - description
		re := regexp.MustCompile(`^\w\S+ -`)

	LOOP:
		for fnDecl, err := range AstFns(dir) {
			if err != nil {
				yield(GetOptFn{}, err)
				return
			}
			getOptFn := GetOptFn{FnDecl: fnDecl}

			x := fnDecl.Node.(*ast.FuncDecl)

			getOptFn.Description = strings.TrimSpace(getOptFn.Description)

			// Expect function of type:
			// func Name(opt *getoptions.GetOpt) getoptions.CommandFn

			// Check Params
			// Expect opt *getoptions.GetOpt
			if len(x.Type.Params.List) != 1 {
				continue
			}
			for _, param := range x.Type.Params.List {
				name := param.Names[0].Name
				var buf bytes.Buffer
				printer.Fprint(&buf, fnDecl.ParsedFile.fset, param.Type)
				// Logger.Printf("name: %s, %s\n", name, buf.String())
				if buf.String() != "*getoptions.GetOpt" {
					continue LOOP
				}
				getOptFn.OptFieldName = name
			}

			// Check Results
			// Expect getoptions.CommandFn
			if len(x.Type.Results.List) != 1 {
				continue
			}
			for _, result := range x.Type.Results.List {
				var buf bytes.Buffer
				printer.Fprint(&buf, fnDecl.ParsedFile.fset, result.Type)
				// Logger.Printf("result: %s\n", buf.String())
				if buf.String() != "getoptions.CommandFn" {
					continue LOOP
				}
			}

			// TODO: The yield probably goes here
			// Add function to OptTree
			if getOptFn.Description != "" {
				// Logger.Printf("description '%s'\n", description)
				if re.MatchString(getOptFn.Description) {
					// Get first word from string
					getOptFn.DescName = strings.Split(getOptFn.Description, " ")[0]
					getOptFn.Description = strings.TrimPrefix(getOptFn.Description, getOptFn.DescName+" -")
					getOptFn.Description = strings.TrimSpace(getOptFn.Description)
				}
			} else {
				getOptFn.DescName = camelToKebab(getOptFn.Name)
			}
			if !yield(getOptFn, nil) {
				return
			}
		}
	}
}
