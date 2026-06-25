package main

import (
	"slices"
	"strings"
)

var (
	commands = []string{

		"ANALYZE",
		"ALTER",
		"CALL",
		"CHECKPOINT",
		"COPY",
		"CREATE",
		"DELETE",
		"DESCRIBE",
		"DROP",
		"EXPORT",
		"IMPORT",
		"DATABASE",
		"DATABASES",
		"INSERT",
		"LOAD",
		"INSTALL",
		"MERGE",
		"INTO",
		"PIVOT",
		"SELECT",
		"RESET",
		"SET",
		"VARIABLE",
		"SHOW",
		"SUMMARIZE",
		"UNPIVOT",
		"UPDATE",
		"USE",
		"VACUUM",

		".help",
		".mode",
		".output",
	}

	createKeywords = []string{
		"INDEX",
		"MACRO",
		"SCHEMA",
		"SECRET",
		"SEQUENCE",
		"TABLE",
		"VIEW",
		"TYPE",
	}

	selectKeywords = []string{
		"ALL",
		"AND",
		"ANY",
		"ARRAY",
		"AS",
		"ASC",
		"BINARY",
		"BOTH",
		"BY",
		"CASE",
		"CAST",
		"COLLATE",
		"COLUMN",
		"COLUMNS",
		"CROSS",
		"CURRENT_DATE",
		"CURRENT_TIME",
		"CURRENT_TIMESTAMP",
		"DATABASES",
		"DEFAULT",
		"DESC",
		"DISTINCT",
		"DO",
		"FROM",
		"GROUP",
		"HAVING",
		"INTO",
		"IS",
		"JOIN",
		"LIKE",
		"LIMIT",
		"MERGE",
		"NOT",
		"ORDER",
		"QUALIFY",
		"SAMPLE",
		"USING",
		"WHERE",
		"WINDOW",
	}
)

func AllKeywords() []string {
	keywords := append(commands, createKeywords...)
	keywords = append(keywords, selectKeywords...)
	slices.Sort(keywords)
	return slices.Compact(keywords)
}

func completionCandidates(fieldsBeforeCursor []string) (completionSet []string, listingSet []string) {
	candidates := commands
	for _, word := range fieldsBeforeCursor {
		if strings.EqualFold(word, ".mode") {
			candidates = []string{"pretty", "single_line", "table", "csv"}
		}
		if strings.EqualFold(word, ".output") {
			candidates = []string{"stdout", "file"}
		}
		if strings.EqualFold(word, "SELECT") {
			candidates = append(candidates, selectKeywords...)
		}
		if strings.EqualFold(word, "CREATE") {
			candidates = append(candidates, createKeywords...)
		}
	}
	return candidates, candidates
}
