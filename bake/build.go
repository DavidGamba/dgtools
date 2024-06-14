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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/DavidGamba/dgtools/buildutils"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
)

func buildBinary(dir string) error {
	files, modified, err := fsmodtime.Target(os.DirFS(dir),
		[]string{"bake"},
		[]string{"*.go", "go.mod", "go.sum"})
	if err != nil {
		return err
	}
	if modified {
		Logger.Printf("Found modifications on %v, rebuilding binary...\n", files)
		_ = run.CMD("go", "get").Dir(dir).Log().Run()
		err = run.CMD("go", "build").Dir(dir).Log().Run()
		if err != nil {
			os.Remove(filepath.Join(dir, "bake"))
			return fmt.Errorf("failed to build binary: %w", err)
		}
	}
	return nil
}

var ErrNotFound = fmt.Errorf("not found")

// findBakeDir - searches for bakefiles/ dir first then for bake/ dir.
// This allows me to use bake in this repo.
func findBakeDir(ctx context.Context) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return ".", fmt.Errorf("failed to get working directory: %w", err)
	}
	Logger.Printf("Working directory: %s\n", wd)

	// First case, bake folder lives in CWD
	// This has higher priority to allow me to have a bake folder for bake itself.
	dir := filepath.Join(wd, "bakefiles")
	if fi, err := os.Stat(dir); err == nil && fi.Mode().IsDir() {
		return dir, nil
	}

	// Second case, we are withing the bake folder
	base := filepath.Base(wd)
	if base == "bakefiles" {
		return ".", nil
	}

	// Third case, search for bake folder in parent directories
	d, err := buildutils.FindDirUpwards(ctx, "bakefiles")
	if err == nil {
		return d, nil
	}
	if err != nil {
		if !errors.Is(err, buildutils.ErrNotFound) {
			return ".", fmt.Errorf("failed to find bake folder: %w", err)
		}
	}

	// First case, bake folder lives in CWD
	dir = filepath.Join(wd, "bake")
	if fi, err := os.Stat(dir); err == nil && fi.Mode().IsDir() {
		return dir, nil
	}

	// Second case, we are withing the bake folder
	if base == "bake" {
		return ".", nil
	}

	// Third case, search for bake folder in parent directories
	d, err = buildutils.FindDirUpwards(ctx, "bake")
	if err == nil {
		return d, nil
	}
	if err != nil {
		if !errors.Is(err, buildutils.ErrNotFound) {
			return ".", fmt.Errorf("failed to find bake folder: %w", err)
		}
	}
	return ".", ErrNotFound
}

func GenerateMainFile(ot *OptTree, dir string) error {
	files, modified, err := fsmodtime.Target(os.DirFS(dir),
		[]string{generatedMainFilename},
		[]string{"*.go", "go.mod", "go.sum"})
	if err != nil {
		return err
	}

	binaryExists := true
	if _, err := os.Stat(filepath.Join(dir, "bake")); os.IsNotExist(err) {
		binaryExists = false
	}

	if !modified && binaryExists {
		return nil
	}
	Logger.Printf("Found source modifications on %v, regenerating template...\n", files)

	// Render template
	tmpl, err := template.ParseFS(templates, "templates/main.go.gotmpl")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	data := map[string]string{
		"Tree": ot.String(),
	}
	// get writer to write to main.go
	w, err := os.Create(filepath.Join(dir, generatedMainFilename))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return run.CMD("go", "fmt", generatedMainFilename).Dir(dir).Log().Run()
}
