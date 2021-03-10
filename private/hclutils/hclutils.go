// This file is part of dgtools.
//
// Copyright (C) 2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package hclutils provides utilities to work with custom DSLs based on HCL.
*/
package hclutils

import (
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// ParseHCL - Given HCL data it parses it into an *hclparse.Parser.
// It prints error diagnostics to the given writer.
// Even though the data is provided as an input, the filename is used for error diagnostic output.
func ParseHCL(w io.Writer, data []byte, filename string) (*hclparse.Parser, *hcl.File, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCL(data, filename)
	err := HandleDiags(w, parser, diags)
	if err != nil {
		return parser, f, fmt.Errorf("failure during input configuration parsing")
	}
	return parser, f, nil
}

// ParseHCLFile - Given an HCL file it parses it into an *hclparse.Parser.
// It prints error diagnostics to the given writer.
func ParseHCLFile(w io.Writer, filename string) (*hclparse.Parser, *hcl.File, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	err := HandleDiags(w, parser, diags)
	if err != nil {
		return parser, f, fmt.Errorf("failure during input configuration parsing")
	}
	return parser, f, nil
}

// HandleDiags - Outputs HCL parsing error diagnostics to the given writer.
func HandleDiags(w io.Writer, parser *hclparse.Parser, diags hcl.Diagnostics) error {
	if diags.HasErrors() {
		wr := hcl.NewDiagnosticTextWriter(
			w,              // writer to send messages to
			parser.Files(), // the parser's file cache, for source snippets
			100,            // wrapping width
			true,           // generate colored/highlighted output
		)
		err := wr.WriteDiagnostics(diags)
		if err != nil {
			return fmt.Errorf("errors found %w", err)
		}
		return fmt.Errorf("errors found")
	}
	return nil
}
