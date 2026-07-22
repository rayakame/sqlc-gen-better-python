package transform

import (
	"strings"
)

// rewritePsycopgSQL converts sqlc's PostgreSQL placeholders into psycopg's
// named pyformat style: $N becomes %(pN)s, and every literal % is doubled,
// since psycopg scans the whole query text for placeholders - including
// string literals and comments - once parameters are passed. String
// literals, quoted identifiers, dollar-quoted strings, and comments are
// tracked so a $N inside them stays text. Callers only rewrite queries that
// actually have parameters; psycopg leaves parameterless queries unscanned.
func rewritePsycopgSQL(sql string) string {
	var out strings.Builder
	out.Grow(len(sql) + len(sql)/8)
	for i := 0; i < len(sql); {
		c := sql[i]
		switch {
		case c == '%':
			out.WriteString("%%")
			i++
		case c == '$':
			if num, end := scanParamNumber(sql, i); end != -1 {
				out.WriteString("%(p" + num + ")s")
				i = end

				continue
			}
			if end := scanDollarQuote(sql, i); end != -1 {
				writeDoubled(&out, sql[i:end])
				i = end

				continue
			}
			out.WriteByte(c)
			i++
		case c == '\'':
			end := scanStringLiteral(sql, i, isEscapeString(sql, i))
			writeDoubled(&out, sql[i:end])
			i = end
		case c == '"':
			end := scanQuoted(sql, i, '"')
			writeDoubled(&out, sql[i:end])
			i = end
		case c == '-' && strings.HasPrefix(sql[i:], "--"):
			end := strings.IndexByte(sql[i:], '\n')
			if end == -1 {
				end = len(sql)
			} else {
				end += i + 1
			}
			writeDoubled(&out, sql[i:end])
			i = end
		case c == '/' && strings.HasPrefix(sql[i:], "/*"):
			end := scanBlockComment(sql, i)
			writeDoubled(&out, sql[i:end])
			i = end
		default:
			out.WriteByte(c)
			i++
		}
	}

	return out.String()
}

// writeDoubled copies a skipped segment, still doubling every % in it.
func writeDoubled(out *strings.Builder, segment string) {
	out.WriteString(strings.ReplaceAll(segment, "%", "%%"))
}

// scanParamNumber reads a $N parameter reference at i and returns its digits
// and the index after them, or -1 when $ does not start a parameter.
func scanParamNumber(sql string, i int) (string, int) {
	j := i + 1
	for j < len(sql) && sql[j] >= '0' && sql[j] <= '9' {
		j++
	}
	if j == i+1 {
		return "", -1
	}

	return sql[i+1 : j], j
}

// scanDollarQuote returns the index after a complete dollar-quoted string
// starting at i ($$...$$ or $tag$...$tag$), or -1 if i does not start one.
// A digit cannot start a tag - the caller already consumed $N parameters -
// and an unterminated quote swallows the rest of the input.
func scanDollarQuote(sql string, i int) int {
	j := i + 1
	if j < len(sql) && (sql[j] == '_' || isAlphaByte(sql[j])) {
		for j < len(sql) && isIdentByte(sql[j]) {
			j++
		}
	}
	if j >= len(sql) || sql[j] != '$' {
		return -1
	}
	delim := sql[i : j+1]
	end := strings.Index(sql[j+1:], delim)
	if end == -1 {
		return len(sql)
	}

	return j + 1 + end + len(delim)
}

// isEscapeString reports whether the quote at i opens an E'...' escape
// string, where backslashes escape the following character.
func isEscapeString(sql string, i int) bool {
	if i == 0 {
		return false
	}
	if sql[i-1] != 'e' && sql[i-1] != 'E' {
		return false
	}
	// The e must be its own token, not the tail of an identifier like "table".
	return i < 2 || !isIdentByte(sql[i-2])
}

// scanStringLiteral returns the index after a single-quoted literal starting
// at i, honoring ” doubling and, for escape strings, backslash escapes.
func scanStringLiteral(sql string, i int, escapes bool) int {
	j := i + 1
	for j < len(sql) {
		switch {
		case escapes && sql[j] == '\\':
			j += 2
		case sql[j] != '\'':
			j++
		case j+1 < len(sql) && sql[j+1] == '\'':
			j += 2
		default:
			return j + 1
		}
	}

	return len(sql)
}

// scanQuoted returns the index after a quoted region starting at i, where a
// doubled quote is an escape (used for "identifiers").
func scanQuoted(sql string, i int, quote byte) int {
	j := i + 1
	for j < len(sql) {
		if sql[j] != quote {
			j++

			continue
		}
		if j+1 < len(sql) && sql[j+1] == quote {
			j += 2

			continue
		}

		return j + 1
	}

	return len(sql)
}

// scanBlockComment returns the index after a /* */ comment starting at i.
// PostgreSQL block comments nest.
func scanBlockComment(sql string, i int) int {
	depth := 0
	j := i
	for j < len(sql) {
		switch {
		case strings.HasPrefix(sql[j:], "/*"):
			depth++
			j += 2
		case strings.HasPrefix(sql[j:], "*/"):
			depth--
			j += 2
			if depth == 0 {
				return j
			}
		default:
			j++
		}
	}

	return len(sql)
}

func isAlphaByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentByte(c byte) bool {
	return isAlphaByte(c) || (c >= '0' && c <= '9') || c == '_'
}
