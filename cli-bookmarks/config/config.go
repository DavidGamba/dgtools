// This file is part of cli-bookmarks.
//
// Copyright (C) 2018  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// Config - config file bookmarks.
//
//   map[alias]fullPath
type Config struct {
	Bookmarks map[string]string
}

// ParseFile - Parse config file.
// Config file is of the form:
//
//     [bookmarks]
//     alias = "fullPath"
func ParseFile(filename string) (Config, error) {
	var c Config
	c.Bookmarks = make(map[string]string)
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return c, fmt.Errorf("opening config file: %s", err)
	}
	return Parse(string(bs))
}

// Parse - Parse config string.
// Config string is of the form:
//
//     [bookmarks]
//     alias = "fullPath"
func Parse(configStr string) (Config, error) {
	var c Config
	c.Bookmarks = make(map[string]string)
	_, err := toml.Decode(configStr, &c)
	if err != nil {
		return c, fmt.Errorf("parsing config: %s", err)
	}
	return c, nil
}
