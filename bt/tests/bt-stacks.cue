// Define the list of components
component: "networking": {}
component: "kubernetes": {
	depends_on: ["networking"]
}
component: "node_groups": {
	depends_on: ["kubernetes"]
}
component: "addons": {
	depends_on: ["kubernetes"]
}
component: "dns": {
	depends_on: ["kubernetes"]
}
component: "dev-rbac": {
	depends_on: ["kubernetes", "addons"]
}

// Create component groupings with additional variable definitions
_standard_cluster: {
	"networking": component["networking"] & {
		variables: [
			{name: "subnet_size", value: "/28"},
		]
	}
	"kubernetes": component["kubernetes"]
	"node_groups": component["node_groups"]
	"addons": component["addons"]
	"dns": component["dns"] & {
		variables: [
			{name: "api_endpoint", value: "api.example.com"},
		]
	}
}

// Create a stack with a list of components
stack: "dev-us-west-2": {
	id: string
	components: [
		for k, v in _standard_cluster {
			[// switch
				if k == "networking" {
					v & {
						workspaces: [
							"\(id)-k8s",
						]
					}
				},
				if k == "node_groups" {
					v & {
						workspaces: [
							"\(id)a",
							"\(id)b",
							"\(id)c",
						]
					}
				},
				v & {
					workspaces: [id]
				},
			][0]
		},
		// Custom component that only applies to this stack
		component["dev-rbac"] & {
			workspaces: [id]
		}
	]
}

stack: "prod-us-west-2": {
	id: string
	components: [
		for k, v in _standard_cluster {
			[// switch
				if k == "networking" {
					v & {
						workspaces: [
							"\(id)-k8s",
						]
					}
				},
				if k == "node_groups" {
					v & {
						workspaces: [
							"\(id)a",
							"\(id)b",
							"\(id)c",
						]
					}
				},
				v & {
					workspaces: [id]
				},
			][0]
		}
	]
}
