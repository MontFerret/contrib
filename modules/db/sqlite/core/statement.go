package core

import "strings"

func isInsertStatement(sqlText string) bool {
	switch firstSQLStatementKeyword(sqlText) {
	case "INSERT", "REPLACE":
		return true
	default:
		return false
	}
}

func firstSQLStatementKeyword(sqlText string) string {
	idx := skipSQLSpaceAndComments(sqlText, 0)
	keyword, next := readSQLKeyword(sqlText, idx)
	if keyword != "WITH" {
		return keyword
	}

	return keywordAfterWith(sqlText, next)
}

func keywordAfterWith(sqlText string, idx int) string {
	idx = skipSQLSpaceAndComments(sqlText, idx)
	keyword, next := readSQLKeyword(sqlText, idx)
	if keyword == "RECURSIVE" {
		idx = next
	}

	for {
		idx = skipSQLName(sqlText, skipSQLSpaceAndComments(sqlText, idx))
		if idx < 0 {
			return "WITH"
		}

		idx = skipSQLSpaceAndComments(sqlText, idx)
		if idx < len(sqlText) && sqlText[idx] == '(' {
			idx = skipSQLBalanced(sqlText, idx)
		}

		idx = skipSQLSpaceAndComments(sqlText, idx)
		keyword, next := readSQLKeyword(sqlText, idx)
		if keyword != "AS" {
			return "WITH"
		}

		idx = skipSQLMaterialization(sqlText, next)
		idx = skipSQLSpaceAndComments(sqlText, idx)
		if idx >= len(sqlText) || sqlText[idx] != '(' {
			return "WITH"
		}

		idx = skipSQLBalanced(sqlText, idx)
		idx = skipSQLSpaceAndComments(sqlText, idx)
		if idx < len(sqlText) && sqlText[idx] == ',' {
			idx++
			continue
		}

		keyword, _ = readSQLKeyword(sqlText, idx)
		return keyword
	}
}

func skipSQLMaterialization(sqlText string, idx int) int {
	idx = skipSQLSpaceAndComments(sqlText, idx)
	keyword, next := readSQLKeyword(sqlText, idx)
	switch keyword {
	case "MATERIALIZED":
		return next
	case "NOT":
		afterNot := skipSQLSpaceAndComments(sqlText, next)
		if nextKeyword, afterMaterialized := readSQLKeyword(sqlText, afterNot); nextKeyword == "MATERIALIZED" {
			return afterMaterialized
		}
	}

	return idx
}

func skipSQLName(sqlText string, idx int) int {
	if idx >= len(sqlText) {
		return -1
	}

	switch sqlText[idx] {
	case '"', '\'', '`':
		return skipSQLQuoted(sqlText, idx)
	case '[':
		return skipSQLBracketQuoted(sqlText, idx)
	default:
		_, next := readSQLKeyword(sqlText, idx)
		if next == idx {
			return -1
		}

		return next
	}
}

func skipSQLBalanced(sqlText string, idx int) int {
	depth := 0
	for idx < len(sqlText) {
		idx = skipSQLSpaceAndComments(sqlText, idx)
		if idx >= len(sqlText) {
			return idx
		}

		switch sqlText[idx] {
		case '"', '\'', '`':
			idx = skipSQLQuoted(sqlText, idx)
			continue
		case '[':
			idx = skipSQLBracketQuoted(sqlText, idx)
			continue
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return idx + 1
			}
		}

		idx++
	}

	return idx
}

func skipSQLSpaceAndComments(sqlText string, idx int) int {
	for idx < len(sqlText) {
		switch sqlText[idx] {
		case ' ', '\t', '\n', '\r', '\f', '\v':
			idx++
			continue
		}

		if strings.HasPrefix(sqlText[idx:], "--") {
			idx += 2
			for idx < len(sqlText) && sqlText[idx] != '\n' && sqlText[idx] != '\r' {
				idx++
			}
			continue
		}

		if strings.HasPrefix(sqlText[idx:], "/*") {
			idx += 2
			for idx+1 < len(sqlText) && !(sqlText[idx] == '*' && sqlText[idx+1] == '/') {
				idx++
			}
			if idx+1 < len(sqlText) {
				idx += 2
			}
			continue
		}

		break
	}

	return idx
}

func skipSQLQuoted(sqlText string, idx int) int {
	quote := sqlText[idx]
	idx++
	for idx < len(sqlText) {
		if sqlText[idx] != quote {
			idx++
			continue
		}
		if idx+1 < len(sqlText) && sqlText[idx+1] == quote {
			idx += 2
			continue
		}

		return idx + 1
	}

	return idx
}

func skipSQLBracketQuoted(sqlText string, idx int) int {
	idx++
	for idx < len(sqlText) {
		if sqlText[idx] == ']' {
			return idx + 1
		}
		idx++
	}

	return idx
}

func readSQLKeyword(sqlText string, idx int) (string, int) {
	if idx >= len(sqlText) || !isSQLIdentifierStart(sqlText[idx]) {
		return "", idx
	}

	start := idx
	idx++
	for idx < len(sqlText) && isSQLIdentifierPart(sqlText[idx]) {
		idx++
	}

	return strings.ToUpper(sqlText[start:idx]), idx
}

func isSQLIdentifierStart(ch byte) bool {
	return ch == '_' || ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z')
}

func isSQLIdentifierPart(ch byte) bool {
	return isSQLIdentifierStart(ch) || ('0' <= ch && ch <= '9')
}
