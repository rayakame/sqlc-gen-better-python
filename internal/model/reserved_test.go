package model_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
)

// pythonKeywords mirrors every arm of the IsReserved switch in reserved.go.
var pythonKeywords = []string{
	"False", "None", "True", "and", "as", "assert", "async", "await",
	"break", "class", "continue", "def", "del", "elif", "else", "except",
	"finally", "for", "from", "global", "if", "import", "in", "is",
	"lambda", "nonlocal", "not", "or", "pass", "raise", "return", "try",
	"while", "with", "yield", "id",
}

func TestIsReserved(t *testing.T) {
	t.Parallel()
	for _, keyword := range pythonKeywords {
		t.Run("reserved "+keyword, func(t *testing.T) {
			t.Parallel()
			if !model.IsReserved(keyword) {
				t.Errorf("IsReserved(%q) = false, want true", keyword)
			}
		})
	}
	nonReserved := []string{"", "user", "true", "IF", "Id", "ids", "match", "case", "type", "print"}
	for _, name := range nonReserved {
		t.Run("non-reserved "+name, func(t *testing.T) {
			t.Parallel()
			if model.IsReserved(name) {
				t.Errorf("IsReserved(%q) = true, want false", name)
			}
		})
	}
}

func TestEscape(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "reserved keyword gets underscore suffix", in: "class", want: "class_"},
		{name: "id gets underscore suffix", in: "id", want: "id_"},
		{name: "non-reserved unchanged", in: "users", want: "users"},
		{name: "empty unchanged", in: "", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := model.Escape(tc.in); got != tc.want {
				t.Errorf("Escape(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
