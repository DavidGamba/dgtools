package config

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/cueutils"
)

//go:embed schema.cue
var f embed.FS

var Logger = log.New(os.Stderr, "", log.LstdFlags)

type Config struct {
	Terraform struct {
		Init struct {
			BackendConfig []string `json:"backend_config"`
		}
		Plan struct {
			VarFile []string `json:"var_file"`
		}
		Workspaces struct {
			Enabled bool
			Dir     string
		}
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("backend_config files: %v, var files: %v, workspaces enabled: %t, ws dir: '%s'",
		c.Terraform.Init.BackendConfig,
		c.Terraform.Plan.VarFile,
		c.Terraform.Workspaces.Enabled,
		c.Terraform.Workspaces.Dir,
	)
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

func Get(ctx context.Context, filename string) (*Config, string, error) {
	f, err := FindFileUpwards(ctx, filename)
	if err != nil {
		return &Config{}, f, fmt.Errorf("failed to find config file: %w", err)
	}
	cfg, err := Read(ctx, f)
	if err != nil {
		return &Config{}, f, fmt.Errorf("failed to read config: %w", err)
	}
	return cfg, f, nil
}

func Read(ctx context.Context, filename string) (*Config, error) {
	configs := []cueutils.CueConfigFile{}

	schemaFilename := "schema.cue"
	schemaFH, err := f.Open(schemaFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", schemaFilename, err)
	}
	defer schemaFH.Close()
	configs = append(configs, cueutils.CueConfigFile{Data: schemaFH, Name: schemaFilename})

	configFH, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", filename, err)
	}
	defer configFH.Close()
	configs = append(configs, cueutils.CueConfigFile{Data: configFH, Name: filename})

	c := Config{}
	err = cueutils.Unmarshal(configs, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return &c, nil
}

func FindFileUpwards(ctx context.Context, filename string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get cwd: %w", err)
	}
	check := func(dir string) bool {
		f := filepath.Join(dir, filename)
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return false
		}
		return true
	}
	d := cwd
	for {
		found := check(d)
		if found {
			return filepath.Join(d, filename), nil
		}
		a, err := filepath.Abs(d)
		if err != nil {
			return "", fmt.Errorf("failed to get abs path: %w", err)
		}
		if a == "/" {
			break
		}
		d = filepath.Join(d, "../")
	}

	return "", fmt.Errorf("not found: %s", filename)
}
