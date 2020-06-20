// This file is part of grepp.
//
// Copyright (C) 2012-2017  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package grepp

import (
	"testing"
)

func TestCheckMimeTypeForText(t *testing.T) {
	if IsTextMIME("grepp.png") {
		t.Errorf("Wrong mime detected for a PNG file\n")
	}
	if IsTextMIME("grepp") {
		t.Errorf("Wrong mime detected for an executable file\n")
	}
	if !IsTextMIME("grepp.txt") {
		t.Errorf("Wrong mime detected for a text file\n")
	}
	if !IsTextMIME("grepp.adoc") {
		t.Errorf("Wrong mime detected for an adoc file\n")
	}
}
