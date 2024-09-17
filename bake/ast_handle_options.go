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
	"go/printer"
	"os"
	"strconv"

	"github.com/DavidGamba/go-getoptions"
)

func addOptionsToCMD(getOptFn GetOptFn, cmd *getoptions.GetOpt, name string) error {
	Logger.Printf("Adding options to %s\n", name)

	var outerErr error
	// Check for Expressions of opt type
	ast.Inspect(getOptFn.Node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BlockStmt:
			for _, stmt := range x.List {
				var buf bytes.Buffer
				printer.Fprint(&buf, getOptFn.ParsedFile.fset, stmt)
				// We are expecting the expression before the return function
				_, ok := stmt.(*ast.ReturnStmt)
				if ok {
					return false
				}
				// Logger.Printf("stmt: %s\n", buf.String())
				exprStmt, ok := stmt.(*ast.ExprStmt)
				if !ok {
					continue
				}
				// spew.Dump(exprStmt)

				// Check for CallExpr
				ast.Inspect(exprStmt, func(n ast.Node) bool {
					switch x := n.(type) {
					case *ast.CallExpr:
						fun, ok := x.Fun.(*ast.SelectorExpr)
						if !ok {
							return false
						}
						xIdent, ok := fun.X.(*ast.Ident)
						if !ok {
							return false
						}
						if xIdent.Name != getOptFn.OptFieldName {
							return false
						}
						// Logger.Printf("handling %s.%s\n", xIdent.Name, fun.Sel.Name)

						switch fun.Sel.Name {
						case "Bool", "String", "StringOptional", "Int", "IntOptional", "Increment", "Float64", "Float64Optional":
							mfns := []getoptions.ModifyFn{}
							if len(x.Args) > 2 {
								mfns = handleOptionModifiers(cmd, getOptFn.OptFieldName, x.Args[2:])
							}
							name, defaultValue, err := extractNameAndDefault(n, 0)
							if err != nil {
								outerErr = err
								return false
							}
							optionTypeSwitch(cmd, fun.Sel.Name, name, defaultValue, mfns)

						case "BoolVar", "StringVar", "StringVarOptional", "IntVar", "IntVarOptional", "IncrementVar", "Float64Var", "Float64VarOptional":
							mfns := []getoptions.ModifyFn{}
							if len(x.Args) > 3 {
								mfns = handleOptionModifiers(cmd, getOptFn.OptFieldName, x.Args[3:])
							}
							name, defaultValue, err := extractNameAndDefault(n, 1)
							if err != nil {
								outerErr = err
								return false
							}
							optionTypeSwitch(cmd, fun.Sel.Name, name, defaultValue, mfns)
						case "StringSliceVar", "StringMapVar", "IntSliceVar", "Float64SliceVar":
							mfns := []getoptions.ModifyFn{}
							if len(x.Args) > 4 {
								mfns = handleOptionModifiers(cmd, getOptFn.OptFieldName, x.Args[4:])
							}
							name, defaultValue, err := extractNameAndDefault(n, 1)
							if err != nil {
								outerErr = err
								return false
							}
							optionTypeSwitch(cmd, fun.Sel.Name, name, defaultValue, mfns)
						}

						return false
					}
					return true
				})
			}
		}
		return true
	})
	return outerErr
}

