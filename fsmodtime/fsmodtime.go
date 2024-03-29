// This file is part of fsmodtime.
//
// Copyright (C) 2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// package fsmodtime provides functions to compare fs mod times.
//
// The goal of this package is to have functions that allow me to determine if
// my sources (fs dependencies) have been modified (changed) so that based on
// that I can rebuild my targets.
//
// Additionally I want to log the file (not necessarily all of them but at least one) that
// changed for informational/verification purposes when building build systems downstream.
//
// The function that meets that goal is [Target].
//
// Requirements:
//
//   - Targets might not exist yet.
//     In that case, build them.
//   - All declared targets must exist.
//     Rebuild otherwise.
//   - Not all sources are required to exist.
//     For example, I might want a blank *.jpg and *.png in my build system but some image types might not exist.
//   - ExpandEnv but don't fail silently if my Env Var expansions fail.
//   - Allow for globs.
//
// Example:
//
//	targets := []string{"$outputs_dir/doc.pdf", "$outputs_dir/*.html"}
//	sources := []string{"$src_dir/*.adoc", "$images_dir/*.jpg", "$images_dir/*.png"}
//	paths, modified, err := Target(os.DirFS("."), targets, sources)
package fsmodtime

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"path"
	"path/filepath"
)

var Logger = log.New(io.Discard, "", log.LstdFlags)

var (
	ErrInvalidPath = fmt.Errorf("invalid path")
	ErrInvalidFS   = fmt.Errorf("invalid fs")
	ErrNotFound    = fmt.Errorf("not found")
)

type WalkOpts struct {
	recursive bool
	// followSymlinks bool
}

type WalkOpt func(*WalkOpts)

func Recursive(enabled bool) WalkOpt {
	return func(opts *WalkOpts) {
		opts.recursive = enabled
	}
}

// TODO: Add option to follow symlinks.
// func FollowSymlinks(enabled bool) WalkOpt {
// 	return func(opts *WalkOpts) {
// 		opts.followSymlinks = enabled
// 	}
// }

// Last - given a list of paths, it finds the file with the latest modTime and returns it.
//
//	root := "."
//	fileSystem := os.DirFS(root)
//	path, fi, err := fsmodtime.Last(fileSystem, paths, fsmodtime.Recursive(true))
//
// Use fsmodtime.Recursive(true) to recurse into directories.
func Last(fsys fs.FS, paths []string, opts ...WalkOpt) (filepath string, fileInfo fs.FileInfo, err error) {
	wo := &WalkOpts{}
	for _, opt := range opts {
		opt(wo)
	}

	afterFn := func(root string, fi fs.FileInfo) error {
		Logger.Printf("fn: %s\n", path.Join(root, fi.Name()))
		if fileInfo == nil {
			fileInfo = fi
			filepath = root
			return nil
		}
		if fi.ModTime().After(fileInfo.ModTime()) {
			fileInfo = fi
			filepath = root
		}
		return nil
	}

	err = walkPaths(fsys, paths, wo, afterFn)
	if err != nil {
		return "", nil, err
	}

	if fileInfo == nil {
		return "", nil, fmt.Errorf("%w: %q", ErrNotFound, paths)
	}

	return filepath, fileInfo, nil
}

// First - given a list of paths, it finds the file with the earliest modTime and returns it.
//
//	root := "."
//	fileSystem := os.DirFS(root)
//	path, fi, err := fsmodtime.First(fileSystem, paths, fsmodtime.Recursive(true))
//
// Use fsmodtime.Recursive(true) to recurse into directories.
func First(fsys fs.FS, paths []string, opts ...WalkOpt) (filepath string, fileInfo fs.FileInfo, err error) {
	wo := &WalkOpts{}
	for _, opt := range opts {
		opt(wo)
	}

	beforeFn := func(root string, fi fs.FileInfo) error {
		Logger.Printf("fn: %s\n", path.Join(root, fi.Name()))
		if fileInfo == nil {
			fileInfo = fi
			filepath = root
			return nil
		}
		if fi.ModTime().Before(fileInfo.ModTime()) {
			fileInfo = fi
			filepath = root
		}
		return nil
	}

	err = walkPaths(fsys, paths, wo, beforeFn)
	if err != nil {
		return "", nil, err
	}

	if fileInfo == nil {
		return "", nil, fmt.Errorf("%w: %q", ErrNotFound, paths)
	}

	return filepath, fileInfo, nil
}

func walkPaths(fsys fs.FS, paths []string, wo *WalkOpts, fn fileInfoFn) error {
	if fsys == nil {
		return ErrInvalidFS
	}

	for _, path := range paths {
		Logger.Printf("path: %s\n", path)

		// validate path
		if path == "" {
			return fmt.Errorf("%w: '%s'", ErrInvalidPath, path)
		}
		fi, err := fs.Stat(fsys, path)
		if err != nil {
			return err
		}

		err = fileInfoIterate(fsys, filepath.Dir(path), fs.FileInfoToDirEntry(fi), fn, 1, wo)
		if err != nil {
			return err
		}
	}
	return nil
}

type fileInfoFn func(root string, fi fs.FileInfo) error

// Given a fs.DirEntry (or a fs.FileInfo using `fs.FileInfoToDirEntry(fi)`) it
// expands every dir and runs fn on every resulting child fs.DirEntry.
//
// NOTE: It doesn't run fn on dirs.
//
// TODO: It doesn't follow symlinks
func fileInfoIterate(fsys fs.FS, root string, de fs.DirEntry, fn fileInfoFn, depth int, wo *WalkOpts) error {
	if de.IsDir() {
		Logger.Printf("depth: %d\n", depth)
		if !wo.recursive {
			return nil
		}
		dir := path.Join(root, de.Name())
		Logger.Printf("expand: %s\n", dir)
		dirEntries, err := fs.ReadDir(fsys, dir)
		if err != nil {
			return err
		}
		for _, de := range dirEntries {
			err := fileInfoIterate(fsys, dir, de, fn, depth+1, wo)
			if err != nil {
				return err
			}
		}
		return nil
	}
	fi, err := de.Info()
	if err != nil {
		return err
	}
	err = fn(root, fi)
	if err != nil {
		return err
	}
	return nil
}
