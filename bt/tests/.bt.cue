terraform: {
	init: {
		backend_config: ["backend.tfvars"]
	}
	plan: {
		var_file: ["vars.tfvars"]
	}
	workspaces: {
		enabled: true
		dir: "envs"
	}
}
