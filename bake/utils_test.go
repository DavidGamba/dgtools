// This file is part of bake.
//
// Copyright (C) 2023-2024  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import "testing"

func TestCamelToKebab(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "single",
			in:   "A",
			out:  "a",
		},
		{
			name: "lower",
			in:   "abc",
			out:  "abc",
		},
		{
			name: "upper",
			in:   "ABC",
			out:  "abc",
		},
		{
			name: "mixed",
			in:   "aBC",
			out:  "a-bc",
		},
		{
			name: "mixed2",
			in:   "AbC",
			out:  "ab-c",
		},
		{
			name: "mixed10",
			in:   "AbCdEfGhIjK",
			out:  "ab-cd-ef-gh-ij-k",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := camelToKebab(tt.in); got != tt.out {
				t.Errorf("got %v, want %v", got, tt.out)
			}
		})
	}
}

func TestKebabToCamel(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "single",
			in:   "a",
			out:  "a",
		},
		{
			name: "lower",
			in:   "abc",
			out:  "abc",
		},
		{
			name: "upper",
			in:   "ABC",
			out:  "abc",
		},
		{
			name: "mixed",
			in:   "a-bc",
			out:  "aBc",
		},
		{
			name: "mixed2",
			in:   "ab-c",
			out:  "abC",
		},
		{
			name: "mixed10",
			in:   "ab-cd-ef-gh-ij-k",
			out:  "abCdEfGhIjK",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kebabToCamel(tt.in); got != tt.out {
				t.Errorf("got %v, want %v", got, tt.out)
			}
		})
	}
}
