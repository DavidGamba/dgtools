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
	"os"
	"strings"
	"time"

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

// TODO: ALlow for fs override for testing

func SymlinkF(target, name string) error {
	_ = os.Remove(name)
	return os.Symlink(target, name)
}

func Touch(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(filename, currentTime, currentTime)
		if err != nil {
			return err
		}
	}
	return nil
}
