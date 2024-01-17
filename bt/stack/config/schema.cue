package stack

#ID: string & =~"^[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$"

#Variable: {
	name: string
	value: string
}

#Component: {
	id: #ID
	path: string | *id
	depends_on: [...#Component.id]
	variables: [...#Variable]
	workspaces: [...string]
}

#Stack: {
	id: #ID
	components: [...#Component]
}

component: [ID=_]: #Component & {id: ID}

stack: [ID=_]: #Stack & {id: ID}
