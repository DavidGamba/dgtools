// This file is part of ffind.
//
// Copyright (C) 2017  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffind

import (
	"fmt"
	"sort"
)

func PrintTypeList() {
	typeMap := getTypeMap()
	var types []string
	for key := range typeMap {
		types = append(types, key)
	}
	sort.Strings(types)
	for _, key := range types {
		fmt.Printf("%s: %v\n", key, typeMap[key])
	}
}

// KnownFileType - Given a filetype, returns true if it is known and there are
// rules for it or false if unknown.
func KnownFileType(fileType string) bool {
	typeMap := getTypeMap()
	if _, ok := typeMap[fileType]; ok {
		return true
	}
	return false
}

func getTypeMap() map[string][]string {
	typeMap := make(map[string][]string)
	for key, value := range typeListExt {
		for _, ext := range value {
			typeMap[key] = append(typeMap[key], "*"+ext)
		}
	}
	for key, value := range typeListFiles {
		typeMap[key] = append(typeMap[key], value...)
	}
	return typeMap
}

// This list was retrieved and adapted to golang from
// https://github.com/BurntSushi/ripgrep/blob/0668c74ed49320c4224c4ea526aca07692bf590a/ignore/src/types.rs
// Licensed under the UNLICENSE and the MIT.

var typeListExt = map[string][]string{
	"agda":         []string{".agda", ".lagda"},
	"asciidoc":     []string{".adoc", ".asc", ".asciidoc"},
	"asm":          []string{".asm", ".s", ".S"},
	"awk":          []string{".awk"},
	"c":            []string{".c", ".h", ".H"},
	"cabal":        []string{".cabal"},
	"cbor":         []string{".cbor"},
	"ceylon":       []string{".ceylon"},
	"clojure":      []string{".clj", ".cljc", ".cljs", ".cljx"},
	"cmake":        []string{".cmake"},
	"coffeescript": []string{".coffee"},
	"config":       []string{".config"},
	"cpp":          []string{".C", ".cc", ".cpp", ".cxx", ".h", ".H", ".hh", ".hpp", ".inl"},
	"creole":       []string{".creole"},
	"crystal":      []string{".cr"},
	"cs":           []string{".cs"},
	"csharp":       []string{".cs"},
	"cshtml":       []string{".cshtml"},
	"css":          []string{".css", ".scss"},
	"cython":       []string{".pyx"},
	"d":            []string{".d"},
	"dart":         []string{".dart"},
	"elisp":        []string{".el"},
	"elixir":       []string{".ex", ".eex", ".exs"},
	"erlang":       []string{".erl", ".hrl"},
	"fish":         []string{".fish"},
	"fortran":      []string{".f", ".F", ".f77", ".F77", ".pfo", ".f90", ".F90", ".f95", ".F95"},
	"fsharp":       []string{".fs", ".fsx", ".fsi"},
	"go":           []string{".go"},
	"golang":       []string{".go"},
	"groovy":       []string{".groovy", ".gradle"},
	"h":            []string{".h", ".hpp"},
	"haskell":      []string{".hs", ".lhs"},
	"hbs":          []string{".hbs"},
	"html":         []string{".htm", ".html", ".ejs"},
	"java":         []string{".java"},
	"jinja":        []string{".jinja", ".jinja2"},
	"js":           []string{".js", ".jsx", ".vue"},
	"json":         []string{".json"},
	"jsonl":        []string{".jsonl"},
	"julia":        []string{".jl"},
	"kotlin":       []string{".kt", ".kts"},
	"less":         []string{".less"},
	"lisp":         []string{".el", ".jl", ".lisp", ".lsp", ".sc", ".scm"},
	"log":          []string{".log"},
	"lua":          []string{".lua"},
	"m4":           []string{".ac", ".m4"},
	"make":         []string{".mk", ".mak"},
	"markdown":     []string{".markdown", ".md", ".mdown", ".mkdn"},
	"matlab":       []string{".m"},
	"md":           []string{".markdown", ".md", ".mdown", ".mkdn"},
	"ml":           []string{".ml"},
	"msbuild":      []string{".csproj", ".fsproj", ".vcxproj", ".proj", ".props", ".targets"},
	"nim":          []string{".nim"},
	"nix":          []string{".nix"},
	"objc":         []string{".h", ".m"},
	"objcpp":       []string{".h", ".mm"},
	"ocaml":        []string{".ml", ".mli", ".mll", ".mly"},
	"org":          []string{".org"},
	"pdf":          []string{".pdf"},
	"perl":         []string{".perl", ".pl", ".PL", ".plh", ".plx", ".pm", ".t"},
	"php":          []string{".php", ".php3", ".php4", ".php5", ".phtml"},
	"pod":          []string{".pod"},
	"ps":           []string{".cdxml", ".ps1", ".ps1xml", ".psd1", ".psm1"},
	"py":           []string{".py"},
	"python":       []string{".py"},
	"qmake":        []string{".pro", ".pri"},
	"r":            []string{".R", ".r", ".Rmd", ".Rnw"},
	"rdoc":         []string{".rdoc"},
	"rst":          []string{".rst"},
	"ruby":         []string{".gemspec", ".rb"},
	"rust":         []string{".rs"},
	"sass":         []string{".sass", ".scss"},
	"scala":        []string{".scala"},
	"sh":           []string{".bash", ".csh", ".ksh", ".sh", ".tcsh"},
	"spark":        []string{".spark"},
	"sql":          []string{".sql"},
	"stylus":       []string{".styl"},
	"sv":           []string{".v", ".vg", ".sv", ".svh", ".h"},
	"svg":          []string{".svg"},
	"swift":        []string{".swift"},
	"swig":         []string{".def", ".i"},
	"taskpaper":    []string{".taskpaper"},
	"tcl":          []string{".tcl"},
	"tex":          []string{".tex", ".ltx", ".cls", ".sty", ".bib"},
	"textile":      []string{".textile"},
	"toml":         []string{".toml"},
	"ts":           []string{".ts", ".tsx"},
	"twig":         []string{".twig"},
	"txt":          []string{".txt"},
	"vala":         []string{".vala"},
	"vb":           []string{".vb"},
	"vim":          []string{".vim"},
	"vimscript":    []string{".vim"},
	"wiki":         []string{".mediawiki", ".wiki"},
	"xml":          []string{".xml"},
	"yacc":         []string{".y"},
	"yaml":         []string{".yaml", ".yml"},
	"yocto":        []string{".bb", ".bbappend", ".bbclass"},
	"zsh":          []string{".zsh"},
}

var typeListFiles = map[string][]string{
	"cmake":   []string{"CMakeLists.txt"},
	"crystal": []string{"Projectfile"},
	"make":    []string{"gnumakefile", "Gnumakefile", "makefile", "Makefile"},
	"mk":      []string{"mkfile"},
	"ruby":    []string{"Gemfile", ".irbrc", "Rakefile"},
	"toml":    []string{"Cargo.lock"},
	"zsh":     []string{"zshenv", ".zshenv", "zprofile", ".zprofile", "zshrc", ".zshrc", "zlogin", ".zlogin", "zlogout", ".zlogout"},
}
