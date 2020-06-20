// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package semver

import "testing"

func TestVersion(t *testing.T) {
	v := Version{"1.2.3", ""}
	if v.String() != "1.2.3" {
		t.Fatalf("Version: 1.2.3 != %s\n", v.String())
	}
}

func TestPreReleaseLabel(t *testing.T) {
	v := Version{"1.2.3", "alpha"}
	if v.String() != "1.2.3-alpha" {
		t.Fatalf("Version: 1.2.3-alpha != %s\n", v.String())
	}
}

func TestBuildMetadata(t *testing.T) {
	BuildMetadata = "abc"
	v := Version{"1.2.3", "alpha"}
	if v.String() != "1.2.3-alpha+abc" {
		t.Fatalf("Version: 1.2.3-alpha+abc != %s\n", v.String())
	}
}
