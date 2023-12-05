package config

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/cueutils"
)

//go:embed schema.cue
var f embed.FS

var Logger = log.New(os.Stderr, "", log.LstdFlags)

type Config struct {
	Terraform  map[string]*TerraformProfile `json:"terraform"`
	ConfigRoot string                       `json:"config_root"`
}

type TerraformProfile struct {
	ID   string
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
	PreApplyChecks struct {
		Enabled  bool
		Commands []Command
	} `json:"pre_apply_checks"`
	BinaryName string `json:"binary_name"`
}

type Command struct {
	Name    string
	Command []string
	Files   []string
}

func (t *TerraformProfile) String() string {

	output := fmt.Sprintf("%s backend_config files: %v, var files: %v, workspaces enabled: %t, ws dir: '%s'",
		t.BinaryName,
		t.Init.BackendConfig,
		t.Plan.VarFile,
		t.Workspaces.Enabled,
		t.Workspaces.Dir,
	)
	if t.PreApplyChecks.Enabled {
		output += ", pre_apply_checks: "
		names := []string{}
		for _, cmd := range t.PreApplyChecks.Commands {
			names = append(names, cmd.Name)
		}
		output += fmt.Sprintf("%v", names)
	}
	return output
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

	configFH, err := os.Open(f)
	if err != nil {
		return &Config{}, f, fmt.Errorf("failed to open config file '%s': %w", f, err)
	}
	defer configFH.Close()

	cfg, err := Read(ctx, f, configFH)
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
	cfg.ConfigRoot = filepath.Dir(filename)
	for k, v := range cfg.Terraform {
		if v.BinaryName == "" {
			v.BinaryName = "terraform"
			cfg.Terraform[k] = v
		}
	}
	return nil
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
