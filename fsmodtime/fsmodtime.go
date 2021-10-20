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
// * Targets might not exist yet.
//   In that case, build them.
// * All declared targets must exist.
//   Rebuild otherwise.
// * Not all sources are required to exist.
//   For example, I might want a blank *.jpg and *.png in my build system but some image types might not exist.
// * ExpandEnv but don't fail silently if my Env Var expansions fail.
// * Allow for globs.
//
// Example:
//
//   targets := []string{"$outputs_dir/doc.pdf", "$outputs_dir/*.html"}
//   sources := []string{"$src_dir/*.adoc", "$images_dir/*.jpg", "$images_dir/*.png"}
//   paths, modified, err := Target(os.DirFS("."), targets, sources)
package fsmodtime

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
)

var Logger = log.New(ioutil.Discard, "", log.LstdFlags)

var ErrorInvalidPath = fmt.Errorf("invalid path")
var ErrorInvalidFS = fmt.Errorf("invalid fs")

// Last - given a list of paths, it finds the file with the latest modTime and returns it.
//
//   root := "."
//   fileSystem := os.DirFS(root)
//   path, fi, err := Last(fileSystem, paths)
func Last(fsys fs.FS, paths []string) (filepath string, fileInfo fs.FileInfo, err error) {
	// TODO: Add option to skip descending into dirs.
	// TODO: Add option to follow symlinks.
	// TODO: Add variadic option definitions.
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

	err = walkPaths(fsys, paths, afterFn)
	if err != nil {
		return "", nil, err
	}

	return filepath, fileInfo, nil
}

// First - given a list of paths, it finds the file with the earliest modTime and returns it.
//
//   root := "."
//   fileSystem := os.DirFS(root)
//   path, fi, err := First(fileSystem, paths)
func First(fsys fs.FS, paths []string) (filepath string, fileInfo fs.FileInfo, err error) {
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

	err = walkPaths(fsys, paths, beforeFn)
	if err != nil {
		return "", nil, err
	}

	return filepath, fileInfo, nil
}

func walkPaths(fsys fs.FS, paths []string, fn fileInfoFn) error {
	if fsys == nil {
		return ErrorInvalidFS
	}

	for _, path := range paths {
		Logger.Printf("path: %s\n", path)

		// validate path
		if path == "" {
			return fmt.Errorf("%w: '%s'", ErrorInvalidPath, path)
		}
		fi, err := fs.Stat(fsys, path)
		if err != nil {
			return err
		}

		err = fileInfoIterate(fsys, filepath.Dir(path), fs.FileInfoToDirEntry(fi), fn)
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
func fileInfoIterate(fsys fs.FS, root string, de fs.DirEntry, fn fileInfoFn) (err error) {
	if de.IsDir() {
		dir := path.Join(root, de.Name())
		Logger.Printf("expand: %s\n", dir)
		dirEntries, err := fs.ReadDir(fsys, dir)
		if err != nil {
			return err
		}
		for _, de := range dirEntries {
			err := fileInfoIterate(fsys, dir, de, fn)
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
	return
}
