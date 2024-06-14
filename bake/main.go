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
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/DavidGamba/dgtools/buildutils"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

//go:embed templates/*.go.gotmpl
var templates embed.FS

const (
	version               = "0.1.0"
	generatedMainFilename = "generated_bake.go"
)

var InputArgs []string
var Dir string

var Logger = log.New(io.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	run.Logger = Logger

	if os.Getenv("BAKE_TRACE") != "" {
		Logger.SetOutput(os.Stderr)
		run.Logger.SetOutput(os.Stderr)
		buildutils.Logger.SetOutput(os.Stderr)
	}

	opt := getoptions.New()
	opt.Self("bake", "Go Build + Something like Make = Bake ¯\\_(ツ)_/¯")
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))

	dir, err := findBakeDir(ctx)
	if err != nil && !errors.Is(err, ErrNotFound) {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	Dir = dir
	InputArgs = args[1:]
	Logger.Printf("Running bake in %s with args: %v\n", dir, InputArgs)

	requiredFilesPresent := false
	if err == nil {
		requiredFilesPresent = bakeDirHasRequiredFiles(dir)
	}

	if requiredFilesPresent {
		ot, err := LoadAst(ctx, opt, dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return 1
		}
		err = GenerateMainFile(ot, dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return 1
		}
		err = buildBinary(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return 1
		}
	} else {
		Logger.Printf("Required files not present\n")
	}

	b := opt.NewCommand("_bake", "")

	bld := b.NewCommand("list-fns", "lists all functions in the package")
	bld.SetCommandFn(PrintFuncDeclRun(dir))

	binit := b.NewCommand("init", "initialize a new bake project")
	binit.SetCommandFn(initRun(dir))

	bforce := b.NewCommand("force", "force rebuild of the generated bake file and the binary on the next run")
	bforce.SetCommandFn(InvalidateCache(dir))

	bversion := b.NewCommand("version", "print the version of bake")
	bversion.SetCommandFn(Version(dir))

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func PrintFuncDeclRun(dir string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		err := PrintFuncDecl(dir)
		if err != nil {
			return fmt.Errorf("failed to inspect package: %w", err)
		}
		return nil
	}
}

func InvalidateCache(dir string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		fmt.Printf("Invalidating bake cache...\n")
		err := buildutils.Touch(filepath.Join(dir, "go.mod"))
		if err != nil {
			return fmt.Errorf("failed to invalidate cache: %w", err)
		}
		return nil
	}
}

func Version(dir string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		fmt.Printf("bake version %s\n", version)
		fmt.Printf("go version %s\n", runtime.Version())
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return fmt.Errorf("failed to read build info")
		}
		fmt.Printf("%14s %s\n", "module.path", info.Main.Path)
		for _, s := range info.Settings {
			if s.Value != "" {
				fmt.Printf("%14s %s\n", s.Key, s.Value)
			}
		}
		return nil
	}
}
