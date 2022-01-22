// This file is part of ffind.
//
// Copyright (C) 2017-2022  David Gamba Rios
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
	"agda":         {".agda", ".lagda"},
	"asciidoc":     {".adoc", ".asc", ".asciidoc"},
	"asm":          {".asm", ".s", ".S"},
	"awk":          {".awk"},
	"c":            {".c", ".h", ".H"},
	"cabal":        {".cabal"},
	"cbor":         {".cbor"},
	"ceylon":       {".ceylon"},
	"clojure":      {".clj", ".cljc", ".cljs", ".cljx"},
	"cmake":        {".cmake"},
	"coffeescript": {".coffee"},
	"config":       {".config"},
	"cpp":          {".C", ".cc", ".cpp", ".cxx", ".h", ".H", ".hh", ".hpp", ".inl"},
	"creole":       {".creole"},
	"crystal":      {".cr"},
	"cs":           {".cs"},
	"csharp":       {".cs"},
	"cshtml":       {".cshtml"},
	"css":          {".css", ".scss"},
	"cython":       {".pyx"},
	"d":            {".d"},
	"dart":         {".dart"},
	"elisp":        {".el"},
	"elixir":       {".ex", ".eex", ".exs"},
	"erlang":       {".erl", ".hrl"},
	"fish":         {".fish"},
	"fortran":      {".f", ".F", ".f77", ".F77", ".pfo", ".f90", ".F90", ".f95", ".F95"},
	"fsharp":       {".fs", ".fsx", ".fsi"},
	"go":           {".go"},
	"golang":       {".go"},
	"groovy":       {".groovy", ".gradle"},
	"h":            {".h", ".hpp"},
	"haskell":      {".hs", ".lhs"},
	"hbs":          {".hbs"},
	"html":         {".htm", ".html", ".ejs"},
	"java":         {".java"},
	"jinja":        {".jinja", ".jinja2"},
	"js":           {".js", ".jsx", ".vue"},
	"json":         {".json"},
	"jsonl":        {".jsonl"},
	"julia":        {".jl"},
	"kotlin":       {".kt", ".kts"},
	"less":         {".less"},
	"lisp":         {".el", ".jl", ".lisp", ".lsp", ".sc", ".scm"},
	"log":          {".log"},
	"lua":          {".lua"},
	"m4":           {".ac", ".m4"},
	"make":         {".mk", ".mak"},
	"markdown":     {".markdown", ".md", ".mdown", ".mkdn"},
	"matlab":       {".m"},
	"md":           {".markdown", ".md", ".mdown", ".mkdn"},
	"ml":           {".ml"},
	"msbuild":      {".csproj", ".fsproj", ".vcxproj", ".proj", ".props", ".targets"},
	"nim":          {".nim"},
	"nix":          {".nix"},
	"objc":         {".h", ".m"},
	"objcpp":       {".h", ".mm"},
	"ocaml":        {".ml", ".mli", ".mll", ".mly"},
	"org":          {".org"},
	"pdf":          {".pdf"},
	"perl":         {".perl", ".pl", ".PL", ".plh", ".plx", ".pm", ".t"},
	"php":          {".php", ".php3", ".php4", ".php5", ".phtml"},
	"pod":          {".pod"},
	"ps":           {".cdxml", ".ps1", ".ps1xml", ".psd1", ".psm1"},
	"py":           {".py"},
	"python":       {".py"},
	"qmake":        {".pro", ".pri"},
	"r":            {".R", ".r", ".Rmd", ".Rnw"},
	"rdoc":         {".rdoc"},
	"rst":          {".rst"},
	"ruby":         {".gemspec", ".rb"},
	"rust":         {".rs"},
	"sass":         {".sass", ".scss"},
	"scala":        {".scala"},
	"sh":           {".bash", ".csh", ".ksh", ".sh", ".tcsh"},
	"spark":        {".spark"},
	"sql":          {".sql"},
	"stylus":       {".styl"},
	"sv":           {".v", ".vg", ".sv", ".svh", ".h"},
	"svg":          {".svg"},
	"swift":        {".swift"},
	"swig":         {".def", ".i"},
	"taskpaper":    {".taskpaper"},
	"tcl":          {".tcl"},
	"tex":          {".tex", ".ltx", ".cls", ".sty", ".bib"},
	"textile":      {".textile"},
	"toml":         {".toml"},
	"ts":           {".ts", ".tsx"},
	"twig":         {".twig"},
	"txt":          {".txt"},
	"vala":         {".vala"},
	"vb":           {".vb"},
	"vim":          {".vim"},
	"vimscript":    {".vim"},
	"wiki":         {".mediawiki", ".wiki"},
	"xml":          {".xml"},
	"yacc":         {".y"},
	"yaml":         {".yaml", ".yml"},
	"yocto":        {".bb", ".bbappend", ".bbclass"},
	"zsh":          {".zsh"},
}

var typeListFiles = map[string][]string{
	"cmake":   {"CMakeLists.txt"},
	"crystal": {"Projectfile"},
	"make":    {"gnumakefile", "Gnumakefile", "makefile", "Makefile"},
	"mk":      {"mkfile"},
	"ruby":    {"Gemfile", ".irbrc", "Rakefile"},
	"toml":    {"Cargo.lock"},
	"zsh":     {"zshenv", ".zshenv", "zprofile", ".zprofile", "zshrc", ".zshrc", "zlogin", ".zlogin", "zlogout", ".zlogout"},
}
