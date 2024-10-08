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
	"path/filepath"
	"strings"
	"time"
)

// ExpandEnv - like os.ExpandEnv on many lines, except that it reports error if any of the env vars is
// not found. It also expands ~/ to $HOME/
//
// replacements is a list of additional replacements and can be nil.
func ExpandEnv(lines []string, replacements map[string]string) ([]string, error) {
	expanded := []string{}
	var err error
	notFound := make(map[string]struct{})
	mappingFn := func(name string) string {
		if replacements != nil {
			if value, ok := replacements[name]; ok {
				return value
			}
		}
		if value, ok := os.LookupEnv(name); ok {
			return value
		}
		notFound[name] = struct{}{}
		return ""
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "~/") {
			line = strings.Replace(line, "~/", "$HOME/", 1)
		}
		line = strings.ReplaceAll(line, " ~/", " $HOME/")
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
func Glob(fsys fs.FS, stop bool, patterns []string) (matches []string, stopped bool, err error) {
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
//
// Use fsmodtime.Recursive(true) to recurse into directories.
func Target(fsys fs.FS, targets []string, sources []string, opts ...WalkOpt) ([]string, bool, error) {
	targets, err := ExpandEnv(targets, nil)
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
	p, fi, err := Last(fsys, targets, opts...)
	if err != nil {
		return nil, false, err
	}
	Logger.Printf("last target: %q\n", path.Join(p, fi.Name()))
	return TargetTime(fsys, fi.ModTime(), sources, opts...)
}

// TargetTime - Given a time it indicates whether or not the sources have modifications past the time.
// The first return is the file modified.
//
// Use fsmodtime.Recursive(true) to recurse into directories.
func TargetTime(fsys fs.FS, targetTime time.Time, sources []string, opts ...WalkOpt) ([]string, bool, error) {
	sources, err := ExpandEnv(sources, nil)
	if err != nil {
		return nil, false, err
	}
	sources, _, err = Glob(fsys, false, sources)
	if err != nil {
		return nil, false, err
	}
	Logger.Printf("sources: %q\n", sources)
	// TODO: This is the wrong implementation, I don't need the times of all files, only one newer than targetTime
	p, fi, err := Last(fsys, sources, opts...)
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

// ParentDir - Given a list of paths, it finds the common parent dir.
// Assumes / as the filepath separator.
func ParentDir(paths []string) (string, error) {
	switch len(paths) {
	case 0:
		return "", nil
	case 1:
		return filepath.Clean(paths[0]), nil
	}
	p, err := filepath.Abs(paths[0])
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	pl := strings.Split(p, "/")
	for _, v := range paths[1:] {
		v, err := filepath.Abs(v)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}
		vl := strings.Split(v, "/")
		if len(vl) < len(pl) {
			pl = pl[:len(vl)]
		}
		for i := 0; i < len(pl); i++ {
			if vl[i] != pl[i] {
				pl = pl[:i]
			}
		}
	}
	r := "/" + filepath.Join(pl...)
	return r, nil
}
