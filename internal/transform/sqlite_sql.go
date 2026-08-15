package transform

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
)

// sqlitePlaceholders reports the bind order of a sqlc SQLite query. The text
// is executed as sqlc emits it - the sqlite drivers bind "?" natively - so
// unlike MySQL there is nothing to rewrite, only to enumerate.
//
// SQLite's lexer differs from MySQL's: "--" needs no trailing whitespace, "#"
// is not a comment, backslashes are not escapes, "/*!" is an ordinary
// comment, and both `...` and [...] are quoted identifiers. Those identifiers
// matter: a "?" inside one is not a bind slot, and counting it as one
// misorders the arguments of a reused sqlc.slice.
func sqlitePlaceholders(sql string) []model.Placeholder {
	var slots []model.Placeholder
	for i := 0; i < len(sql); {
		c := sql[i]
		switch {
		case c == '?':
			// sqlc numbers SQLite parameters (?1, ?2); the digits belong to
			// the placeholder.
			i++
			for i < len(sql) && sql[i] >= '0' && sql[i] <= '9' {
				i++
			}
			slots = append(slots, model.Placeholder{SliceName: "", Marker: ""})
		case c == '\'' || c == '"' || c == '`':
			i = scanQuoted(sql, i, c)
		case c == '[':
			i = scanBracketIdent(sql, i)
		case c == '-' && strings.HasPrefix(sql[i:], "--"):
			i = scanLineEnd(sql, i)
		case c == '/' && strings.HasPrefix(sql[i:], sliceMarkerPrefix):
			end := scanMySQLBlockComment(sql, i)
			marker := sql[i:end]
			i = end
			if i < len(sql) && sql[i] == '?' {
				i++
				slots = append(slots, model.Placeholder{
					SliceName: sliceMarkerName(marker, len(sliceMarkerPrefix)),
					Marker:    marker + "?",
				})
			}
		case c == '/' && strings.HasPrefix(sql[i:], "/*"):
			i = scanMySQLBlockComment(sql, i)
		default:
			i++
		}
	}

	return slots
}

// scanBracketIdent returns the index after a [quoted identifier]. SQLite has
// no escape inside brackets: the first ] ends the identifier.
func scanBracketIdent(sql string, i int) int {
	end := strings.IndexByte(sql[i:], ']')
	if end == -1 {
		return len(sql)
	}

	return i + end + 1
}
