@experiment(aliasv2)
package jsql

provider: csv: {
	getCommands: load: {
		args: [
			{name: "name", description: "Table name"}
			{name: "file", description: "CSV file"}
		]
		create: [
			"CREATE SCHEMA IF NOT EXISTS $arg1;"
			"DROP TABLE IF EXISTS $arg1;"
			"CREATE TABLE $arg1 AS SELECT * FROM read_csv('$arg2')"
		]
	}
}
