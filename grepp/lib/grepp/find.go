// This file is part of grepp.
//
// Copyright (C) 2012-2017  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package grepp

import (
	"mime"
	"path"
	"strings"
)

// TODO: Check if there are any additional mime types we should allow.
// IsTextMIME - Determines if the file is a text based file by its extension.
func IsTextMIME(filename string) bool {
	ext := path.Ext(filename)
	// If there is no extension assume binary
	if ext == "" {
		return false
	}
	s := mime.TypeByExtension(ext)
	// If there is no associated mime assume text
	if s == "" {
		return true
	}
	return strings.HasPrefix(s, "text/")
}
