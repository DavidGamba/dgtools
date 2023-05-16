package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/cueutils"
)

//go:embed schema.cue
var schema embed.FS

type Config struct {
	DefaultGroup string `json:"default_group"`
	Group        map[string]struct {
		Member map[string]struct {
			Name        string
			City        string
			CountryCode string `json:"country_code"`
			TimeZone    string `json:"time_zone"`
			Admin1      string // Admin division 1 name (state, province, etc.)
			Type        string // person, city, country
		}
	}
}

func ReadConfig(ctx context.Context, filename string) (*Config, error) {

	configs := []cueutils.CueConfigFile{}

	schemaFilename := "schema.cue"
	schemaFH, err := schema.Open(schemaFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", schemaFilename, err)
	}
	defer schemaFH.Close()
	configs = append(configs, cueutils.CueConfigFile{Data: schemaFH, Name: schemaFilename})

	configFilename := filepath.Clean(filename)
	configFH, err := os.Open(configFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", configFilename, err)
	}
	defer configFH.Close()
	configs = append(configs, cueutils.CueConfigFile{Data: configFH, Name: configFilename})

	c := Config{}
	err = cueutils.Unmarshal(configs, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return &c, nil
}
