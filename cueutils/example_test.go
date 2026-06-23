package cueutils_test

import (
	"embed"
	"fmt"
	"os"

	"github.com/DavidGamba/dgtools/cueutils"
)

//go:embed testschemas/myPackage-schema.cue
var f embed.FS

func Example() {
	cueutils.Logger.SetOutput(os.Stderr)

	configs := []cueutils.CueConfigFile{}

	// Read embedded schemas or explicit config files
	schemaFilename := "testschemas/myPackage-schema.cue"
	schemaFH, err := f.Open(schemaFilename)
	if err != nil {
		fmt.Printf("failed to open '%s': %s", schemaFilename, err)
		return
	}
	defer schemaFH.Close()
	configs = append(configs, cueutils.CueConfigFile{Data: schemaFH, Name: schemaFilename})

	configFilename := "testdata/.myPackage-data-hola.cue"
	configFH, err := os.Open(configFilename)
	if err != nil {
		fmt.Printf("failed to open config file: %s", err)
		return
	}
	defer configFH.Close()
	configs = append(configs, cueutils.CueConfigFile{Data: configFH, Name: configFilename})

	extraFilename := "testdata/hola.txt"
	extraFH, err := os.Open(extraFilename)
	if err != nil {
		fmt.Printf("failed to open config file: %s", err)
		return
	}
	defer extraFH.Close()
	configs = append(configs, cueutils.CueConfigFile{Data: extraFH, Name: extraFilename})

	// Read all cue files from a given dir
	// Doesn't read hidden files, those need to be given explicitly in the configs list.
	dir := "testdata"

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

	err = cueutils.Unmarshal(configs, dir, packageName, virtualCueModuleName, value, &d)
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

	// Output: value:
	// {
	// 	hello: "hello"
	// 	hola:  "hola"
	// }
	// data structure:
	// {Hello:hello Hola:hola}
	return
}
