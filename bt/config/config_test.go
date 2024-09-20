package config

import (
	"context"
	"strings"
	"testing"

	"github.com/DavidGamba/dgtools/cueutils"
)

func TestConfig(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		c := `
package bt

_common: {
	workspaces: {
		enabled: true
		dir: "envs"
	}
	pre_apply_checks: {
		enabled: true
		commands: [
			{name: "conftest", command: ["conftest", "test", "$TERRAFORM_JSON_PLAN"]},
		]
	}
}
config: {
	default_terraform_profile: "default"
	terraform_profile_env_var: "BT_TERRAFORM_PROFILE"
}
terraform_profile: {
	default: {
		_common
		init: {
			backend_config: ["backend.tfvars"]
		}
		plan: {
			var_file: ["vars.tfvars"]
		}
	}
	tofu: {
		_common
		binary_name: "tofu"
	}
}
`
		ctx := context.Background()
		r := strings.NewReader(c)
		cfgValue := cueutils.NewValue()
		cfg, err := Read(ctx, cfgValue, "config.cue", r)
		if err != nil {
			t.Fatalf("failed to read config: %s", err)
		}
		err = SetDefaults(ctx, cfg, "config.cue")
		if err != nil {
			t.Fatalf("failed to update config: %s", err)
		}
		if cfg.ConfigRoot != "." {
			t.Errorf("expected ConfigRoot to be '.', got '%s'", cfg.ConfigRoot)
		}
		if cfg.TFProfile["default"].BinaryName != "terraform" {
			t.Errorf("expected BinaryName to be 'terraform', got '%s'", cfg.TFProfile["default"].BinaryName)
		}
		if cfg.TFProfile["tofu"].BinaryName != "tofu" {
			t.Errorf("expected BinaryName to be 'tofu', got '%s'", cfg.TFProfile["tofu"].BinaryName)
		}
		if !cfg.TFProfile["tofu"].Workspaces.Enabled {
			t.Errorf("workspaces should be enabled for tofu")
		}
		if cfg.Config.TerraformProfileEnvVar != "BT_TERRAFORM_PROFILE" {
			t.Errorf("expected TerraformProfileEnvVar to be 'BT_TERRAFORM_PROFILE', got '%s'", cfg.Config.TerraformProfileEnvVar)
		}
		t.Logf("%#v", cfg)
		t.Logf("%#v", cfg.TFProfile["default"])
		t.Logf("%#v", cfg.TFProfile["tofu"])
	})
}
