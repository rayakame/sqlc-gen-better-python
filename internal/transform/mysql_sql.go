package transform

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/sqllex"
)

// rewriteMySQLSQL converts sqlc's MySQL placeholders into pyformat style:
// every ? becomes %s, and every literal % is doubled, since PyMySQL and
// asyncmy interpolate the whole query text with Python %-formatting once
// parameters are passed - including string literals and comments. The lexing
// itself lives in internal/sqllex so the drivers, which scan the text this
// produces, cannot disagree about which "?" was bindable.
func rewriteMySQLSQL(sql string) string {
	var out strings.Builder
	out.Grow(len(sql) + len(sql)/8)
	for _, token := range sqllex.Scan(sql, sqllex.MySQLRaw) {
		switch token.Kind {
		case sqllex.KindPlaceholder:
			out.WriteString(sqllex.MySQLPyformat.Placeholder())
		case sqllex.KindSliceMarker:
			// The marker is comment text and doubles like any other; only
			// the placeholder it binds is rewritten.
			writeDoubled(&out, sql[token.Start:token.MarkerEnd])
			out.WriteString(sqllex.MySQLPyformat.Placeholder())
		case sqllex.KindText, sqllex.KindSkipped:
			writeDoubled(&out, sql[token.Start:token.End])
		}
	}

	return out.String()
}
