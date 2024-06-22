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
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func initRun(dir string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		err := initFn(opt)
		if err != nil {
			return fmt.Errorf("failed to inspect package: %w", err)
		}
		return nil
	}
}

func initFn(opt *getoptions.GetOpt) error {
	Logger.Printf("Initializing bake project in\n")
	dir := "bakefiles"
	os.MkdirAll(dir, 0755)

	_ = run.CMD("go", "mod", "init", "bake").Dir(dir).Log().Run()
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize go mod: %w", err)
	// }
	_ = run.CMD("go", "work", "init").Dir(dir).Log().Env("GOWORK=off").Run()
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize go work: %w", err)
	// }
	_ = run.CMD("go", "work", "use", ".").Dir(dir).Log().Run()
	// if err != nil {
	// 	return fmt.Errorf("failed to configure go work: %w", err)
	// }
	// github.com/DavidGamba/dgtools/buildutils
	// github.com/DavidGamba/dgtools/fsmodtime
	// github.com/DavidGamba/dgtools/run
	// github.com/DavidGamba/go-getoptions

	_ = run.CMD("go", "get", "-u", "github.com/DavidGamba/dgtools/buildutils").Dir(dir).Log().Run()
	_ = run.CMD("go", "get", "-u", "github.com/DavidGamba/dgtools/fsmodtime").Dir(dir).Log().Run()
	_ = run.CMD("go", "get", "-u", "github.com/DavidGamba/dgtools/run").Dir(dir).Log().Run()
	_ = run.CMD("go", "get", "-u", "github.com/DavidGamba/go-getoptions@command-name").Dir(dir).Log().Run()

	ot := NewOptTree(opt)
	err := GenerateMainFile(ot, dir)
	if err != nil {
		return fmt.Errorf("failed to generate file: %w", err)
	}

	return nil
}

func bakeDirHasRequiredFiles(dir string) bool {
	// check that files exist: go.mod, go.sum, go.work
	for _, file := range []string{"go.mod", "go.sum", "go.work"} {
		file = filepath.Join(dir, file)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			Logger.Printf("Missing file %s\n", file)
			return false
		}
	}

	return true
}
