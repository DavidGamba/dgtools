// This file is part of ffind.
//
// Copyright (C) 2017  David Gamba Rios
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
  • Is there a case where you don't want to? Allow disabling the follow anyway.

• Ignore hidden files (configurable).

  • In windows?

  • In Linux, ignore files starting with .

• Ignore git, svn and mercurial files (configurable).

*/
package ffind

import (
	"os"
	"path/filepath"
)

// FileError - Struct containing the File and Error information.
type FileError struct {
	FileInfo os.FileInfo
	Path     string
	Error    error
}

// NewFileError - Given a filepath returns a FileError struct.
func NewFileError(path string) (*FileError, error) {
	logger.Printf("NewFileError: %s", path)
	fInfo, err := os.Lstat(path)
	if err != nil {
		logger.Printf("NewFileError ERROR: %s", err)
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

// ReadDirNoSort - Same as ioutil/ReadDir but doesn't sort results.
// It also cleans up the error from the open call.
//
//   Taken from https://golang.org/src/io/ioutil/ioutil.go
//   Copyright 2009 The Go Authors. All rights reserved.
//   Use of this source code is governed by a BSD-style
//   license that can be found in the LICENSE file.
func ReadDirNoSort(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		logger.Printf("ReadDirNoSort ERROR: %s", err)
		if os.IsPermission(err) {
			// Clean up error context to make the output nicer
			err = os.ErrPermission
		}
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	return list, nil
}

// ListOneLevel - will return a one level list of FileError results under `path`.
func ListOneLevel(path string, follow bool, sortFn SortFn) <-chan FileError {
	// Error gets passed to fe.Error, OK to ignore.
	fe, _ := NewFileError(path)
	return listOneLevel(fe, follow, sortFn)
}

// listOneLevel - will return a one level list of files under `FileError`.
// If `file` is a regular file, will return a FileError channel with itself.
// If `file` is a symlink and we are not following symlinks, will return a FileError channel with itself.
// If `file` is a symlink and we are following symlinks, will return a FileError channel with the readlink file.
// If `file` is a dir, will return a FileError channel with one level list under the dir.
func listOneLevel(
	fe *FileError,
	follow bool,
	sortFn SortFn) <-chan FileError {
	fInfo := fe.FileInfo
	file := fe.Path
	logger.Printf("file: %s\n", file)
	c := make(chan FileError)
	go func() {
		// Check for error
		if fe.Error != nil {
			logger.Printf("listOneLevel entry error: %s", fe.Error.Error())
			c <- *fe
			close(c)
			return
		}
		// Check if file is symlink.
		nfe := fe
		if fe.IsSymlink() && follow {
			logger.Printf("\tIsSymlink: %s", file)
			eval, err := filepath.EvalSymlinks(fe.Path)
			if err != nil {
				logger.Printf("EvalSymlinks error: %s", err)
				// TODO: Clean up error description
				fe.Error = err
				c <- *fe
				close(c)
				return
			}
			nfe, err = NewFileError(eval)
			// TODO: Figure out how to add a test for this!
			if err != nil {
				logger.Printf("NewFileError error: %s", err)
				fe.Error = err
				c <- *fe
				close(c)
				return
			}
			logger.Printf("\tSymlink: %s", nfe.Path)
		}
		if nfe.FileInfo.IsDir() {
			logger.Printf("\tDir: %s\n", fInfo.Name())
			fileMatches, err := ReadDirNoSort(file)
			if err != nil {
				c <- FileError{fInfo, filepath.Join(filepath.Dir(file), fInfo.Name()), err}
				close(c)
				return
			}
			sortFn(fileMatches)
			for _, fm := range fileMatches {
				c <- FileError{fm, filepath.Join(filepath.Clean(file), fm.Name()), err}
				logger.Printf("\tFile: %s\n", fm.Name())
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
	fe, _ := NewFileError(path)
	return listRecursive(fe, follow, s, sortFn)
}

// listRecursive - will return a recursive list of files under `file`.
// If `file` is a regular file, will return a FileError channel with itself.
// If `file` is a symlink and we are not following symlinks, will return a FileError channel with itself.
// If `file` is a symlink and we are following symlinks, will return a FileError channel with the readlink file.
// If `file` is a dir, will return a FileError channel with one level list under the dir.
func listRecursive(fe *FileError, follow bool, s FileMatcher, sortFn SortFn) <-chan FileError {
	c := make(chan FileError)
	go func() {
		if fe.Error != nil {
			logger.Printf("\tError received: %s", fe.Error)
			c <- *fe
			close(c)
			return
		}
		logger.Printf("Query: %s", fe.Path)
		ch := listOneLevel(fe, follow, sortFn)
		for e := range ch {
			logger.Printf("\tReceived: %s", e.FileInfo.Name())
			if e.Error != nil {
				logger.Printf("\tError received: %s", e.Error)
				c <- e
				continue
			}

			// Check if file is symlink.
			ne := &e
			checkSymlink := func() {
				if e.IsSymlink() && follow {
					logger.Printf("\tIsSymlink: %s", e.Path)
					eval, err := filepath.EvalSymlinks(e.Path)
					if err != nil {
						logger.Printf("\tEvalSymlinks error: %s", err)
						// If the link is broken then just return the original file
						if os.IsNotExist(err) {
							return
						}
						e.Error = err
						return
					}
					ne, err = NewFileError(eval)
					if err != nil {
						logger.Printf("\tNew Error received: %s", err)
						e.Error = err
						return
					}
					logger.Printf("\tSymlink: %s", ne.Path)
				}
			}
			checkSymlink()

			if ne.FileInfo.IsDir() {
				// TODO: Make sure to test SkipDirName
				if s.SkipDirName(e.FileInfo.Name()) {
					continue
				}
				if !s.SkipDirResults() {
					logger.Printf("DIR: %s - %s", e.Path, ne.Path)
					c <- e
				}
				cr := listRecursive(&e, follow, s, sortFn)
				for e := range cr {
					logger.Printf("Recurse: %s", e.Path)
					c <- e
				}
			} else {
				// TODO: Make sure to test SkipFileName
				if s.SkipFileResults() || s.SkipFileName(e.FileInfo.Name()) {
					continue
				}
				logger.Printf("Else: %s", e.Path)
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
