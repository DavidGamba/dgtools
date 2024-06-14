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
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"iter"
	"strings"

	"golang.org/x/tools/go/packages"
)

type ParsedFile struct {
	file string
	fset *token.FileSet
	f    *ast.File
}

// Requires GOEXPERIMENT=rangefunc
func parsedFiles(dir string) iter.Seq2[ParsedFile, error] {
	return func(yield func(ParsedFile, error) bool) {
		cfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax, Dir: dir}
		pkgs, err := packages.Load(cfg, ".")
		if err != nil {
			yield(ParsedFile{}, fmt.Errorf("failed to load packages: %w", err))
			return
		}
		for _, pkg := range pkgs {
			// Logger.Println(pkg.ID, pkg.GoFiles)
			for _, file := range pkg.GoFiles {
				if strings.Contains(file, "generated") {
					continue
				}
				p := ParsedFile{}
				// Logger.Printf("file: %s\n", file)
				// parse file
				fset := token.NewFileSet()
				fset.AddFile(file, fset.Base(), len(file))
				p.file = file
				p.fset = fset
				f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
				if err != nil {
					yield(p, fmt.Errorf("failed to parse file: %w", err))
					return
				}
				p.f = f
				if !yield(p, nil) {
					return
				}
			}
		}
	}
}

type FnDecl struct {
	Name        string // function name
	Description string

	Node       ast.Node
	ParsedFile ParsedFile
	Type       string
}

// Requires GOEXPERIMENT=rangefunc
func AstFns(dir string) iter.Seq2[FnDecl, error] {
	return func(yield func(FnDecl, error) bool) {
		for p, err := range parsedFiles(dir) {
			if err != nil {
				yield(FnDecl{}, err)
				return
			}
			fnDecl := FnDecl{ParsedFile: p}

			doneYield := false
			// Iterate through every node in the file
			ast.Inspect(p.f, func(n ast.Node) bool {
				if doneYield {
					return false
				}
				fnDecl.Node = n
				switch x := n.(type) {
				// Check function declarations for exported functions
				case *ast.FuncDecl:
					if x.Name.IsExported() {
						fnDecl.Name = x.Name.Name
						fnDecl.Description = x.Doc.Text()
						var buf bytes.Buffer
						printer.Fprint(&buf, p.fset, x.Type)
						fnDecl.Type = buf.String()
						if !yield(fnDecl, nil) {
							doneYield = true
							return false
						}
					}
				}
				return true
			})

			if doneYield {
				return
			}
		}
	}
}

func PrintFuncDecl(dir string) error {
	for fnDecl, err := range AstFns(dir) {
		if err != nil {
			return err
		}
		fmt.Printf("file: %s\n\tname: %s\n\ttype: %s\n\tdesc: %s\n", fnDecl.ParsedFile.file, fnDecl.Name, fnDecl.Type, fnDecl.Description)
	}

	return nil
}
