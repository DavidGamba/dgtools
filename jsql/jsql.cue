@experiment(aliasv2)
package jsql

import (
	"strings"
)

provider: k8s: {
	getCommands: {
		let X = self
		"get": {
			args:    [{name: "resource", description: "Kubernetes Resource"}]
			table:   "$arg1"
			command: ["kubectl", "get", "-o", "json", "$arg1"]
			post:    ["qq", ".items", "-o", "json"]
			create: [
				"CREATE SCHEMA IF NOT EXISTS $schemaName;"
				"DROP TABLE IF EXISTS $schemaName.$table"
				"CREATE TABLE $schemaName.$table AS SELECT * FROM '$filename'"
				"ALTER TABLE $schemaName.$table ADD COLUMN name VARCHAR;"
			]
		}
		"getAll": {
			command: ["kubectl", "get", "-o", "json", "-A", "$arg1"]
			table:   X.get.table
			args:    X.get.args
			post:    X.get.post
			create:  X.get.create
		}
	}
}

// provider: aws: {
// 	getCommands: [
// 		{
// 			name:    "ebs"
// 			command: ["aws", "aws", "-o", "json", "$arg1"]
// 			post:    ["qq", ".items", "-o", "json"]
// 			create: [
// 				["CREATE SCHEMA IF NOT EXISTS aws;"],
// 				["DROP TABLE IF EXISTS $arg1"],
// 				["CREATE TABLE $arg1 AS SELECT * FROM '$filename'"],
// 			]
// 			update: [
// 				["ALTER TABLE $arg1 ADD COLUMN name VARCHAR;"]
// 			]
// 		}
// 	]
// }

provider: gh: {
	getCommands: repos: {
		args: [{name: "org", description: "GitHub Organization"}]
		command: [
			"gh", "repo", "list", "$arg1"
			"--limit", "100"
			"--json", strings.Join([
				"archivedAt"
				// "assignableUsers"
				// "codeOfConduct"
				// "contactLinks"
				"createdAt"
				"defaultBranchRef"
				// "deleteBranchOnMerge"
				"description"
				"diskUsage"
				"forkCount"
				// "fundingLinks"
				// "hasDiscussionsEnabled"
				// "hasIssuesEnabled"
				// "hasProjectsEnabled"
				// "hasWikiEnabled"
				// "homepageUrl"
				"id"
				"isArchived"
				// "isBlankIssuesEnabled"
				"isEmpty"
				"isFork"
				"isInOrganization"
				"isMirror"
				"isPrivate"
				"isSecurityPolicyEnabled"
				"isTemplate"
				"isUserConfigurationRepository"
				// "issueTemplates"
				// "issues"
				// "labels"
				// "languages"
				"latestRelease"
				"licenseInfo"
				// "mentionableUsers"
				// "mergeCommitAllowed"
				// "milestones"
				"mirrorUrl"
				"name"
				"nameWithOwner"
				// "openGraphImageUrl"
				"owner"
				// "parent"
				"primaryLanguage"
				// "pullRequestTemplates"
				"pullRequests"
				"pushedAt"
				// "rebaseMergeAllowed"
				"repositoryTopics"
				// "squashMergeAllowed"
				"sshUrl"
				// "stargazerCount"
				// "templateRepository"
				"updatedAt"
				"url"
				// "usesCustomOpenGraphImage"
				// "viewerCanAdminister"
				// "viewerDefaultCommitEmail"
				// "viewerDefaultMergeMethod"
				// "viewerHasStarred"
				// "viewerPermission"
				// "viewerPossibleCommitEmails"
				// "viewerSubscription"
				"visibility"
				// "watchers"
			], ",")
			// "projects" // deprecated
			// "projectsV2" // requires extra perms
		]
		create: [
			"CREATE SCHEMA IF NOT EXISTS $schemaName;"
			"DROP TABLE IF EXISTS $schemaName.$table"
			"CREATE TABLE $schemaName.$table AS SELECT * FROM '$filename'"
		]
	}
}
