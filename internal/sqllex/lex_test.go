package sqllex_test

import (
	"reflect"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/sqllex"
)

// plain is a bind slot that is not a sqlc.slice marker.
var plain = sqllex.Slot{Name: "", Marker: ""} //nolint:gochecknoglobals

func TestSlotsMySQLRaw(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sql  string
		want []sqllex.Slot
	}{
		{name: "no placeholders", sql: "SELECT 1"},
		{
			name: "plain placeholders in text order",
			sql:  "SELECT a FROM t WHERE b = ? AND c = ?",
			want: []sqllex.Slot{plain, plain},
		},
		{
			name: "slice marker carries its text",
			sql:  "SELECT a FROM t WHERE id IN (/*SLICE:ids*/?)",
			want: []sqllex.Slot{{Name: "ids", Marker: "/*SLICE:ids*/?"}},
		},
		{
			// The ordering case the drivers depend on.
			name: "reused marker interleaved with a plain slot",
			sql:  "SELECT a FROM t WHERE x IN (/*SLICE:ids*/?) AND y = ? OR z IN (/*SLICE:ids*/?)",
			want: []sqllex.Slot{
				{Name: "ids", Marker: "/*SLICE:ids*/?"},
				plain,
				{Name: "ids", Marker: "/*SLICE:ids*/?"},
			},
		},
		{
			name: "placeholders inside strings are not slots",
			sql:  "SELECT '?', \"?\", `we?rd` FROM t WHERE a = ?",
			want: []sqllex.Slot{plain},
		},
		{
			name: "backslash escapes a quote",
			sql:  "SELECT 'a\\'?b' FROM t WHERE a = ?",
			want: []sqllex.Slot{plain},
		},
		{
			name: "doubled quotes and backticks escape",
			sql:  "SELECT 'a''?b', `bt``q?` FROM t WHERE a = ?",
			want: []sqllex.Slot{plain},
		},
		{
			name: "placeholders inside comments are not slots",
			sql:  "SELECT a -- ?\n# ?\n/* ? */ FROM t WHERE b = ?",
			want: []sqllex.Slot{plain},
		},
		{
			// "--x" is arithmetic, so the rest of the line stays live SQL.
			name: "dash run without a gap keeps its slot",
			sql:  "SELECT a FROM t WHERE b = 5--? AND c = ?",
			want: []sqllex.Slot{plain, plain},
		},
		{
			name: "odd dash run still starts a comment",
			sql:  "SELECT a FROM t WHERE b = ? --------- don't edit\nAND c = ?",
			want: []sqllex.Slot{plain, plain},
		},
		{
			name: "bare carriage return stays inside a line comment",
			sql:  "SELECT a FROM t WHERE b = ? -- note ?\r more ?\nAND c = ?",
			want: []sqllex.Slot{plain, plain},
		},
		{
			name: "version comment body is live SQL",
			sql:  "SELECT a FROM t /*! WHERE b = ? */ AND c = ?",
			want: []sqllex.Slot{plain, plain},
		},
		{
			// MySQL has no ?N: the digits are ordinary text.
			name: "digits after a placeholder are text",
			sql:  "SELECT a FROM t WHERE b = ?1",
			want: []sqllex.Slot{plain},
		},
		{
			name: "block comment that is not a marker hides its slot",
			sql:  "SELECT a FROM t /*SLICE ids*/ WHERE b = ?",
			want: []sqllex.Slot{plain},
		},
		{
			name: "detached marker binds nothing",
			sql:  "SELECT a FROM t WHERE id IN (/*SLICE:ids*/ ?)",
			want: []sqllex.Slot{plain},
		},
		{
			name: "brackets are ordinary text",
			sql:  "SELECT [a? FROM t WHERE b = ?",
			want: []sqllex.Slot{plain, plain},
		},
		{name: "unterminated string swallows the rest", sql: "SELECT a FROM t WHERE b = 'open ?"},
		{name: "unterminated block comment swallows the rest", sql: "SELECT a /* open ? FROM t"},
	}
	runSlotCases(t, cases, sqllex.MySQLRaw)
}

func TestSlotsMySQLPyformat(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sql  string
		want []sqllex.Slot
	}{
		{
			name: "rewritten placeholders in text order",
			sql:  "SELECT a FROM t WHERE b = %s AND c = %s",
			want: []sqllex.Slot{plain, plain},
		},
		{
			// A doubled percent is a literal, not a slot, and must be
			// matched before the placeholder itself.
			name: "doubled percent is not a slot",
			sql:  "SELECT a FROM t WHERE b LIKE '50%%' AND c = %s",
			want: []sqllex.Slot{plain},
		},
		{
			name: "lone percent is not a slot",
			sql:  "SELECT a %% 2 FROM t WHERE b = %s",
			want: []sqllex.Slot{plain},
		},
		{
			name: "rewritten marker carries its text",
			sql:  "SELECT a FROM t WHERE id IN (/*SLICE:ids*/%s)",
			want: []sqllex.Slot{{Name: "ids", Marker: "/*SLICE:ids*/%s"}},
		},
		{
			// The rewriter doubles a percent inside the marker; the name has
			// to come back undoubled so it matches sqlc's parameter.
			name: "doubled percent in a slice name is undone",
			sql:  "SELECT a FROM t WHERE id IN (/*SLICE:a%%b*/%s)",
			want: []sqllex.Slot{{Name: "a%b", Marker: "/*SLICE:a%%b*/%s"}},
		},
		{
			name: "reused marker interleaved with a plain slot",
			sql:  "SELECT a FROM t WHERE x IN (/*SLICE:ids*/%s) AND y = %s OR z IN (/*SLICE:ids*/%s)",
			want: []sqllex.Slot{
				{Name: "ids", Marker: "/*SLICE:ids*/%s"},
				plain,
				{Name: "ids", Marker: "/*SLICE:ids*/%s"},
			},
		},
	}
	runSlotCases(t, cases, sqllex.MySQLPyformat)
}

