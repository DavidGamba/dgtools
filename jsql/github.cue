package jsql

import (
	"strings"
)

provider: gh: {
	getCommands: repos: {
		args: [{name: "org", description: "GitHub Organization"}]
		command: [
			"gh", "repo", "list", "$arg1"
			"--limit", "10"
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
