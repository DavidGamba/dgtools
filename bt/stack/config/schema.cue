@experiment(aliasv2,explicitopen,try)
package bt_stacks

#ID: string & =~"^[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$"

#Variable: {
	name:  string
	value: string
}

#Component: {
	id:   #ID
	path: string | *id
	depends_on: [...#Component.id]
	variables: [...#Variable]
	workspaces: [...string]
	retries: int | *0
}

#Stack: {
	id: #ID
	components: [...#Component]
}

component: [string]~(ID,_): #Component & {id: ID}

stack: [string]~(ID,_): #Stack & {id: ID}
