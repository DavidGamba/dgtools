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
