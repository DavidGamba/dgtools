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
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	cueErrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/interpreter/embed"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/gocode/gocodec"
)

var Logger = log.New(io.Discard, "", log.LstdFlags)

type CueConfigFile struct {
	Data io.Reader
	Name string
}

func NewValue() *cue.Value {
	return &cue.Value{}
}

// Given a set of cue files, it will aggregate them into a single cue config and then Unmarshal it unto the given data structure.
// If dir == "" it will default to the current directory.
// packageName can be set to _ to load files without a package.
// virtualCueModuleName can be set to a module name to add a module overlay or blank if the dir has a module already.
// Because CUE doesn't support hidden files, hidden files need to be passed as configs.
// value is a pointer receiver to a cue.Value and can be used on the caller side to print the cue values.
// Initialize with cueutils.NewValue()
func Unmarshal(configs []CueConfigFile, dir, packageName, virtualCueModuleName string, value *cue.Value, target any) error {
	err := GetValue(configs, dir, packageName, virtualCueModuleName, value)
	if err != nil {
		return err
	}

	err = value.Validate(
		cue.Final(),
		cue.Concrete(true),
		cue.Definitions(true),
		cue.Hidden(true),
		cue.Optional(true),
	)
	if err != nil {
		return fmt.Errorf("failed config validation: %v", cueErrors.Details(err, nil))
	}

	g := gocodec.New(cuecontext.New(), nil)
	err = g.Encode(*value, target)
	if err != nil {
		return fmt.Errorf("failed to encode cue values: %w", err)
	}
	return nil
}

// Given a set of cue files, it will aggregate them into a single cue config and update the given cue.Value.
// This allows for incomplete configuration that can be completed by the caller.
//
// If dir == "" it will default to the current directory.
// packageName can be set to _ to load files without a package.
// virtualCueModuleName can be set to a module name to add a module overlay or blank if the dir has a module already.
// Because CUE doesn't support hidden files, hidden files need to be passed as configs so that they can be loaded as an overlay.
//
// Completing the value can be done in a couple of ways:
//
// Using a go struct:
//
//	p := &GoStruct{}
//	c := cuecontext.New()
//	targetCue := c.Encode(p)
//	*value = (*value).Unify(targetCue)
//
// Using value.Fillpath (or cueutils.FillPaths):
//
//	data := "some data"
//	*value = value.FillPath(cue.ParsePath("path.in.cue"), data)
func GetValue(configs []CueConfigFile, dir, packageName, virtualCueModuleName string, value *cue.Value) error {
	embedding := cuecontext.Interpreter(embed.New())
	ctxOpts := []cuecontext.Option{embedding}
	c := cuecontext.New(ctxOpts...)

	insts := []*build.Instance{}
	var err error
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	Logger.Printf("dir abs: %s\n", dirAbs)

	overlay, packagePaths, err := BuildOverlay(configs, dirAbs, virtualCueModuleName)
	if err != nil {
		return fmt.Errorf("failed to build overlay: %w", err)
	}

	if dir == "" {
		dir = dirAbs
	}

	// Load the CUE package in the dir directory
	lc := &load.Config{
		Package:             packageName,
		ModuleRoot:          ".",
		AcceptLegacyModules: true,
		Dir:                 dir,
		Overlay:             overlay,
	}
	Logger.Printf("dir: %s\nModuleRoot: %s\npackagePaths: %v\n", dir, lc.ModuleRoot, packagePaths)
	ii := load.Instances(packagePaths, lc)
	logInstancesFiles(dir, ii)
	insts = append(insts, ii...)

	logInstancesFiles("building", insts)
	vv, err := c.BuildInstances(insts)
	if err != nil {
		return fmt.Errorf("failed to build instances: %w", err)
	}
	for _, v := range vv {
		*value = (*value).Unify(v)
	}

	if value.Err() != nil {
		return fmt.Errorf("failed to compile: %s", cueErrors.Details(value.Err(), nil))
	}

	return nil
}

// BuildOverlay - Builds an overlay of cue configs files.
// If virtualCueModuleName is not empty, it will add a module overlay with the given name.
// Returns an overlay map and the package paths found.
// Initially the package paths are set to the current directory but hidden files will be added so that they can be read.
func BuildOverlay(configs []CueConfigFile, overlayRootDir, virtualCueModuleName string) (map[string]load.Source, []string, error) {
	packagePaths := []string{"."}
	overlay := map[string]load.Source{}

	for i, cf := range configs {
		Logger.Printf("config: n: %d, name: %s\n", i, cf.Name)
		d, err := io.ReadAll(cf.Data)
		if err != nil {
			return overlay, packagePaths, fmt.Errorf("failed to read: %w", err)
		}

		abs, err := filepath.Abs(cf.Name)
		if err != nil {
			return overlay, packagePaths, fmt.Errorf("failed to get absolute path: %w", err)
		}
		fdir := filepath.Dir(abs)
		Logger.Printf("abs: %s, dir: %s\n", abs, fdir)

		overlayPath := filepath.Join(overlayRootDir, filepath.Base(cf.Name))
		overlay[overlayPath] = load.FromBytes(d)
		Logger.Printf("overlay: %s\n", overlayPath)
		// Add hidden file to the package paths
		if strings.HasPrefix(filepath.Base(cf.Name), ".") {
			packagePaths = append(packagePaths, overlayPath)
		}
	}
	if virtualCueModuleName != "" {
		modPath := filepath.Join(overlayRootDir, "cue.mod/module.cue")
		overlay[modPath] = load.FromString(fmt.Sprintf(`module: "%s"`, virtualCueModuleName))
		Logger.Printf("overlay: %s\n", modPath)
	}
	return overlay, packagePaths, nil
}

func logInstancesFiles(kind string, insts []*build.Instance) {
	for _, inst := range insts {
		for i, f := range inst.BuildFiles {
			Logger.Printf("%s: , n: %d, name: %s, file %s\n", kind, i, inst.ID(), f.Filename)
		}
	}
}

// Fill multiple paths in a cue.Value with the given data.
// The fillMap is a map of paths and data to fill.
func FillPaths(value *cue.Value, fillMap map[string]any) error {
	for path, data := range fillMap {
		*value = value.FillPath(cue.ParsePath(path), data)
		if value.Err() != nil {
			return fmt.Errorf("%s", cueErrors.Details(value.Err(), nil))
		}
	}
	return nil
}

// Validate a cue.Value
func Validate(value *cue.Value) error {
	err := value.Validate(
		cue.Final(),
		cue.Concrete(true),
		cue.Definitions(true),
		cue.Hidden(true),
		cue.Optional(true),
	)
	if err != nil {
		return fmt.Errorf("%s", cueErrors.Details(err, nil))
	}
	return nil
}
