package transform

import (
	"strings"
)

// rewriteMySQLSQL converts sqlc's MySQL placeholders into pyformat style:
// every ? becomes %s, and every literal % is doubled, since PyMySQL and
// asyncmy interpolate the whole query text with Python %-formatting once
// parameters are passed - including string literals and comments. String
// literals, backtick identifiers, and comments are tracked so a ? inside
// them stays text. Only default sql_mode lexing is supported: sqlc's
// dolphin (TiDB) parser lexes with backslash escapes on and treats "..."
// as a string, so any query that reached the plugin already parsed under
// those rules; NO_BACKSLASH_ESCAPES and ANSI_QUOTES are deliberately
// unsupported.
func rewriteMySQLSQL(sql string) string {
	var out strings.Builder
	out.Grow(len(sql) + len(sql)/8)
	for i := 0; i < len(sql); {
		c := sql[i]
		switch {
		case c == '?':
			// MySQL has no ?N syntax (that is sqlite-only), so digits after
			// ? are ordinary text.
			out.WriteString("%s")
			i++
		case c == '%':
			out.WriteString("%%")
			i++
		case c == '\'' || c == '"':
			end := scanMySQLString(sql, i, c)
			writeDoubled(&out, sql[i:end])
			i = end
		case c == '`':
			end := scanQuoted(sql, i, '`')
			writeDoubled(&out, sql[i:end])
			i = end
		case c == '#':
			end := scanLineEnd(sql, i)
			writeDoubled(&out, sql[i:end])
			i = end
		case c == '-' && strings.HasPrefix(sql[i:], "--") && isMySQLLineComment(sql, i):
			end := scanLineEnd(sql, i)
			writeDoubled(&out, sql[i:end])
			i = end
		case c == '/' && strings.HasPrefix(sql[i:], "/*!"):
			// MySQL executes /*! version comments and sqlc's parser agrees:
			// the body is live SQL and a ? inside it is a real parameter.
			// Emit the opener and scan the body with the normal rules; the
			// closing */ falls through the default case as ordinary text.
			out.WriteString("/*!")
			i += len("/*!")
		case c == '/' && strings.HasPrefix(sql[i:], "/*"):
			end := scanMySQLBlockComment(sql, i)
			writeDoubled(&out, sql[i:end])
			i = end
		default:
			out.WriteByte(c)
			i++
		}
	}

	return out.String()
}

// isMySQLLineComment reports whether the -- at i starts a comment. MySQL
// requires the second dash to be followed by whitespace, a control
// character, or end of input; "a--1" is double unary minus, not a comment.
// When it is not a comment the caller copies the dashes as ordinary text.
// (DEL needs no arm: sqlc's parser rejects a bare 0x7f at generate time.)
func isMySQLLineComment(sql string, i int) bool {
	if i+2 >= len(sql) {
		return true
	}

	return sql[i+2] <= ' '
}

// scanLineEnd returns the index of the \n terminating a line comment at or
// after i, or end of input. MySQL and sqlc's dolphin parser end -- and #
// comments only at \n (a bare \r is comment text, unlike PostgreSQL). The
// terminator itself is not consumed; the caller copies it as ordinary text.
func scanLineEnd(sql string, i int) int {
	end := strings.IndexByte(sql[i:], '\n')
	if end == -1 {
		return len(sql)
	}

	return i + end
}

// scanMySQLString returns the index after a string literal starting at i,
// honoring backslash escapes and quote doubling. Both '...' and "..." are
// strings in default MySQL and share these rules. An unterminated literal
// swallows the rest of the input.
func scanMySQLString(sql string, i int, quote byte) int {
	j := i + 1
	for j < len(sql) {
		switch {
		case sql[j] == '\\':
			j += 2
		case sql[j] != quote:
			j++
		case j+1 < len(sql) && sql[j+1] == quote:
			j += 2
		default:
			return j + 1
		}
	}

	return len(sql)
}

// scanMySQLBlockComment returns the index after a /* */ comment starting at
// i. MySQL block comments do not nest: the first */ ends the comment. /*+
// optimizer hints and sqlc's /*SLICE:name*/ markers scan the same way; /*!
// version comments never reach here (their body is live SQL).
func scanMySQLBlockComment(sql string, i int) int {
	body := i + len("/*")
	end := strings.Index(sql[body:], "*/")
	if end == -1 {
		return len(sql)
	}

	return body + end + len("*/")
}
