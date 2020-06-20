// This file is part of ffind.
//
// Copyright (C) 2017  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffind

import (
	"os"
	"sort"
	"strconv"
)

// TODO: This is rather hard to test, very simple code though.
// Might want to make this string based by making a slice of names.

// TODO: Add sortFnByVersion that allows to sort by a number within the name string.

// SortFn - takes a slice of os.FileInfo and sorts it.
type SortFn func(a []os.FileInfo)

// byName implements a num sorted sort.Interface.
type byNum []os.FileInfo

func (f byNum) Len() int      { return len(f) }
func (f byNum) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f byNum) Less(i, j int) bool {
	nai, err := strconv.Atoi(f[i].Name())
	if err != nil {
		return f[i].Name() < f[j].Name()
	}
	naj, err := strconv.Atoi(f[j].Name())
	if err != nil {
		return f[i].Name() < f[j].Name()
	}
	return nai < naj
}

// SortFnByNum - SortFn that returns a list sorted in numerical order in case
// the full filename is a number. Otherwise, it does string comparison.
func SortFnByNum(a []os.FileInfo) {
	sort.Sort(byNum(a))
}

// byName implements a name sorted sort.Interface.
type byName []os.FileInfo

func (f byName) Len() int      { return len(f) }
func (f byName) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f byName) Less(i, j int) bool {
	return f[i].Name() < f[j].Name()
}

// SortFnByName - SortFn that does a simple string comparison to return a sorted list.
func SortFnByName(a []os.FileInfo) {
	sort.Sort(byName(a))
}
