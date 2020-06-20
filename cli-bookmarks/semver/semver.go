// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package semver - Allows you to provide a simple consistent way to version
different projects using SemVer.

It doesn't verify the PreReleaseLabel or the BuildMetadata for correctness.
*/
package semver

import "fmt"

// The Version struct contains the information that MUST be provided in order to properly version a program.
type Version struct {
	Version         string
	PreReleaseLabel string
}

// BuildMetadata is an optional field of the SemVer version.
var BuildMetadata string

func (sv Version) String() string {
	var label string
	var build string
	if sv.PreReleaseLabel != "" {
		label = fmt.Sprintf("-%s", sv.PreReleaseLabel)
	}
	if BuildMetadata != "" {
		build = fmt.Sprintf("+%s", BuildMetadata)
	}
	return fmt.Sprintf("%s%s%s", sv.Version, label, build)
}
