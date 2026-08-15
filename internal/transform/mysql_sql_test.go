package transform

import "testing"

func TestRewriteMySQLSQL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sql  string
		want string
	}{
		{
			name: "no placeholders or percents is unchanged",
			sql:  "SELECT id, name FROM t",
			want: "SELECT id, name FROM t",
		},
		{
			name: "single parameter at end of input",
			sql:  "SELECT id FROM t WHERE id = ?",
			want: "SELECT id FROM t WHERE id = %s",
		},
		{
			name: "multiple parameters",
			sql:  "INSERT INTO t (a, b, c) VALUES (?, ?, ?)",
			want: "INSERT INTO t (a, b, c) VALUES (%s, %s, %s)",
		},
		{
			name: "parameter at start of input",
			sql:  "? = ?",
			want: "%s = %s",
		},
		{
			name: "digits after a placeholder stay text",
			sql:  "SELECT ?1",
			want: "SELECT %s1",
		},
		{
			name: "modulo operator is doubled",
			sql:  "SELECT id % 2 FROM t WHERE id = ?",
			want: "SELECT id %% 2 FROM t WHERE id = %s",
		},
		{
			name: "literal percent in string is doubled",
			sql:  "SELECT ? WHERE note LIKE '50%' OR note LIKE 'a%b'",
			want: "SELECT %s WHERE note LIKE '50%%' OR note LIKE 'a%%b'",
		},
		{
			name: "percent in comments is doubled",
			sql:  "SELECT ? -- 50%\n# 10%\n/* 5% */",
			want: "SELECT %s -- 50%%\n# 10%%\n/* 5%% */",
		},
		{
			name: "trailing lone percent",
			sql:  "SELECT 100 %",
			want: "SELECT 100 %%",
		},
		{
			name: "placeholder inside single-quoted string stays text",
			sql:  "SELECT '?', 'it''s ?', ?",
			want: "SELECT '?', 'it''s ?', %s",
		},
		{
			name: "backslash-escaped quote keeps the string closed",
			sql:  `SELECT 'It\'s ok', ?`,
			want: `SELECT 'It\'s ok', %s`,
		},
		{
			name: "escaped backslash before closing quote",
			sql:  `SELECT 'a\\', ?`,
			want: `SELECT 'a\\', %s`,
		},
		{
			name: "placeholder inside double-quoted string stays text",
			sql:  `SELECT "?", "a""b ?", ?`,
			want: `SELECT "?", "a""b ?", %s`,
		},
		{
			name: "backslash-escaped double quote keeps the string closed",
			sql:  `SELECT "he said \" ?", ?`,
			want: `SELECT "he said \" ?", %s`,
		},
		{
			name: "backtick identifier stays text",
			sql:  "SELECT `weird?col`, `a``b ?` FROM t WHERE x = ?",
			want: "SELECT `weird?col`, `a``b ?` FROM t WHERE x = %s",
		},
		{
			name: "backslash is not an escape inside backticks",
			sql:  "SELECT `a\\` FROM t WHERE x = ?",
			want: "SELECT `a\\` FROM t WHERE x = %s",
		},
		{
			name: "double dash before digit is not a comment",
			sql:  "SELECT a--1 FROM t WHERE b = ?",
			want: "SELECT a--1 FROM t WHERE b = %s",
		},
		{
			name: "double dash before letter is not a comment",
			sql:  "SELECT a--x, ?",
			want: "SELECT a--x, %s",
		},
		{
			name: "line comment kills the first placeholder only",
			sql:  "SELECT a -- comment ?\nFROM t WHERE b = ?",
			want: "SELECT a -- comment ?\nFROM t WHERE b = %s",
		},
		{
			name: "double dash followed by tab is a comment",
			sql:  "SELECT a --\tdead ?\n, ?",
			want: "SELECT a --\tdead ?\n, %s",
		},
		{
			name: "double dash at end of input is a comment",
			sql:  "SELECT ? --",
			want: "SELECT %s --",
		},
		{
			name: "bare carriage return stays inside a line comment",
			sql:  "SELECT ? -- note ?\r, ?\nAND ?",
			want: "SELECT %s -- note ?\r, ?\nAND %s",
		},
		{
			name: "hash comment stays text",
			sql:  "#comment ?\n?",
			want: "#comment ?\n%s",
		},
		{
			name: "bare carriage return stays inside a hash comment",
			sql:  "SELECT 1 # note ?\r, ?\n? ",
			want: "SELECT 1 # note ?\r, ?\n%s ",
		},
		{
			name: "block comment stays text",
			sql:  "SELECT ? /* not ? */ FROM t",
			want: "SELECT %s /* not ? */ FROM t",
		},
		{
			name: "block comments do not nest",
			sql:  "SELECT /* a /* b */ ? */ 1",
			want: "SELECT /* a /* b */ %s */ 1",
		},
		{
			name: "version comment body is live SQL",
			sql:  "/*!40101 SET x=1, y='50%'*/ SELECT ?",
			want: "/*!40101 SET x=1, y='50%%'*/ SELECT %s",
		},
		{
			name: "placeholder inside a version comment is rewritten",
			sql:  "SELECT id FROM t /*! WHERE id = ? AND s = 'a?b' */",
			want: "SELECT id FROM t /*! WHERE id = %s AND s = 'a?b' */",
		},
		{
			name: "optimizer hint comment stays text",
			sql:  "SELECT /*+ MAX_EXECUTION_TIME(1000) ? */ id FROM t WHERE id = ?",
			want: "SELECT /*+ MAX_EXECUTION_TIME(1000) ? */ id FROM t WHERE id = %s",
		},
		{
			name: "slice marker placeholder is rewritten",
			sql:  "SELECT * FROM t WHERE id IN (/*SLICE:ids*/?)",
			want: "SELECT * FROM t WHERE id IN (/*SLICE:ids*/%s)",
		},
		{
			name: "reused slice marker",
			sql:  "SELECT * FROM t WHERE a IN (/*SLICE:ids*/?) AND b IN (/*SLICE:ids*/?)",
			want: "SELECT * FROM t WHERE a IN (/*SLICE:ids*/%s) AND b IN (/*SLICE:ids*/%s)",
		},
		{
			name: "unterminated string swallows the rest",
			sql:  "SELECT ?, 'open ?",
			want: "SELECT %s, 'open ?",
		},
		{
			name: "unterminated string with trailing backslash swallows the rest",
			sql:  `SELECT ?, 'open\`,
			want: `SELECT %s, 'open\`,
		},
		{
			name: "unterminated backtick identifier swallows the rest",
			sql:  "SELECT ?, `open ?",
			want: "SELECT %s, `open ?",
		},
		{
			name: "unterminated block comment swallows the rest",
			sql:  "SELECT ? /* dangling ?",
			want: "SELECT %s /* dangling ?",
		},
		{
			name: "multi-byte characters inside a string stay text",
			sql:  "SELECT 'entr\xc3\xa9e ?', ?",
			want: "SELECT 'entr\xc3\xa9e ?', %s",
		},
		{
			name: "slash without comment is copied",
			sql:  "SELECT ? / 2",
			want: "SELECT %s / 2",
		},
		{
			name: "single dash is copied",
			sql:  "SELECT ? - 2",
			want: "SELECT %s - 2",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := rewriteMySQLSQL(tc.sql); got != tc.want {
				t.Errorf("rewriteMySQLSQL() = %q, want %q", got, tc.want)
			}
		})
	}
}
