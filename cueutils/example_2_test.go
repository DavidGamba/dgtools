package cueutils_test

import (
	"fmt"
	"io"
	"os"

	"github.com/DavidGamba/dgtools/cueutils"
)

func ExampleUnmarshalFS() {
	cueutils.Logger.SetOutput(os.Stderr)

	// Read embedded schemas or explicit config files
	schemaFilename := "testschemas/myPackage-schema.cue"
	configFilename := ".myPackage-data-hola.cue"

	// Filter to a given package name or set to _ for files without a package
	packageName := "myPackage"

	// Virtual module name to be used for the cue files, can be left blank if there is already a module in the dir
	virtualCueModuleName := "my.module"

	// Provide a pointer receiver for evaluated data for mostly debugging purposes
	value := cueutils.NewValue()

	// Unmarshal into local data structure
	type MyDataStructure struct {
		Hello string
		Hola  string
	}
	d := MyDataStructure{}

	config := []cueutils.CueConfigFS{
		{FS: f, Files: []string{schemaFilename}, Dir: "."},
		{FS: os.DirFS("testdata"), Files: []string{configFilename, "."}, Dir: "."},
	}

	err := cueutils.UnmarshalFS(config, packageName, virtualCueModuleName, value, &d)
	if err != nil {
		fmt.Printf("failed to unmarshal: %s", err)
		return
	}

	// Print the config values
	opts := cueutils.StringValueOpts{
		Definitions:    true,
		Hidden:         true,
		Attributes:     true,
		Optional:       true,
		ErrorsAsValues: true,
		Concrete:       true,
	}

	v, err := cueutils.StringValue(value, opts)
	if err != nil {
		fmt.Printf("failed to print value: %s", err)
		return
	}
	fmt.Printf("value:\n%v\n", string(v))

	fmt.Printf("data structure:\n%+v\n", d)
	cueutils.Logger.SetOutput(io.Discard)

	// Output: value:
	// {
	// 	hello: "hello"
	// 	hola:  "hola"
	// }
	// data structure:
	// {Hello:hello Hola:hola}
	return
}
