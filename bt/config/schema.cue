package bt

#Terraform: {
	[ID=_]: #TerraformProfile & {id: ID}
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
}

#Command: {
	name: string
	command: [...string]
}
