// This file is part of ffind.
//
// Copyright (C) 2017  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffind

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestBasicFileMatcher(t *testing.T) {
	log.SetOutput(os.Stderr)
	log.SetOutput(ioutil.Discard)

	// TODO: Windows test
	t.Run("nameIsHidden", func(t *testing.T) {
		bfm := BasicFileMatch{IgnoreHidden: true, IgnoreVCSDirs: true}
		if !bfm.SkipDirName(".dir") {
			t.Fatal("Not skipped")
		}
		if !bfm.SkipFileName(".file") {
			t.Fatal("Not skipped")
		}
		bfm = BasicFileMatch{IgnoreHidden: false, IgnoreVCSDirs: true}
		if bfm.SkipDirName(".dir") {
			t.Fatal("Skipped wrong file")
		}
		if bfm.SkipFileName(".file") {
			t.Fatal("Skipped wrong file")
		}
	})

	// log.SetOutput(os.Stderr)
	t.Run("MatchFileTypeList", func(t *testing.T) {
		bfm := BasicFileMatch{
			MatchFileTypeList: []string{"ruby"},
		}
		if bfm.MatchFileName("python.py") {
			t.Fatal("Wrong extension matched")
		}
		if !bfm.MatchFileName("ruby.rb") {
			t.Fatal("Expected extension not matched")
		}
		if !bfm.MatchFileName("Gemfile") {
			t.Fatal("Wrong filename matched")
		}
	})
	// log.SetOutput(ioutil.Discard)

	// log.SetOutput(os.Stderr)
	t.Run("MatchFileExtensionList", func(t *testing.T) {
		bfm := BasicFileMatch{
			MatchFileExtensionList: []string{".rb"},
		}
		if bfm.MatchFileName("python.py") {
			t.Fatal("Wrong extension matched")
		}
		if !bfm.MatchFileName("ruby.rb") {
			t.Fatal("Expected extension not matched")
		}
		if bfm.MatchFileName("Gemfile") {
			t.Fatal("Wrong filename matched")
		}
	})
	// log.SetOutput(ioutil.Discard)

	// log.SetOutput(os.Stderr)
	t.Run("IgnoreFileTypeList", func(t *testing.T) {
		bfm := BasicFileMatch{
			IgnoreFileTypeList: []string{"ruby"},
		}
		if bfm.SkipFileName("python.py") {
			t.Fatal("Expected extension ignored")
		}
		if !bfm.SkipFileName("ruby.rb") {
			t.Fatal("Wrong extension matched")
		}
		if !bfm.SkipFileName("Gemfile") {
			t.Fatal("Wrong filename matched")
		}
	})
	// log.SetOutput(ioutil.Discard)

	// log.SetOutput(os.Stderr)
	t.Run("IgnoreFileTypeList", func(t *testing.T) {
		bfm := BasicFileMatch{
			IgnoreFileExtensionList: []string{".rb"},
		}
		if bfm.SkipFileName("python.py") {
			t.Fatal("Expected extension ignored")
		}
		if !bfm.SkipFileName("ruby.rb") {
			t.Fatal("Wrong extension matched")
		}
	})
	// log.SetOutput(ioutil.Discard)
}