func optionTypeSwitch(cmd *getoptions.GetOpt, identifierName, name, defaultValue string, mfns []getoptions.ModifyFn) {
	switch identifierName {
	case "Bool", "BoolVar":
		d := false
		if defaultValue == "true" {
			d = true
		}
		cmd.Bool(name, d, mfns...)
	case "String", "StringVar":
		cmd.String(name, defaultValue, mfns...)
	case "StringOptional", "StringVarOptional":
		cmd.StringOptional(name, defaultValue, mfns...)
	case "Int", "IntVar":
		x, err := strconv.Atoi(defaultValue)
		if err != nil {
			x = 0
		}
		cmd.Int(name, x, mfns...)
	case "IntOptional", "IntVarOptional":
		x, err := strconv.Atoi(defaultValue)
		if err != nil {
			x = 0
		}
		cmd.IntOptional(name, x, mfns...)
	case "Increment", "IncrementVar":
		x, err := strconv.Atoi(defaultValue)
		if err != nil {
			x = 0
		}
		cmd.Increment(name, x, mfns...)
	case "Float64", "Float64Var":
		x, err := strconv.ParseFloat(defaultValue, 64)
		if err != nil {
			x = 0.0
		}
		cmd.Float64(name, x, mfns...)
	case "Float64Optional", "Float64VarOptional":
		x, err := strconv.ParseFloat(defaultValue, 64)
		if err != nil {
			x = 0.0
		}
		cmd.Float64Optional(name, x, mfns...)
	case "StringSliceVar":
		cmd.StringSlice(name, 1, 99, mfns...)
	case "IntSliceVar":
		cmd.IntSlice(name, 1, 99, mfns...)
	case "Float64SliceVar":
		cmd.Float64Slice(name, 1, 99, mfns...)
	}
}

func extractNameAndDefault(n ast.Node, offset int) (string, string, error) {
	x := n.(*ast.CallExpr)
	name, err := extractName(x.Args, offset)
	if err != nil {
		return "", "", err
	}
	defaultValue, err := extractDefault(x.Args, offset)
	if err != nil {
		return "", "", err
	}
	return name, defaultValue, nil
}

func extractName(args []ast.Expr, offset int) (string, error) {
	// First argument is the Name
	if len(args) < 1+offset {
		return "", fmt.Errorf("missing name argument")
	}
	name, err := strconv.Unquote(args[0+offset].(*ast.BasicLit).Value)
	if err != nil {
		name = args[0+offset].(*ast.BasicLit).Value
	}
	return name, nil
}

func extractDefault(args []ast.Expr, offset int) (string, error) {
	// Second argument is the Default
	if len(args) < 2+offset {
		return "", fmt.Errorf("missing default argument")
	}
	defaultValue := ""
	var err error
	switch args[1+offset].(type) {
	case *ast.BasicLit:
		defaultValue, err = strconv.Unquote(args[1+offset].(*ast.BasicLit).Value)
		if err != nil {
			defaultValue = args[1+offset].(*ast.BasicLit).Value
		}
	case *ast.Ident:
		defaultValue = args[1+offset].(*ast.Ident).String()
	}
	return defaultValue, nil
}

func handleOptionModifiers(cmd *getoptions.GetOpt, optFieldName string, args []ast.Expr) []getoptions.ModifyFn {
	mfns := []getoptions.ModifyFn{}
	for _, arg := range args {
		callE, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}
		fun, ok := callE.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		xIdent, ok := fun.X.(*ast.Ident)
		if !ok {
			continue
		}
		if xIdent.Name != optFieldName {
			continue
		}
		// Logger.Printf("\t%s.%s\n", xIdent.Name, fun.Sel.Name)
		if fun.Sel.Name == "SetCalled" {
			// TODO: SetCalled function receives a bool
			fmt.Fprintf(os.Stderr, "WARNING: bake: SetCalled is not implemented\n")
			continue
		}
		values := []string{}
		for _, arg := range callE.Args {
			// Logger.Printf("Value: %s\n", arg.(*ast.BasicLit).Value)
			value, err := strconv.Unquote(arg.(*ast.BasicLit).Value)
			if err != nil {
				value = arg.(*ast.BasicLit).Value
			}
			values = append(values, value)
		}
		switch fun.Sel.Name {
		case "Alias":
			mfns = append(mfns, cmd.Alias(values...))
		case "ArgName":
			if len(values) > 0 {
				mfns = append(mfns, cmd.ArgName(values[0]))
			}
		case "Description":
			if len(values) > 0 {
				mfns = append(mfns, cmd.Description(values[0]))
			}
		case "GetEnv":
			if len(values) > 0 {
				mfns = append(mfns, cmd.GetEnv(values[0]))
			}
		case "Required":
			mfns = append(mfns, cmd.Required(values...))
		case "SuggestedValues":
			mfns = append(mfns, cmd.SuggestedValues(values...))
		case "ValidValues":
			mfns = append(mfns, cmd.ValidValues(values...))
		}
	}
	return mfns
}
