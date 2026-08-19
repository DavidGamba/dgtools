package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"github.com/DavidGamba/dgtools/cueutils"
	"github.com/DavidGamba/go-getoptions"
)

//go:embed schema.cue
var f embed.FS

var schemaFilename = "schema.cue"
var configFilename = "jsql.cue"

type Config struct {
	Provider map[string]ConfigProvider
}

type ConfigProvider struct {
	Name        string
	SchemaName  string `json:"schemaName"`
	GetCommands map[string]struct {
		Name    string
		Table   string
		Command []string
		Filter  []string
		Create  []string
		Args    []struct {
			Name        string
			Description string
		}
	} `json:"getCommands"`
	Macros []string
	Views  map[string]struct {
		Dependencies []string
		Queries      []string
	}
}

func ConfigRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")
	d := &Config{}
	// Provide a pointer receiver for evaluated data for mostly debugging purposes
	value := cueutils.NewValue()

	err := ReadConfig(d, value)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
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
		return fmt.Errorf("failed to print value: %w", err)
	}
	fmt.Printf("value:\n%v\n", string(v))
	return nil
}

func ReadConfig(data *Config, value *cue.Value) error {
	Logger.Printf("Reading Config")
	packageName := "jsql"
	virtualCueModuleName := "jsql.config"

	// Provide a pointer receiver for evaluated data for mostly debugging purposes
	if value == nil {
		value = cueutils.NewValue()
	}

	config := []cueutils.CueConfigFS{
		// schema
		{FS: f, Files: []string{schemaFilename}, Dir: "."},
		// global config
		{FS: os.DirFS(filepath.Join(os.Getenv("HOME"), ".config", "jsql")), Files: []string{"config.cue", "."}, Dir: ".", SkipNotExist: true},
		// local config
		{FS: os.DirFS("."), Files: []string{configFilename, "."}, Dir: ".", SkipNotExist: true},
	}

	err := cueutils.UnmarshalFS(config, packageName, virtualCueModuleName, value, data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
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

	_, err = cueutils.StringValue(value, opts)
	if err != nil {
		return fmt.Errorf("failed to print value: %w", err)
	}
	// Logger.Printf("value:\n%v\n", string(v))

	// Logger.Printf("data structure:\n%+v\n", data)

	return nil
}
