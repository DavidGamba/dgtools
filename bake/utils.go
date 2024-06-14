// This file is part of bake.
//
// Copyright (C) 2023-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"unicode"
)

func camelToKebab(camel string) string {
	var buffer bytes.Buffer
	for i, ch := range camel {
		if unicode.IsUpper(ch) && i > 0 && !unicode.IsUpper([]rune(camel)[i-1]) {
			buffer.WriteRune('-')
		}
		buffer.WriteRune(unicode.ToLower(ch))
	}
	return buffer.String()
}
