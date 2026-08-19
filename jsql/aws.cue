@experiment(aliasv2)
package jsql

provider: aws: {
	getCommands: ebs: {
		command: ["aws", "ec2", "describe-volumes"]
		filter:  ["qq", ".Volumes", "-o", "json"]
		create: [
			"CREATE SCHEMA IF NOT EXISTS $schemaName;"
			"DROP TABLE IF EXISTS $schemaName.$table"
			"CREATE TABLE $schemaName.$table AS SELECT * FROM '$filename'"
			"ALTER TABLE $schemaName.$table ADD COLUMN name VARCHAR;"
			"UPDATE $schemaName.$table SET name = list_filter(Tags, lambda x : x.Key = 'Name')[1].Value;"
		]

		// instances:list_transform(attachments, lambda x : x.InstanceId)[1]

	}
}
