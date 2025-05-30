package config

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"github.com/DavidGamba/dgtools/buildutils"
	"github.com/DavidGamba/dgtools/cueutils"
)

//go:embed schema.cue
var f embed.FS

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func (c Config) Profile(profile string) string {
	_, ok := c.TFProfile[profile]
	if !ok {
		return c.Config.DefaultTerraformProfile
	}
	return profile
}

type contextKey string

const configKey contextKey = "config"

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

func Get(ctx context.Context, value *cue.Value, filename string) (*Config, string, error) {
	f, err := buildutils.FindFileUpwards(ctx, filename)
	if err != nil {
		cfg := &Config{
			TFProfile: map[string]TerraformProfile{
				"default": {
					ID:         "default",
					BinaryName: "terraform",
				},
			},
		}
		cfg.Config.DefaultTerraformProfile = "default"
		SetDefaults(ctx, cfg, "")
		return cfg, f, fmt.Errorf("failed to find config file: %w", err)
	}

	configFH, err := os.Open(f)
	if err != nil {
		return &Config{}, f, fmt.Errorf("failed to open config file '%s': %w", f, err)
	}
	defer configFH.Close()

	cfg, err := Read(ctx, value, f, configFH)
	if err != nil {
		return &Config{}, f, fmt.Errorf("failed to read config: %w", err)
	}
	err = SetDefaults(ctx, cfg, f)
	if err != nil {
		return &Config{}, f, fmt.Errorf("failed to set config defaults: %w", err)
	}

	return cfg, f, nil
}

func SetDefaults(ctx context.Context, cfg *Config, filename string) error {
	if filename == "" {
		cfg.ConfigRoot = filepath.Dir(".")
		return nil
	}
	cfg.ConfigRoot = filepath.Dir(filename)
	cfg.ConfigFile = filename
	if len(cfg.TFProfile) == 0 {
		cfg.TFProfile = map[string]TerraformProfile{
			"default": {
				ID:         "default",
				BinaryName: "terraform",
			},
		}
		cfg.Config.DefaultTerraformProfile = "default"
	}
	return nil
}

func Read(ctx context.Context, value *cue.Value, filename string, configFH io.Reader) (*Config, error) {
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
	err = cueutils.Unmarshal(configs, "", "bt", "bt.cue", value, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return &c, nil
}
