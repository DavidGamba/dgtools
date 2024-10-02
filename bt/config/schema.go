package config

import (
	"fmt"
)

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
	PostApplyChecks struct {
		Enabled  bool
		Commands []Command
	} `json:"post_apply_checks"`
	BinaryName string   `json:"binary_name"`
	Platforms  []string `json:"platforms"`
}

type Command struct {
	Name       string
	Command    []string
	Files      []string
	OutputFile string `json:"output_file,omitempty"`
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
	if t.PostApplyChecks.Enabled {
		output += ", post_apply_checks: "
		names := []string{}
		for _, cmd := range t.PostApplyChecks.Commands {
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
