@experiment(aliasv2)
package jsql

#Provider: {
	name:       string
	schemaName: string
	getCommands: [string]: #GetCommand
	macros: [...string]
}

#GetCommand: {
	name:    string
	table:   string
	command: [...string]
	filter:  [...string]
	create:  [...string]
	args:    [...#CommandArg]
	...
}

#CommandArg: {
	name:        string
	description: string | *""
}

provider: [string]~(X,_): #Provider & {
	name:       X
	schemaName: string | *X
	getCommands: [string]~(Y,_): #GetCommand & {
		name:  Y
		table: string | *Y
	}
}
