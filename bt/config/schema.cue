package bt

#Terraform: {
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
}
