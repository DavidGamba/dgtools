// This file is part of grepp.
//
// Copyright (C) 2012-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
package logging - Wrapper around log
*/
package logging

import (
	"io"
	"log"
)

var (
	Trace   *log.Logger
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func LogInit(th, dh, ih, wh, eh io.Writer) {
	Trace = log.New(th, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(dh, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(ih, "", 0)
	Warning = log.New(wh, "WARNING: ", 0)
	Error = log.New(eh, "ERROR: ", 0)
}
