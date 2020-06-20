// This file is part of ffind.
//
// Copyright (C) 2017  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffind

import (
	"strings"
)

// FileMatcher - inteface used to skip files and dirs
type FileMatcher interface {
	// Do not show dir results in output.
	// It still recurses into dirs when set.
	SkipDirResults() bool
	// Do not show file results in output.
	SkipFileResults() bool
	// Skip directory based on name.
	// Will not recurse into this directory.
	SkipDirName(name string) bool
	// Skip file based on name.
	SkipFileName(name string) bool
	// Include file based on name.
	MatchFileName(name string) bool
}

// BasicFileMatch - A simple FileMatcher interface implementation.
type BasicFileMatch struct {
	// Ignore files from output.
	IgnoreFileResults bool
	// Ignore dirs from output.
	IgnoreDirResults bool
	// Ignores .git, .svn and .hg directories.
	IgnoreVCSDirs bool
	// Ignores files and dirs with names starting in .
	IgnoreHidden bool
	// Ignores dirs that are case insensitive equal to any of the provided list of strings.
	IgnoreDirEqualsList []string
	// Ignores dirs that are case sensitive equal to any of the provided list of strings.
	IgnoreDirEqualsListCase []string
	// Ignores dirs that contain any of the case insensitive provided strings.
	IgnoreDirContainsList []string
	// Ignores dirs that contain any of the case sensitive provided strings.
	IgnoreDirContainsListCase []string
	// Ignores files that are case insensitive equal to any of the provided list of strings.
	IgnoreFileEqualsList []string
	// Ignores files that are case sensitive equal to any of the provided list of strings.
	IgnoreFileEqualsListCase []string
	// Ignores files that contain any of the case insensitive provided strings.
	IgnoreFileContainsList []string
	// Ignores files that contain any of the case sensitive provided strings.
	IgnoreFileContainsListCase []string
	// Ignores files that end in any of the case insensitive provided strings.
	// Dot is not included by default, so it must be provided in the list.
	IgnoreFileExtensionList []string
	// Matches files that end in any of the case insensitive provided strings.
	// Dot is not included by default, so it must be provided in the list.
	MatchFileExtensionList []string
	// Ignores files that end in any of the extensions listed in the provided type.
	IgnoreFileTypeList []string
	// Matches files that end in any of the extensions listed in the provided type.
	MatchFileTypeList []string
}

// nameInEqualsList - Case Insensitive equals matching.
func nameInEqualsList(name string, list []string) bool {
	for _, entry := range list {
		if strings.EqualFold(name, entry) {
			logger.Printf("Exclude %s with entry %s", name, entry)
			return true
		}
	}
	return false
}

// nameInEqualsListCase - Case Sensitive equals matching.
func nameInEqualsListCase(name string, list []string) bool {
	for _, entry := range list {
		if name == entry {
			logger.Printf("Exclude %s with entry %s", name, entry)
			return true
		}
	}
	return false
}

// nameInContainsList - Case Insensitive contains matching.
func nameInContainsList(name string, list []string) bool {
	for _, entry := range list {
		if strings.Contains(
			strings.ToLower(name),
			strings.ToLower(entry)) {
			logger.Printf("Exclude %s with entry %s", name, entry)
			return true
		}
	}
	return false
}

// nameInContainsListCase - Case Sensitive contains matching.
func nameInContainsListCase(name string, list []string) bool {
	for _, entry := range list {
		if strings.Contains(name, entry) {
			logger.Printf("Exclude %s with entry %s", name, entry)
			return true
		}
	}
	return false
}

// nameInExtensionList - Case Insensitive suffix matching.
func nameInExtensionList(name string, list []string) bool {
	for _, entry := range list {
		if strings.HasSuffix(
			strings.ToLower(name),
			strings.ToLower(entry)) {
			logger.Printf("Exclude %s with entry %s", name, entry)
			return true
		}
	}
	return false
}

// TODO: If I ever want to try this in Windows, look into making this portable.
func nameIsHidden(name string, ignoreHidden bool) bool {
	if ignoreHidden && name != "." && strings.HasPrefix(name, ".") {
		logger.Printf("Exclude hidden file %s", name)
		return true
	}
	return false
}

// SkipDirName - Skip Dir based on Name
func (l *BasicFileMatch) SkipDirName(name string) bool {
	if l.IgnoreVCSDirs {
		vcsList := []string{".git", ".svn", ".hg"}
		if nameInEqualsList(name, vcsList) {
			return true
		}
	} else {
		// TODO: Add test to catch when this is not overriden to false
		l.IgnoreHidden = false
	}
	switch {
	case nameIsHidden(name, l.IgnoreHidden):
		return true
	case nameInEqualsListCase(name, l.IgnoreDirEqualsListCase):
		return true
	case nameInEqualsList(name, l.IgnoreDirEqualsList):
		return true
	case nameInContainsListCase(name, l.IgnoreDirContainsListCase):
		return true
	case nameInContainsList(name, l.IgnoreDirContainsList):
		return true
	}
	return false
}

// SkipDirResults  - Skip Dir Result Listing
func (l *BasicFileMatch) SkipDirResults() bool {
	if len(l.MatchFileExtensionList) != 0 || len(l.MatchFileTypeList) != 0 {
		return true
	}
	return l.IgnoreDirResults
}

// SkipFileName - Skip File based on Name
func (l *BasicFileMatch) SkipFileName(name string) bool {
	switch {
	case nameIsHidden(name, l.IgnoreHidden):
		return true
	case nameInEqualsListCase(name, l.IgnoreFileEqualsListCase):
		return true
	case nameInEqualsList(name, l.IgnoreFileEqualsList):
		return true
	case nameInContainsListCase(name, l.IgnoreFileContainsListCase):
		return true
	case nameInContainsList(name, l.IgnoreFileContainsList):
		return true
	case nameInExtensionList(name, l.IgnoreFileExtensionList):
		return true
	case matchFileToTypeList(name, l.IgnoreFileTypeList):
		return true
	}
	return false
}

// SkipFileResults - Skip File Result Listing
func (l *BasicFileMatch) SkipFileResults() bool {
	return l.IgnoreFileResults
}

func matchFileToTypeList(name string, typeList []string) bool {
	for _, fileType := range typeList {
		logger.Printf("fileType %s\n", fileType)
		// Match against full filename first
		if filenameList, ok := typeListFiles[fileType]; ok {
			logger.Printf("filenameList: %v\n", filenameList)
			if nameInEqualsListCase(name, filenameList) {
				return true
			}
		}
		// Match against ext
		if extList, ok := typeListExt[fileType]; ok {
			logger.Printf("extList: %v\n", extList)
			if nameInExtensionList(name, extList) {
				return true
			}
		}
	}
	return false
}

// MatchFileName - Match File based on Name
func (l *BasicFileMatch) MatchFileName(name string) bool {
	switch {
	case len(l.MatchFileExtensionList) > 0:
		if nameInExtensionList(name, l.MatchFileExtensionList) {
			return true
		}
	case len(l.MatchFileTypeList) > 0:
		if matchFileToTypeList(name, l.MatchFileTypeList) {
			return true
		}
	case len(l.MatchFileExtensionList) == 0 && len(l.MatchFileTypeList) == 0:
		return true
	}
	return false
}
