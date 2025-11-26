// This file is part of ffind.
//
// Copyright (C) 2017-2025  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package ffind - file listing lib that allows to follow symlinks and skip files or directories based on various criteria.

Library goals:

• Return a file list given a dir.

•  Return the given file given a file.

• Do case insensitive (by default) or sensitive file matching.

• Allow to return files or dirs only.
Maybe build a list of common extensions in the skip code to allow for groups.
For example: '.rb' and '.erb' for ruby files.

• Follow Symlinks.
  - Is there a case where you don't want to? Allow disabling the follow anyway.

• Ignore hidden files (configurable).

  - In windows?

  - In Linux, ignore files starting with .

• Ignore git, svn and mercurial files (configurable).
*/
package ffind

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// FileError - Struct containing the File and Error information.
type FileError struct {
	FileInfo fs.FileInfo
	Path     string
	Error    error
}

// NewFileError - Given a filepath returns a FileError struct.
func NewFileError(fsys fs.FS, path string) (*FileError, error) {
	Logger.Printf("NewFileError: %s\n", path)
	ee, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	for _, e := range ee {
		Logger.Printf("\tEntry: %s\n", e.Name())
	}

	fInfo, err := fs.Lstat(fsys, path)
	if err != nil {
		Logger.Printf("NewFileError ERROR: %s\n", err)
		if os.IsNotExist(err) {
			// Clean up error context
			err = os.ErrNotExist
		}
	}
	return &FileError{fInfo, filepath.Clean(path), err}, err
}

// IsSymlink - Determine if FileError is describing a Symlink.
func (fe *FileError) IsSymlink() bool {
	return fe.FileInfo.Mode()&os.ModeSymlink != 0
}

// ReadDirNoSort - Same as fs/ReadDir but doesn't sort results.
//
//	Taken from https://golang.org/src/io/fs/readdir.go
//	Copyright 2020 The Go Authors. All rights reserved.
//	Use of this source code is governed by a BSD-style
//	license that can be found in the LICENSE file.
//
// func ReadDirNoSort(fsys fs.FS, name string) ([]fs.DirEntry, error) {
func ReadDirNoSort(fsys fs.FS, name string) ([]fs.FileInfo, error) {
	if fsys, ok := fsys.(fs.ReadDirFS); ok {
		list, err := fsys.ReadDir(name)
		// TODO
		if err != nil {
			return nil, err
		}
		entries := []fs.FileInfo{}
		for _, e := range list {
			info, _ := e.Info()
			entries = append(entries, info)
		}
		// TODO: Convert back to direntry

		return entries, err
	}
	file, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dir, ok := file.(fs.ReadDirFile)
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: errors.New("not implemented")}
	}

	list, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	entries := []fs.FileInfo{}
	for _, e := range list {
		info, _ := e.Info()
		entries = append(entries, info)
	}
	// TODO: Convert back to direntry

	return entries, err
}

// ListOneLevel - will return a one level list of FileError results under `path`.
func ListOneLevel(path string, follow bool, sortFn SortFn) <-chan FileError {

	fsys := os.DirFS(".")

	// Error gets passed to fe.Error, OK to ignore.
	fe, _ := NewFileError(fsys, path)
	return listOneLevel(fsys, fe, follow, sortFn)
}

