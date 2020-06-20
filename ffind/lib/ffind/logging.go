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
)

var logger = log.New(ioutil.Discard, "", 0)

// SetLogger - Define a logger for the library.
func SetLogger(newLogger *log.Logger) {
	logger = newLogger
}
