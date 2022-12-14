// This file is part of buildutils.
//
// Copyright (C) 2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package buildutils provides functions used when writing build automation.
*/
package buildutils

import (
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
	err := os.MkdirAll(dir, 0755)
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
