package transform

import "testing"

func TestRewritePsycopgSQL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sql  string
		want string
	}{
		{
			name: "single parameter",
			sql:  "SELECT id FROM t WHERE id = $1",
			want: "SELECT id FROM t WHERE id = %(p1)s",
		},
		{
			name: "reused and multi-digit parameters",
			sql:  "SELECT $1, $2, $1, $12",
			want: "SELECT %(p1)s, %(p2)s, %(p1)s, %(p12)s",
		},
		{
			name: "literal percent is doubled",
			sql:  "SELECT $1 WHERE note LIKE '50%' OR note LIKE 'a%b'",
			want: "SELECT %(p1)s WHERE note LIKE '50%%' OR note LIKE 'a%%b'",
		},
		{
			name: "modulo operator is doubled",
			sql:  "SELECT id % 2 FROM t WHERE id = $1",
			want: "SELECT id %% 2 FROM t WHERE id = %(p1)s",
		},
		{
			name: "parameter inside string literal stays text",
			sql:  "SELECT '$1', 'it''s $2', $3",
			want: "SELECT '$1', 'it''s $2', %(p3)s",
		},
		{
			name: "escape string with escaped quote stays closed",
			sql:  `SELECT E'a\'b $1 %', $2`,
			want: `SELECT E'a\'b $1 %%', %(p2)s`,
		},
		{
			name: "escape string honors quote doubling too",
			sql:  "SELECT E'it''s $1', $2",
			want: "SELECT E'it''s $1', %(p2)s",
		},
		{
			name: "identifier ending in e does not start an escape string",
			sql:  "SELECT note FROM t WHERE note LIKE'%' AND id = $1",
			want: "SELECT note FROM t WHERE note LIKE'%%' AND id = %(p1)s",
		},
		{
			name: "escape string at start of input",
			sql:  "'a' || $1",
			want: "'a' || %(p1)s",
		},
		{
			name: "quoted identifier stays text",
			sql:  `SELECT "weird$1col", "a""b" FROM t WHERE x = $1`,
			want: `SELECT "weird$1col", "a""b" FROM t WHERE x = %(p1)s`,
		},
		{
			name: "dollar quoted string stays text with percents doubled",
			sql:  "SELECT $$raw $1 50%$$, $tag$ $2 $tag$, $3",
			want: "SELECT $$raw $1 50%%$$, $tag$ $2 $tag$, %(p3)s",
		},
		{
			name: "bare dollar and invalid tag are copied",
			sql:  "SELECT 1 $ 2 $abc, $1",
			want: "SELECT 1 $ 2 $abc, %(p1)s",
		},
		{
			name: "identifier containing dollar-digits stays text",
			sql:  "SELECT col$2 FROM t WHERE id = $1",
			want: "SELECT col$2 FROM t WHERE id = %(p1)s",
		},
		{
			name: "identifier with dollar-tag shaped tail stays text",
			sql:  "SELECT a$x$ FROM t WHERE id = $1",
			want: "SELECT a$x$ FROM t WHERE id = %(p1)s",
		},
		{
			name: "multi-byte dollar quote tag stays text",
			sql:  "SELECT $\xc3\xa9$50% $1$\xc3\xa9$, $2",
			want: "SELECT $\xc3\xa9$50%% $1$\xc3\xa9$, %(p2)s",
		},
		{
			name: "multi-byte identifier before string is not an escape string",
			sql:  "SELECT entr\xc3\xa9e'\\' AS x, $1",
			want: "SELECT entr\xc3\xa9e'\\' AS x, %(p1)s",
		},
		{
			name: "line comment stays text",
			sql:  "SELECT $1 -- not $2 or 50%\nFROM t",
			want: "SELECT %(p1)s -- not $2 or 50%%\nFROM t",
		},
		{
			name: "carriage return ends a line comment",
			sql:  "SELECT $1 -- note\r, $2",
			want: "SELECT %(p1)s -- note\r, %(p2)s",
		},
		{
			name: "trailing line comment without newline",
			sql:  "SELECT $1 -- 50%",
			want: "SELECT %(p1)s -- 50%%",
		},
		{
			name: "nested block comment stays text",
			sql:  "SELECT $1 /* outer /* $2 50% */ still */ FROM t",
			want: "SELECT %(p1)s /* outer /* $2 50%% */ still */ FROM t",
		},
		{
			name: "unterminated block comment swallows the rest",
			sql:  "SELECT $1 /* dangling $2",
			want: "SELECT %(p1)s /* dangling $2",
		},
		{
			name: "unterminated string swallows the rest",
			sql:  "SELECT $1, 'open $2",
			want: "SELECT %(p1)s, 'open $2",
		},
		{
			name: "unterminated escape string swallows the rest",
			sql:  `SELECT $1, E'open\'`,
			want: `SELECT %(p1)s, E'open\'`,
		},
		{
			name: "unterminated quoted identifier swallows the rest",
			sql:  `SELECT $1, "open $2`,
			want: `SELECT %(p1)s, "open $2`,
		},
		{
			name: "unterminated dollar quote swallows the rest",
			sql:  "SELECT $$open $1",
			want: "SELECT $$open $1",
		},
		{
			name: "dollar at end of input",
			sql:  "SELECT 1 $",
			want: "SELECT 1 $",
		},
		{
			name: "slash without comment is copied",
			sql:  "SELECT $1 / 2",
			want: "SELECT %(p1)s / 2",
		},
		{
			name: "dash without comment is copied",
			sql:  "SELECT $1 - 2",
			want: "SELECT %(p1)s - 2",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := rewritePsycopgSQL(tc.sql); got != tc.want {
				t.Errorf("rewritePsycopgSQL() = %q, want %q", got, tc.want)
			}
		})
	}
}