// listOneLevel - will return a one level list of files under `FileError`.
// If `file` is a regular file, will return a FileError channel with itself.
// If `file` is a symlink and we are not following symlinks, will return a FileError channel with itself.
// If `file` is a symlink and we are following symlinks, will return a FileError channel with the readlink file.
// If `file` is a dir, will return a FileError channel with one level list under the dir.
func listOneLevel(
	fsys fs.FS,
	fe *FileError,
	follow bool,
	sortFn SortFn) <-chan FileError {
	fInfo := fe.FileInfo
	file := fe.Path
	Logger.Printf("file: %s\n", file)
	c := make(chan FileError)
	go func() {
		// Check for error
		if fe.Error != nil {
			Logger.Printf("listOneLevel entry error: %s", fe.Error.Error())
			c <- *fe
			close(c)
			return
		}
		// Check if file is symlink.
		nfe := fe
		if fe.IsSymlink() && follow {
			Logger.Printf("\tIsSymlink: %s", file)
			eval, err := filepath.EvalSymlinks(fe.Path)
			if err != nil {
				Logger.Printf("EvalSymlinks error: %s", err)
				// TODO: Clean up error description
				fe.Error = err
				c <- *fe
				close(c)
				return
			}
			nfe, err = NewFileError(fsys, eval)
			// TODO: Figure out how to add a test for this!
			if err != nil {
				Logger.Printf("NewFileError error: %s", err)
				fe.Error = err
				c <- *fe
				close(c)
				return
			}
			Logger.Printf("\tSymlink: %s", nfe.Path)
		}
		if nfe.FileInfo.IsDir() {
			Logger.Printf("\tDir: %s\n", fInfo.Name())
			fileMatches, err := ReadDirNoSort(fsys, file)
			if err != nil {
				c <- FileError{fInfo, filepath.Join(filepath.Dir(file), fInfo.Name()), err}
				close(c)
				return
			}
			sortFn(fileMatches)
			for _, fm := range fileMatches {
				c <- FileError{fm, filepath.Join(filepath.Clean(file), fm.Name()), err}
				Logger.Printf("\tFile: %s\n", fm.Name())
			}
			close(c)
			return
		}
		// If file is a regular file return the file and update the path to be the
		// dirname of the file in case of resolved symlinks.
		dirname := filepath.Dir(file)
		c <- FileError{fInfo, filepath.Join(dirname, fInfo.Name()), nil}
		close(c)
		return
	}()
	return c
}

// ListRecursive - will return a recursive list of FileError results under `path`.
func ListRecursive(path string,
	follow bool,
	s FileMatcher, sortFn SortFn) <-chan FileError {

	fsys := os.DirFS(".")
	fe, _ := NewFileError(fsys, path)
	return listRecursive(fsys, fe, follow, s, sortFn)
}

// listRecursive - will return a recursive list of files under `file`.
// If `file` is a regular file, will return a FileError channel with itself.
// If `file` is a symlink and we are not following symlinks, will return a FileError channel with itself.
// If `file` is a symlink and we are following symlinks, will return a FileError channel with the readlink file.
// If `file` is a dir, will return a FileError channel with one level list under the dir.
func listRecursive(fsys fs.FS, fe *FileError, follow bool, s FileMatcher, sortFn SortFn) <-chan FileError {
	c := make(chan FileError)
	go func() {
		if fe.Error != nil {
			Logger.Printf("\tError received: %s", fe.Error)
			c <- *fe
			close(c)
			return
		}
		Logger.Printf("Query: %s", fe.Path)
		ch := listOneLevel(fsys, fe, follow, sortFn)
		for e := range ch {
			Logger.Printf("\tReceived: %s", e.FileInfo.Name())
			if e.Error != nil {
				Logger.Printf("\tError received: %s", e.Error)
				c <- e
				continue
			}

			// Check if file is symlink.
			ne := &e
			checkSymlink := func() {
				if e.IsSymlink() && follow {
					Logger.Printf("\tIsSymlink: %s", e.Path)
					eval, err := filepath.EvalSymlinks(e.Path)
					if err != nil {
						Logger.Printf("\tEvalSymlinks error: %s", err)
						// If the link is broken then just return the original file
						if os.IsNotExist(err) {
							return
						}
						e.Error = err
						return
					}
					ne, err = NewFileError(fsys, eval)
					if err != nil {
						Logger.Printf("\tNew Error received: %s", err)
						e.Error = err
						return
					}
					Logger.Printf("\tSymlink: %s", ne.Path)
				}
			}
			checkSymlink()

			if ne.FileInfo.IsDir() {
				// TODO: Make sure to test SkipDirName
				if s.SkipDirName(e.FileInfo.Name()) {
					continue
				}
				if !s.SkipDirResults() {
					Logger.Printf("DIR: %s - %s", e.Path, ne.Path)
					c <- e
				}
				cr := listRecursive(fsys, &e, follow, s, sortFn)
				for e := range cr {
					Logger.Printf("Recurse: %s", e.Path)
					c <- e
				}
			} else {
				// TODO: Make sure to test SkipFileName
				if s.SkipFileResults() || s.SkipFileName(e.FileInfo.Name()) {
					continue
				}
				Logger.Printf("Else: %s", e.Path)
				if s.MatchFileName(e.FileInfo.Name()) {
					c <- e
				}
			}
		}
		close(c)
		return
	}()
	return c
}
