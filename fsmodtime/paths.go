// This file is part of fsmodtime.
//
// Copyright (C) 2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package fsmodtime

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"
)

var ErrNotFound = fmt.Errorf("not found")

// ExpandEnv - like os.ExpandEnv on many lines, except that it reports error if any of the env vars is
// not found.
func ExpandEnv(lines []string) ([]string, error) {
	expanded := []string{}
	var err error
	notFound := make(map[string]struct{})
	mappingFn := func(name string) string {
		if value, ok := os.LookupEnv(name); ok {
			return value
		}
		notFound[name] = struct{}{}
		return ""
	}
	for _, line := range lines {
		expanded = append(expanded, os.Expand(line, mappingFn))
	}
	if len(notFound) > 0 {
		notFoundList := []string{}
		for k := range notFound {
			notFoundList = append(notFoundList, k)
		}
		err = fmt.Errorf("%w: %q", ErrNotFound, notFoundList)
	}
	return expanded, err
}

// Glob - like io/fs.Glob on many paths.
// It allows to stop globbing patterns once it has found a pattern that has no matches.
// Glob syntax described here: https://golang.org/pkg/path/filepath/#Match
//
// Returns the list of matches, a bool indicating if there is a pattern that had no matches and an error
func Glob(fsys fs.FS, stop bool, patterns []string) ([]string, bool, error) {
	matches := []string{}
	for _, p := range patterns {
		m, err := fs.Glob(fsys, p)
		if err != nil {
			return matches, false, err
		}
		if stop && len(m) == 0 {
			return matches, true, nil
		}
		matches = append(matches, m...)
	}
	return matches, false, nil
}

// Target - Given a list of targets it indicates whether or not the sources have modifications past the targets last.
// The first return is the file modified.
func Target(fsys fs.FS, targets []string, sources []string) ([]string, bool, error) {
	targets, err := ExpandEnv(targets)
	if err != nil {
		return nil, false, err
	}
	targets, stopped, err := Glob(fsys, true, targets)
	if err != nil {
		return nil, false, err
	}
	if stopped {
		// targets don't exist, re-create them
		return []string{}, true, nil
	}
	Logger.Printf("targets: %q\n", targets)
	p, fi, err := Last(fsys, targets)
	if err != nil {
		return nil, false, err
	}
	Logger.Printf("last target: %q\n", path.Join(p, fi.Name()))
	return TargetTime(fsys, fi.ModTime(), sources)
}

// TargetTime - Given a time it indicates whether or not the sources have modifications past the time.
// The first return is the file modified.
func TargetTime(fsys fs.FS, targetTime time.Time, sources []string) ([]string, bool, error) {
	sources, err := ExpandEnv(sources)
	if err != nil {
		return nil, false, err
	}
	sources, _, err = Glob(fsys, false, sources)
	if err != nil {
		return nil, false, err
	}
	Logger.Printf("sources: %q\n", sources)
	// TODO: This is the wrong implementation, I don't need the times of all files, only one newer than targetTime
	p, fi, err := Last(fsys, sources)
	if err != nil {
		return nil, false, err
	}
	if fi == nil {
		return nil, false, ErrNotFound
	}
	Logger.Printf("last source: %q\n", path.Join(p, fi.Name()))
	if targetTime.Before(fi.ModTime()) {
		// targets are older, re-create them
		return []string{path.Join(p, fi.Name())}, true, nil
	}
	// targets are newer, nothing to do
	return []string{}, false, nil
}
