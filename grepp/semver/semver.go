// This file is part of grepp.
//
// Copyright (C) 2012-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// package semver
// Allows you to provide a simple consistent way to version different projects using SemVer.
package semver

import "fmt"

// The Version struct contains the information that MUST be provided in order to properly version a program.
type Version struct {
	Major, Minor, Patch int
	PreReleaseLabel     string
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
	return fmt.Sprintf("%d.%d.%d%s%s", sv.Major, sv.Minor, sv.Patch, label, build)
}
