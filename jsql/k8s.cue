@experiment(aliasv2)
package jsql

provider: k8s: {
	getCommands: {
		let X = self
		"get": {
			args:    [{name: "resource", description: "Kubernetes Resource"}]
			table:   "$arg1"
			command: ["kubectl", "get", "-o", "json", "$arg1"]
			filter:  ["qq", ".items", "-o", "json"]
			create: [
				"CREATE SCHEMA IF NOT EXISTS $schemaName;"
				"DROP TABLE IF EXISTS $schemaName.$table;"
				"CREATE TABLE $schemaName.$table AS SELECT * FROM read_json('$filename');"
				"ALTER TABLE $schemaName.$table ADD COLUMN name VARCHAR;"
				"UPDATE $schemaName.$table SET name = metadata.name;"
				"ALTER TABLE $schemaName.$table ADD COLUMN namespace VARCHAR;"
				// Use cast to allow for null values
				"UPDATE $schemaName.$table SET namespace = CAST(metadata AS JSON)->>'namespace';"
			]
		}
		"getAll": {
			args:    X.get.args
			table:   X.get.table
			command: ["kubectl", "get", "-o", "json", "-A", "$arg1"]
			filter:  X.get.filter
			create:  X.get.create
		}
	}

	macros: [

		"CREATE SCHEMA IF NOT EXISTS k8s;"

		"""
			CREATE OR REPLACE MACRO k8s.age(x) AS
			CASE
				WHEN x IS NULL THEN NULL
				WHEN date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC') >= (24*60) THEN
					(date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC')//(24*60))::VARCHAR || 'd' ||
					(date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC')%(60*24)//60)::VARCHAR || 'h'
				WHEN date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC') >= 60 THEN
					(date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC')%(60*24)//60)::VARCHAR || 'h' ||
					(date_diff('minutes', x, current_timestamp AT TIME ZONE 'UTC')%60)::VARCHAR || 'm'
				ELSE
					(date_diff('seconds', x, current_timestamp AT TIME ZONE 'UTC')//60)::VARCHAR || 'm' ||
					(date_diff('seconds', x, current_timestamp AT TIME ZONE 'UTC')%60)::VARCHAR || 's'
			END;
			"""

		"CREATE OR REPLACE MACRO k8s.agem(x) AS k8s.age(x.metadata.creationTimestamp);"

		"CREATE OR REPLACE MACRO k8s.agema(x_metadata) AS k8s.age(x_metadata.creationTimestamp);"

		"""
			CREATE OR REPLACE MACRO k8s.cpu_m(cpu) AS
			CASE
				WHEN trim(cpu) LIKE '%m' THEN
					CAST(CEILING(CAST(RTRIM(trim(cpu), 'm') AS NUMERIC)) AS BIGINT)
				ELSE
					CAST(CAST(trim(cpu) AS NUMERIC) * 1000 AS BIGINT)
			END;
			"""

		"""
			CREATE OR REPLACE MACRO k8s.memory_bytes(memory) AS
			CAST(CASE
				WHEN trim(memory) LIKE '%Ki' THEN
					CAST(RTRIM(trim(memory), 'Ki') AS NUMERIC) * 1024
				WHEN trim(memory) LIKE '%Mi' THEN
					CAST(RTRIM(trim(memory), 'Mi') AS NUMERIC) * 1048576
				WHEN trim(memory) LIKE '%Gi' THEN
					CAST(RTRIM(trim(memory), 'Gi') AS NUMERIC) * 1073741824
				WHEN trim(memory) LIKE '%Ti' THEN
					CAST(RTRIM(trim(memory), 'Ti') AS NUMERIC) * 1099511627776
				WHEN trim(memory) LIKE '%Pi' THEN
					CAST(RTRIM(trim(memory), 'Pi') AS NUMERIC) * 1125899906842624
				WHEN trim(memory) LIKE '%Ei' THEN
					CAST(RTRIM(trim(memory), 'Ei') AS NUMERIC) * 1152921504606846976
				WHEN trim(memory) LIKE '%m' THEN
					CEILING(CAST(RTRIM(trim(memory), 'm') AS NUMERIC) / 1000)
				WHEN trim(memory) LIKE '%k' THEN
					CAST(RTRIM(trim(memory), 'k') AS NUMERIC) * 1000
				WHEN trim(memory) LIKE '%M' THEN
					CAST(RTRIM(trim(memory), 'M') AS NUMERIC) * 1000000
				WHEN trim(memory) LIKE '%G' THEN
					CAST(RTRIM(trim(memory), 'G') AS NUMERIC) * 1000000000
				WHEN trim(memory) LIKE '%T' THEN
					CAST(RTRIM(trim(memory), 'T') AS NUMERIC) * 1000000000000
				WHEN trim(memory) LIKE '%P' THEN
					CAST(RTRIM(trim(memory), 'P') AS NUMERIC) * 1000000000000000
				WHEN trim(memory) LIKE '%E' THEN
					CAST(RTRIM(trim(memory), 'E') AS NUMERIC) * 1000000000000000000
				ELSE
					CAST(trim(memory) AS NUMERIC)
			END AS BIGINT);
			"""

		"""
			CREATE OR REPLACE MACRO k8s.memory_human(bytes) AS
			CASE
				WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 6) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 6)), '.') || 'Ei'
				WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 5) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 5)), '.') || 'Pi'
				WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 4) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 4)), '.') || 'Ti'
				WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 3) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 3)), '.') || 'Gi'
				WHEN CAST(bytes AS DOUBLE) >= (1024::DOUBLE ^ 2) THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / (1024::DOUBLE ^ 2)), '.') || 'Mi'
				WHEN CAST(bytes AS DOUBLE) >= 1024               THEN rtrim(printf('%.1f', CAST(bytes AS DOUBLE) / 1024), '.') || 'Ki'
				ELSE CAST(bytes AS VARCHAR)
			END;
			"""

	]

	views: common: {
		dependencies: ["pods", "rs", "deploy", "sts", "nodes"]
		queries: [

			"""
				CREATE OR REPLACE VIEW drspn AS
				SELECT
					d_kind:d.kind,
					d_apiVersion:d.apiVersion,
					d_name:d.name,
					d_namespace:d.namespace,
					d_metadata:d.metadata,
					d_spec:d.spec,
					d_status:d.status,
					rs_kind:rs.kind,
					rs_apiVersion:rs.apiVersion,
					rs_name:rs.name,
					rs_namespace:rs.namespace,
					rs_metadata:rs.metadata,
					rs_spec:rs.spec,
					rs_status:rs.status,
					p_kind:p.kind,
					p_apiVersion:p.apiVersion,
					p_name:p.name,
					p_namespace:p.namespace,
					p_metadata:p.metadata,
					p_spec:p.spec,
					p_status:p.status,
					n_kind:n.kind,
					n_apiVersion:n.apiVersion,
					n_name:n.name,
					n_namespace:n.namespace,
					n_metadata:n.metadata,
					n_spec:n.spec,
					n_status:n.status
				FROM k8s.deploy AS d
				JOIN k8s.rs ON k8s.rs.namespace = d.namespace AND k8s.rs.metadata.ownerReferences[1].uid = d.metadata.uid
				JOIN k8s.pods AS p ON p.namespace = k8s.rs.namespace AND k8s.rs.metadata.uid = p.metadata.ownerReferences[1].uid
				LEFT OUTER JOIN k8s.nodes as n ON n.name = p.spec.nodeName
				;
				"""

			"""
				CREATE OR REPLACE VIEW spn AS
				SELECT
					s_kind:s.kind,
					s_apiVersion:s.apiVersion,
					s_name:s.name,
					s_namespace:s.namespace,
					s_metadata:s.metadata,
					s_spec:s.spec,
					s_status:s.status,
					p_kind:p.kind,
					p_apiVersion:p.apiVersion,
					p_name:p.name,
					p_namespace:p.namespace,
					p_metadata:p.metadata,
					p_spec:p.spec,
					p_status:p.status,
					n_kind:n.kind,
					n_apiVersion:n.apiVersion,
					n_name:n.name,
					n_namespace:n.namespace,
					n_metadata:n.metadata,
					n_spec:n.spec,
					n_status:n.status
				FROM k8s.sts AS s
				JOIN k8s.pods AS p ON p.namespace = s.namespace AND s.metadata.uid = p.metadata.ownerReferences[1].uid
				LEFT OUTER JOIN k8s.nodes as n ON n.name = p.spec.nodeName
				;
				"""

			"""
				CREATE OR REPLACE VIEW pn AS
				SELECT
					p_kind:p.kind,
					p_apiVersion:p.apiVersion,
					p_name:p.name,
					p_namespace:p.namespace,
					p_metadata:p.metadata,
					p_spec:p.spec,
					p_status:p.status,
					n_kind:n.kind,
					n_apiVersion:n.apiVersion,
					n_name:n.name,
					n_namespace:n.namespace,
					n_metadata:n.metadata,
					n_spec:n.spec,
					n_status:n.status
				FROM k8s.nodes AS n
				JOIN k8s.pods as p ON n.name = p.spec.nodeName
				;
				"""

		]
	}
}