func TestSlotsSQLite(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sql  string
		want []sqllex.Slot
	}{
		{
			// sqlc numbers SQLite parameters; the digits belong to the slot.
			name: "numbered placeholders count once each",
			sql:  "SELECT a FROM t WHERE b = ?1 AND c = ?12",
			want: []sqllex.Slot{plain, plain},
		},
		{
			name: "slice marker carries its text",
			sql:  "SELECT a FROM t WHERE id IN (/*SLICE:ids*/?)",
			want: []sqllex.Slot{{Name: "ids", Marker: "/*SLICE:ids*/?"}},
		},
		{
			name: "reused marker interleaved with a numbered slot",
			sql:  "SELECT a FROM t WHERE x IN (/*SLICE:ids*/?) AND y = ?2 OR z IN (/*SLICE:ids*/?)",
			want: []sqllex.Slot{
				{Name: "ids", Marker: "/*SLICE:ids*/?"},
				plain,
				{Name: "ids", Marker: "/*SLICE:ids*/?"},
			},
		},
		{
			// A "?" inside a quoted identifier is not a bind slot; counting
			// it misorders a reused slice's arguments.
			name: "backtick identifier hides its placeholder",
			sql:  "SELECT `we?rd`, `bt``q?` FROM t WHERE a = ?",
			want: []sqllex.Slot{plain},
		},
		{
			name: "bracket identifier hides its placeholder",
			sql:  "SELECT [br?k] FROM t WHERE a = ?",
			want: []sqllex.Slot{plain},
		},
		{
			// SQLite has no backslash escape: the literal ends at the quote.
			name: "backslash does not escape a quote",
			sql:  "SELECT 'a\\' FROM t WHERE a = ?",
			want: []sqllex.Slot{plain},
		},
		{
			// SQLite needs no whitespace after the dashes.
			name: "dash comment without a gap hides its slot",
			sql:  "SELECT a FROM t WHERE b = 5--?\nAND c = ?",
			want: []sqllex.Slot{plain},
		},
		{
			// "#" is not a comment in SQLite, so the slot after it is live.
			name: "hash is ordinary text",
			sql:  "SELECT a FROM t WHERE b = ? # ?",
			want: []sqllex.Slot{plain, plain},
		},
		{
			// SQLite treats /*! as an ordinary comment, unlike MySQL.
			name: "version comment is an ordinary comment",
			sql:  "SELECT a FROM t /*! AND b = ? */ WHERE c = ?",
			want: []sqllex.Slot{plain},
		},
		{name: "unterminated bracket swallows the rest", sql: "SELECT [br?k FROM t WHERE a = ?"},
	}
	runSlotCases(t, cases, sqllex.SQLite)
}

// TestScanSpansCoverInput pins the property the MySQL rewriter relies on:
// the tokens tile the input exactly, so copying every span reproduces it.
func TestScanSpansCoverInput(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"SELECT a FROM t WHERE b = ? AND c IN (/*SLICE:ids*/?)",
		"SELECT '?' -- c\n# h\n/* b */ `id` /*! live ? */ FROM t WHERE x = ?",
		"SELECT a FROM t WHERE b = 5--? AND c = ?",
		"",
	}
	for _, sql := range inputs {
		t.Run(sql, func(t *testing.T) {
			t.Parallel()
			at := 0
			for _, token := range sqllex.Scan(sql, sqllex.MySQLRaw) {
				if token.Start != at {
					t.Fatalf("token starts at %d, want %d (gap or overlap)", token.Start, at)
				}
				at = token.End
			}
			// The trailing flush covers whatever the loop left pending, so
			// the spans must reach the end: bytes no token covers would be
			// dropped from the rewritten SQL.
			if at != len(sql) {
				t.Fatalf("tokens cover %d bytes, want %d", at, len(sql))
			}
		})
	}
}

func runSlotCases(t *testing.T, cases []struct {
	name string
	sql  string
	want []sqllex.Slot
}, dialect sqllex.Dialect,
) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := sqllex.Slots(tc.sql, dialect); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Slots() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestDialectEmitters(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		dialect     sqllex.Dialect
		placeholder string
		marker      string
	}{
		{name: "mysql raw", dialect: sqllex.MySQLRaw, placeholder: "?", marker: "/*SLICE:ids*/?"},
		{name: "mysql pyformat", dialect: sqllex.MySQLPyformat, placeholder: "%s", marker: "/*SLICE:ids*/%s"},
		{name: "sqlite", dialect: sqllex.SQLite, placeholder: "?", marker: "/*SLICE:ids*/?"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.dialect.Placeholder(); got != tc.placeholder {
				t.Errorf("Placeholder() = %q, want %q", got, tc.placeholder)
			}
			if got := tc.dialect.SliceMarker("ids"); got != tc.marker {
				t.Errorf("SliceMarker() = %q, want %q", got, tc.marker)
			}
			// A rebuilt marker has to scan back to the slot it describes.
			want := []sqllex.Slot{{Name: "ids", Marker: tc.marker}}
			if got := sqllex.Slots(tc.dialect.SliceMarker("ids"), tc.dialect); !reflect.DeepEqual(got, want) {
				t.Errorf("Slots(SliceMarker()) = %+v, want %+v", got, want)
			}
		})
	}
}
