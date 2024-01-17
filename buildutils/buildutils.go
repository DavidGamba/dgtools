// This file is part of buildutils.
//
// Copyright (C) 2021-2023  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package buildutils provides functions used when writing build automation.
*/
package buildutils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/dgtools/run"
)

// GitRepoRoot - Gets the Git repository root directory
func GitRepoRoot() (string, error) {
	out, err := run.CMD("git", "rev-parse", "--show-toplevel").STDOutOutput()
	return strings.TrimSpace(string(out)), err
}

// CDGitRepoRoot - Chdir to the Git repository root directory
func CDGitRepoRoot() error {
	out, err := GitRepoRoot()
	if err != nil {
		return err
	}
	err = os.Chdir(out)
	if err != nil {
		return err
	}
	return nil
}

// GitRepoName - Gets the Git repository name by parsing the origin URL
func GitRepoName() (string, error) {
	out, err := run.CMD("git", "config", "--get", "remote.origin.url").STDOutOutput()
	url := strings.TrimSpace(string(out))
	name := strings.TrimSuffix(filepath.Base(url), ".git")
	return name, err
}

func GetFileFromURL(url, outputFilename string) error {
	dir := filepath.Dir(outputFilename)
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return fmt.Errorf("failed to create dir structure '%s': %s", dir, err)
	}

	f, err := os.Create(outputFilename)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %s", outputFilename, err)
	}
	defer f.Close()

	client := http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from '%s': %s", url, err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file '%s': %s", outputFilename, err)
	}

	return nil
}

// GoModDir - Gets the Go module directory, the root of the Go project.
func GoModDir() (string, error) {
	out, err := run.CMD("go", "list", "-m", "-f", "{{.Dir}}").STDOutOutput()
	return strings.TrimSpace(string(out)), err
}

var ErrNotFound = fmt.Errorf("not found")

// FindFileUpwards - traverses the file system upwards looking for a file
// Returns buildutils.ErrNotFound if the file is not found
func FindFileUpwards(ctx context.Context, filename string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get cwd: %w", err)
	}
	check := func(dir string) bool {
		f := filepath.Join(dir, filename)
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return false
		}
		return true
	}
	d := cwd
	for {
		found := check(d)
		if found {
			return filepath.Join(d, filename), nil
		}
		a, err := filepath.Abs(d)
		if err != nil {
			return "", fmt.Errorf("failed to get abs path: %w", err)
		}
		if a == "/" {
			break
		}
		d = filepath.Join(d, "../")
	}

	return "", fmt.Errorf("%w: %s", ErrNotFound, filename)
}
