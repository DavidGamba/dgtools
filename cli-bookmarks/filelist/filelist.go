// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package filelist

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/ffind/lib/ffind"
)

// listFiles - return a one level list of files under given dir
func listFiles(dir string) ([]string, error) {
	fileList := []string{}
	fileError := ffind.ListOneLevel(dir, false, ffind.SortFnByName)
	for fe := range fileError {
		if fe.Error != nil {
			return fileList, fmt.Errorf("listing dir: %s | %s", dir, fe.Error)
		}
		if fe.FileInfo.IsDir() {
			fileList = append(fileList, fe.Path)
		}
	}
	return fileList, nil
}

// ListFilesWithFilter - return a one level list of files under given dir that
// case insensitive match the given prefixFilter
func ListFilesWithFilter(dir, prefixFilter string) ([]string, error) {
	log.Printf("[listFilesWithFilter] dir: %s, prefixFilter: %s\n", dir, prefixFilter)
	fileList := []string{}
	fullFileList, err := listFiles(dir)
	if err != nil {
		return fileList, err
	}
	prefixFilterLower := strings.ToLower(prefixFilter)
	for _, file := range fullFileList {
		base := filepath.Base(file)
		if strings.HasPrefix(strings.ToLower(base), prefixFilterLower) {
			fileList = append(fileList, file)
		}
	}
	log.Printf("[listFilesWithFilter] fileList: %v\n", fileList)
	return fileList, nil
}
