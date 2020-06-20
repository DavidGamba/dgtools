// This file is part of ffind.
//
// Copyright (C) 2017  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffind

import (
	"testing"
)

func TestKnownFileType(t *testing.T) {
	if !KnownFileType("ruby") {
		t.Fatalf("Expected FileType unknown")
	}
	if KnownFileType("_non_a_type_") {
		t.Fatalf("Unknown FileType known")
	}
}
