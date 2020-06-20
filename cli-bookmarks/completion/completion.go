// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this

package completion

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/DavidGamba/cli-bookmarks/filelist"
)

// fileListCompletion - Given an alias for a dir, a dir and a user
// entry, return a fileList of results where the path is prefixed by alias.
//
// Examples (* denotes fileList under dir):
//
//     alias dir alias - alias/*
//     alias dir alias/d - alias/d*
//     alias dir alias/dir/ - alias/dir/*
func fileListCompletion(alias, dir, entry string) ([]string, error) {
	originalDir := dir
	log.Printf("[fileListCompletion] alias: %s, dir: %s, entry: %s\n", alias, dir, entry)
	aliasFileList := []string{}
	prefixFilter := getPrefixFilter(alias, entry)
	log.Printf("[fileListCompletion] prefixFilter: %s\n", prefixFilter)
	if strings.Contains(prefixFilter, string(os.PathSeparator)) {
		prefixPath := filepath.Dir(prefixFilter)
		dir = filepath.Join(dir, prefixPath)
		if strings.HasSuffix(prefixFilter, string(os.PathSeparator)) {
			prefixFilter = ""
		} else {
			prefixFilter = filepath.Base(prefixFilter)
		}
		log.Printf("[fileListCompletion] prefixPath: %s, dir: %s, prefixFilter: %s\n", prefixPath, dir, prefixFilter)
	}
	fileList, err := filelist.ListFilesWithFilter(dir, prefixFilter)
	if err != nil {
		return aliasFileList, err
	}
	for _, file := range fileList {
		fileRelativeToDir, err := filepath.Rel(originalDir, file)
		if err != nil {
			return []string{}, err
		}
		aliasFileList = append(aliasFileList, alias+string(os.PathSeparator)+fileRelativeToDir)
	}
	log.Printf("[fileListCompletion] aliasFileList: %s\n", aliasFileList)
	return aliasFileList, nil
}

func getPrefixFilter(alias, entry string) string {
	temp := strings.TrimLeft(entry, alias)
	return strings.TrimLeft(temp, string(os.PathSeparator))
}

// CompletionResults - Given a map of bookmarks and a user entry, return a list
// of valid completions.
//
// If there is a single match, append / to the end.
func CompletionResults(bookmarks map[string]string, entry string) ([]string, error) {
	log.Printf("[completionResults] entry: %s\n", entry)
	// If the entry has a / it indicates that it already matches a key and we are processing its subcompletions.
	if strings.Contains(entry, string(os.PathSeparator)) {
		parts := strings.SplitN(entry, string(os.PathSeparator), 2)
		alias := parts[0]
		dir := ""
		if v, ok := bookmarks[alias]; ok {
			dir = v
		} else {
			return []string{}, fmt.Errorf("unknown alias provided: %s", alias)
		}
		aliasFileList, err := fileListCompletion(alias, dir, entry)
		if err != nil {
			return []string{}, err
		}
		// If there is a single match, append / to the end.
		if len(aliasFileList) == 1 {
			if !strings.HasSuffix(aliasFileList[0], string(os.PathSeparator)) {
				aliasFileList[0] = aliasFileList[0] + string(os.PathSeparator)
			}
		}
		return aliasFileList, nil
	}

	matchingAliases := matchAlias(bookmarks, entry)
	return matchingAliases, nil
}

// matchAlias - given a bookmarks map and an entry, it will return all the
// aliases that partially match the entry.
//
// If there is a full match to the alias, append / to the end.
// If there is a single match, append / to the end.
//
// TODO: Should we allow case insensitive matching of aliases?
func matchAlias(bookmarks map[string]string, entry string) []string {
	matchingAliases := []string{}
	// Go through all the bookmarks and filter the ones that match the given entry only partially
	for k := range bookmarks {
		if entry == k {
			matchingAliases = append(matchingAliases, entry+string(os.PathSeparator))
		} else if strings.HasPrefix(k, entry) {
			matchingAliases = append(matchingAliases, k)
		}
	}
	if len(matchingAliases) == 1 {
		if !strings.Contains(matchingAliases[0], string(os.PathSeparator)) {
			matchingAliases[0] = matchingAliases[0] + string(os.PathSeparator)
		}
	}
	sort.Strings(matchingAliases)
	return matchingAliases
}
