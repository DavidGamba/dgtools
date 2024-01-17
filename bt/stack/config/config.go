package config

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DavidGamba/dgtools/buildutils"
	"github.com/DavidGamba/dgtools/cueutils"
)

//go:embed schema.cue
var f embed.FS

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func Get(ctx context.Context, filename string) (*Config, string, error) {
	f, err := buildutils.FindFileUpwards(ctx, filename)
	if err != nil {
		cfg := &Config{}
		return cfg, f, fmt.Errorf("failed to find config file: %w", err)
	}

	configFH, err := os.Open(f)
	if err != nil {
		return &Config{}, f, fmt.Errorf("failed to open config file '%s': %w", f, err)
	}
	defer configFH.Close()

	cfg, err := Read(ctx, f, configFH)
	if err != nil {
		return &Config{}, f, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, f, nil
}

func Read(ctx context.Context, filename string, configFH io.Reader) (*Config, error) {
	configs := []cueutils.CueConfigFile{}

	schemaFilename := "schema.cue"
	schemaFH, err := f.Open(schemaFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", schemaFilename, err)
	}
	defer schemaFH.Close()
	configs = append(configs, cueutils.CueConfigFile{Data: schemaFH, Name: schemaFilename})

	configs = append(configs, cueutils.CueConfigFile{Data: configFH, Name: filename})

	c := Config{}
	err = cueutils.Unmarshal(configs, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return &c, nil
}

type contextKey string

const configKey contextKey = "stack-config"

func NewConfigContext(ctx context.Context, value *Config) context.Context {
	return context.WithValue(ctx, configKey, value)
}

func ConfigFromContext(ctx context.Context) *Config {
	v, ok := ctx.Value(configKey).(*Config)
	if ok {
		return v
	}
	return &Config{}
}
