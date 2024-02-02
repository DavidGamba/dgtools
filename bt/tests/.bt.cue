config: {
	default_terraform_profile: "default"
	terraform_profile_env_var: "BT_TERRAFORM_PROFILE"
}
terraform_profile: {
	default: {
		binary_name: "terraform"
		init: {
			backend_config: ["backend.tfvars"]
		}
		plan: {
			var_file: ["~/auth.tfvars"]
		}
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
		platforms: ["darwin_amd64", "darwin_arm64", "linux_amd64", "linux_arm64"]
	}
}
