package bt

config: #Config
terraform_profile: [ID=_]: #TerraformProfile & {id: ID}

#Config: {
	default_terraform_profile: string | *"default"
	terraform_profile_env_var: string | *"BT_TERRAFORM_PROFILE"
}

#TerraformProfile: {
	id: string
	init?: {
		backend_config: [...string]
	}
	plan?: {
		var_file: [...string]
	}
	workspaces?: {
		enabled: bool
		dir: string
	}
	pre_apply_checks?: {
		enabled: bool
		commands: [...#Command]
	}
	binary_name: string | *"terraform"
	platforms: [...string]
}

#Command: {
	name: string
	command: [...string]
	files: [...string]
	output_file?: string
}
