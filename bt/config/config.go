package config

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/buildutils"
	"github.com/DavidGamba/dgtools/cueutils"
)

//go:embed schema.cue
var f embed.FS

var Logger = log.New(os.Stderr, "", log.LstdFlags)

type Config struct {
	Config struct {
		DefaultTerraformProfile string `json:"default_terraform_profile"`
		TerraformProfileEnvVar  string `json:"terraform_profile_env_var"`
	} `json:"config"`
	TFProfile  map[string]TerraformProfile `json:"terraform_profile"`
	ConfigRoot string                      `json:"config_root"`
	ConfigFile string                      `json:"config_file"`
}

type TerraformProfile struct {
	ID   string `json:"id"`
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
	BinaryName string   `json:"binary_name"`
	Platforms  []string `json:"platforms"`
}

type Command struct {
	Name    string
	Command []string
	Files   []string
}

func (t TerraformProfile) String() string {
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

func (c Config) String() string {
	output := ""
	output += fmt.Sprintf("config_root: %s\n", c.ConfigRoot)
	output += fmt.Sprintf("default_terraform_profile: %s\n", c.Config.DefaultTerraformProfile)
	output += fmt.Sprintf("terraform_profile_env_var: %s\n", c.Config.TerraformProfileEnvVar)

	for k := range c.TFProfile {
		output += fmt.Sprintf("profile '%s': %s\n", k, c.TFProfile[k])
	}

	return output
}

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

func Get(ctx context.Context, filename string) (*Config, string, error) {
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
	if filename == "" {
		cfg.ConfigRoot = filepath.Dir(".")
		return nil
	}
	cfg.ConfigRoot = filepath.Dir(filename)
	cfg.ConfigFile = filename
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
