package transform

import (
	"reflect"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
)

func TestSqlitePlaceholders(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sql  string
		want []model.Placeholder
	}{
		{name: "no placeholders", sql: "SELECT 1"},
		{
			name: "bare placeholders in text order",
			sql:  "SELECT a FROM t WHERE b = ? AND c = ?",
			want: []model.Placeholder{plainSlot, plainSlot},
		},
		{
			// sqlc numbers SQLite parameters; the digits are part of the slot.
			name: "numbered placeholders count once",
			sql:  "SELECT a FROM t WHERE b = ?1 AND c = ?12",
			want: []model.Placeholder{plainSlot, plainSlot},
		},
		{
			name: "slice marker carries its text",
			sql:  "SELECT a FROM t WHERE id IN (/*SLICE:ids*/?)",
			want: []model.Placeholder{{SliceName: "ids", Marker: "/*SLICE:ids*/?"}},
		},
		{
			name: "reused marker interleaved with a numbered slot",
			sql:  "SELECT a FROM t WHERE x IN (/*SLICE:ids*/?) AND y = ?2 OR z IN (/*SLICE:ids*/?)",
			want: []model.Placeholder{
				{SliceName: "ids", Marker: "/*SLICE:ids*/?"},
				plainSlot,
				{SliceName: "ids", Marker: "/*SLICE:ids*/?"},
			},
		},
		{
			name: "placeholders inside strings are not slots",
			sql:  "SELECT '?', \"?\" FROM t WHERE a = ?",
			want: []model.Placeholder{plainSlot},
		},
		{
			name: "doubled quote is an escape, not a terminator",
			sql:  "SELECT 'a''?b' FROM t WHERE a = ?",
			want: []model.Placeholder{plainSlot},
		},
		{
			// SQLite has no backslash escape: the literal ends at the quote
			// and the following text is live SQL.
			name: "backslash does not escape a quote",
			sql:  "SELECT 'a\\' FROM t WHERE a = ?",
			want: []model.Placeholder{plainSlot},
		},
		{
			// A "?" inside a quoted identifier is not a bind slot; counting
			// it misorders a reused slice's arguments.
			name: "backtick identifier hides its question mark",
			sql:  "SELECT `we?rd` FROM t WHERE a = ?",
			want: []model.Placeholder{plainSlot},
		},
		{
			name: "doubled backtick is an escape",
			sql:  "SELECT `bt``q?` FROM t WHERE a = ?",
			want: []model.Placeholder{plainSlot},
		},
		{
			name: "bracket identifier hides its question mark",
			sql:  "SELECT [br?k] FROM t WHERE a = ?",
			want: []model.Placeholder{plainSlot},
		},
		{
			name: "unterminated bracket identifier swallows the rest",
			sql:  "SELECT [br?k FROM t WHERE a = ?",
		},
		{
			// SQLite needs no whitespace after the dashes.
			name: "dash comment without a gap hides its slot",
			sql:  "SELECT a FROM t WHERE b = 5--?\nAND c = ?",
			want: []model.Placeholder{plainSlot},
		},
		{
			// "#" is not a comment in SQLite, so the slot after it is live.
			name: "hash is ordinary text",
			sql:  "SELECT a FROM t WHERE b = ? # ?",
			want: []model.Placeholder{plainSlot, plainSlot},
		},
		{
			// SQLite treats /*! as an ordinary comment, unlike MySQL.
			name: "version comment is an ordinary comment",
			sql:  "SELECT a FROM t /*! AND b = ? */ WHERE c = ?",
			want: []model.Placeholder{plainSlot},
		},
		{
			name: "block comment hides its slot",
			sql:  "SELECT a FROM t /* ? */ WHERE b = ?",
			want: []model.Placeholder{plainSlot},
		},
		{
			name: "detached marker binds nothing",
			sql:  "SELECT a FROM t WHERE id IN (/*SLICE:ids*/ ?)",
			want: []model.Placeholder{plainSlot},
		},
		{
			name: "unterminated string swallows the rest",
			sql:  "SELECT a FROM t WHERE b = 'open ?",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := sqlitePlaceholders(tc.sql); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("sqlitePlaceholders() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
