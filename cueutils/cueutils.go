// This file is part of cueutils.
//
// Copyright (C) 2023  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package cueutils provides helpers to work with Cue
*/
package cueutils

import (
	"fmt"
	"io"
	"log"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueErrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/gocode/gocodec"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

type CueConfigFile struct {
	Data io.Reader
	Name string
}

// Given a set of cue files, it will aggregate them into a single cue config and then Unmarshal it unto the given data structure.
func Unmarshal(configs []CueConfigFile, v any) error {
	c := cuecontext.New()
	value := cue.Value{}
	for i, reader := range configs {
		d, err := io.ReadAll(reader.Data)
		if err != nil {
			return fmt.Errorf("failed to read: %w", err)
		}
		// Logger.Printf("compiling %s\n", reader.Name)
		var t cue.Value
		if i == 0 {
			t = c.CompileBytes(d, cue.Filename(reader.Name))
		} else {
			t = c.CompileBytes(d, cue.Filename(reader.Name), cue.Scope(value))
		}
		value = value.Unify(t)
	}
	if value.Err() != nil {
		return fmt.Errorf("failed to compile: %s", cueErrors.Details(value.Err(), nil))
	}
	err := value.Validate(
		cue.Final(),
		cue.Concrete(true),
		cue.Definitions(true),
		cue.Hidden(true),
		cue.Optional(true),
	)
	if err != nil {
		return fmt.Errorf("failed config validation: %s", cueErrors.Details(err, nil))
	}

	g := gocodec.New(c, nil)
	err = g.Encode(value, &v)
	if err != nil {
		return fmt.Errorf("failed to encode cue values: %w", err)
	}
	return nil
}
