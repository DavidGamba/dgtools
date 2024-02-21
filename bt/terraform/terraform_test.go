package terraform

import (
	"bytes"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
)

func setupLogging() *bytes.Buffer {
	s := ""
	buf := bytes.NewBufferString(s)
	Logger.SetOutput(buf)
	run.Logger.SetOutput(buf)
	return buf
}

func getDefaultConfig() *config.Config {
	defaultConfig := &config.Config{}
	defaultConfig.Config.DefaultTerraformProfile = "dev"
	defaultConfig.Config.TerraformProfileEnvVar = "deployment_env"
	defaultConfig.TFProfile = map[string]config.TerraformProfile{
		"dev": {
			ID:         "dev",
			BinaryName: "tofu",
			Platforms:  []string{"darwin_amd64", "darwin_arm64", "linux_amd64", "linux_arm64"},
		},
		"prod": {
			ID:         "prod",
			BinaryName: "terraform",
			Platforms:  []string{"darwin_arm64", "linux_arm64"},
		},
	}
	dev := defaultConfig.TFProfile["dev"]
	dev.Init.BackendConfig = []string{"~/dev-credentials.json"}
	dev.Plan.VarFile = []string{"~/dev-backend-config.json"}
	dev.Workspaces.Enabled = true
	dev.Workspaces.Dir = "environments"
	defaultConfig.TFProfile["dev"] = dev

	prod := defaultConfig.TFProfile["prod"]
	prod.Init.BackendConfig = []string{"$CONFIG_ROOT/prod-credentials.json"}
	prod.Plan.VarFile = []string{"$CONFIG_ROOT/prod-backend-config.json"}
	prod.Workspaces.Enabled = true
	prod.Workspaces.Dir = "environments"
	prod.PreApplyChecks.Enabled = true
	prod.PreApplyChecks.Commands = []config.Command{{
		Name:    "conftest",
		Command: []string{"conftest", "test", "--all-namespaces", "-p", "$CONFIG_ROOT/policies", "$TERRAFORM_JSON_PLAN"},
	}}
	defaultConfig.TFProfile["prod"] = prod

	defaultConfig.ConfigRoot = "/tmp/terraform-project"
	defaultConfig.ConfigFile = "/tmp/terraform-project/.bt.cue"

	return defaultConfig
}
