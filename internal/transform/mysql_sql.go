package transform

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
)

// mysqlToken is the pyformat placeholder the rewriter emits for every "?".
const mysqlToken = "%s"

// sliceMarkerPrefix opens the marker sqlc leaves for a sqlc.slice parameter.
// The token binding the sequence always follows the marker immediately.
const sliceMarkerPrefix = "/*SLICE:"

// rewriteMySQLSQL converts sqlc's MySQL placeholders into pyformat style and
// reports the bind order of the text it produced:
// every ? becomes %s, and every literal % is doubled, since PyMySQL and
// asyncmy interpolate the whole query text with Python %-formatting once
// parameters are passed - including string literals and comments. String
// literals, backtick identifiers, and comments are tracked so a ? inside
// them stays text. Only default sql_mode lexing is supported: sqlc's
// dolphin (TiDB) parser lexes with backslash escapes on and treats "..."
// as a string, so any query that reached the plugin already parsed under
// those rules; NO_BACKSLASH_ESCAPES and ANSI_QUOTES are deliberately
// unsupported. Returning the placeholders here is what keeps the drivers out
// of the lexing business: they bind against this order instead of scanning
// the rewritten text a second time.
func rewriteMySQLSQL(sql string) (string, []model.Placeholder) {
	var out strings.Builder
	out.Grow(len(sql) + len(sql)/8)
	var slots []model.Placeholder
	for i := 0; i < len(sql); {
		c := sql[i]
		switch {
		case c == '?':
			// MySQL has no ?N syntax (that is sqlite-only), so digits after
			// ? are ordinary text.
			out.WriteString(mysqlToken)
			i++
			slots = append(slots, model.Placeholder{SliceName: "", Marker: ""})
		case c == '%':
			out.WriteString("%%")
			i++
		case c == '\'' || c == '"':
			end := scanEscapedString(sql, i, c)
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
		case c == '/' && strings.HasPrefix(sql[i:], sliceMarkerPrefix):
			// The marker and the token it binds are one slot: the doubled
			// marker text is what the generated code has to replace, so it
			// is carried instead of rebuilt from the raw sqlc name.
			end := scanMySQLBlockComment(sql, i)
			raw := sql[i:end]
			marker := strings.ReplaceAll(raw, "%", "%%")
			out.WriteString(marker)
			i = end
			if i < len(sql) && sql[i] == '?' {
				out.WriteString(mysqlToken)
				i++
				slots = append(slots, model.Placeholder{
					SliceName: sliceMarkerName(raw, len(sliceMarkerPrefix)),
					Marker:    marker + mysqlToken,
				})
			}
		case c == '/' && strings.HasPrefix(sql[i:], "/*"):
			end := scanMySQLBlockComment(sql, i)
			writeDoubled(&out, sql[i:end])
			i = end
		default:
			out.WriteByte(c)
			i++
		}
	}

	return out.String(), slots
}

// sliceMarkerName returns the sqlc.slice name of a marker whose body starts
// at nameStart and whose closing */ ends the given text.
func sliceMarkerName(markerText string, nameStart int) string {
	return strings.TrimSuffix(markerText[nameStart:], "*/")
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

// scanEscapedString returns the index after a string literal starting at i,
// honoring backslash escapes and quote doubling. MySQL applies these rules
// to both '...' and "..."; PostgreSQL to E'...' (via scanStringLiteral). An
// unterminated literal swallows the rest of the input.
func scanEscapedString(sql string, i int, quote byte) int {
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
