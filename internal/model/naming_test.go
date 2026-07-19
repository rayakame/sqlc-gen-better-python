package model_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestSnakeToCamel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		initialisms map[string]struct{}
		in          string
		want        string
	}{
		{name: "plain snake case", in: "user_name", want: "UserName"},
		{name: "single word", in: "users", want: "Users"},
		{name: "initialism uppercased", initialisms: map[string]struct{}{"id": {}}, in: "user_id", want: "UserID"},
		{name: "no initialism entry keeps title case", in: "user_id", want: "UserId"},
		{name: "invalid chars become separators", in: "user-name.first", want: "UserNameFirst"},
		{name: "empty input gets Model prefix", in: "", want: "Model"},
		{name: "only invalid chars gets Model prefix", in: "$$$", want: "Model"},
		{name: "digit-leading result gets Model prefix", in: "123_users", want: "Model123Users"},
		{name: "reserved True gets Model prefix", in: "true", want: "ModelTrue"},
		{name: "reserved False gets Model prefix", in: "false", want: "ModelFalse"},
		{name: "reserved None gets Model prefix", in: "none", want: "ModelNone"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			conf := &config.Config{InitialismsMap: tc.initialisms}
			if got := model.SnakeToCamel(conf, tc.in); got != tc.want {
				t.Errorf("SnakeToCamel(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestUpperSnakeCase(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "lowercase word", in: "camel", want: "CAMEL"},
		{name: "camel case gets underscores", in: "CamelCase", want: "CAMEL_CASE"},
		{name: "leading upper gets no underscore", in: "X", want: "X"},
		{name: "consecutive uppers split individually", in: "myID", want: "MY_I_D"},
		{name: "snake case stays snake case", in: "already_snake", want: "ALREADY_SNAKE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := model.UpperSnakeCase(tc.in); got != tc.want {
				t.Errorf("UpperSnakeCase(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestColumnName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		column *plugin.Column
		pos    int
		want   string
	}{
		{name: "named column", column: &plugin.Column{Name: "email"}, pos: 0, want: "email"},
		{name: "unnamed column at position zero", column: &plugin.Column{Name: ""}, pos: 0, want: "column_1"},
		{name: "unnamed column is one-based", column: &plugin.Column{Name: ""}, pos: 4, want: "column_5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := model.ColumnName(tc.column, tc.pos); got != tc.want {
				t.Errorf("ColumnName(%v, %d) = %q, want %q", tc.column, tc.pos, got, tc.want)
			}
		})
	}
}

func TestEscapedColumnName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		column *plugin.Column
		pos    int
		want   string
	}{
		{name: "plain name", column: &plugin.Column{Name: "email"}, pos: 0, want: "email"},
		{name: "invalid chars sanitized", column: &plugin.Column{Name: "user name"}, pos: 0, want: "user_name"},
		{name: "digit-leading gets column_ prefix", column: &plugin.Column{Name: "1abc"}, pos: 0, want: "column_1abc"},
		{
			name:   "underscore-leading gets column_ prefix",
			column: &plugin.Column{Name: "_private"},
			pos:    0,
			want:   "column__private",
		},
		{name: "reserved keyword escaped", column: &plugin.Column{Name: "class"}, pos: 0, want: "class_"},
		{name: "id is reserved", column: &plugin.Column{Name: "id"}, pos: 0, want: "id_"},
		{name: "unnamed column positional fallback", column: &plugin.Column{Name: ""}, pos: 2, want: "column_3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := model.EscapedColumnName(tc.column, tc.pos); got != tc.want {
				t.Errorf("EscapedColumnName(%v, %d) = %q, want %q", tc.column, tc.pos, got, tc.want)
			}
		})
	}
}

func TestParamName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		param *plugin.Parameter
		want  string
	}{
		{name: "named parameter", param: &plugin.Parameter{Column: &plugin.Column{Name: "email"}}, want: "email"},
		{
			name:  "empty name falls back to dollar_N",
			param: &plugin.Parameter{Number: 3, Column: &plugin.Column{Name: ""}},
			want:  "dollar_3",
		},
		{name: "nil column falls back to dollar_N", param: &plugin.Parameter{Number: 7}, want: "dollar_7"},
		{
			name:  "digit-leading gets column_ prefix",
			param: &plugin.Parameter{Column: &plugin.Column{Name: "1abc"}},
			want:  "column_1abc",
		},
		{name: "leading underscore is allowed", param: &plugin.Parameter{Column: &plugin.Column{Name: "_id"}}, want: "_id"},
		{name: "invalid chars sanitized", param: &plugin.Parameter{Column: &plugin.Column{Name: "user name"}}, want: "user_name"},
		{name: "reserved keyword escaped", param: &plugin.Parameter{Column: &plugin.Column{Name: "for"}}, want: "for_"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := model.ParamName(tc.param); got != tc.want {
				t.Errorf("ParamName(%v) = %q, want %q", tc.param, got, tc.want)
			}
		})
	}
}

func TestEnumConstantName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value string
		index int
		want  string
	}{
		{name: "plain value uppercased", value: "happy", index: 0, want: "HAPPY"},
		{name: "invalid chars sanitized", value: "very-happy", index: 0, want: "VERY_HAPPY"},
		{name: "empty value falls back to VALUE_N", value: "", index: 2, want: "VALUE_3"},
		{name: "only separators falls back to VALUE_N", value: "---", index: 0, want: "VALUE_1"},
		{name: "digit-leading gets VALUE_ prefix", value: "1st", index: 0, want: "VALUE_1ST"},
		{name: "underscore-leading gets VALUE_ prefix", value: "_hidden", index: 0, want: "VALUE__HIDDEN"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seen := map[string]int{}
			if got := model.EnumConstantName(tc.value, tc.index, seen); got != tc.want {
				t.Errorf("EnumConstantName(%q, %d) = %q, want %q", tc.value, tc.index, got, tc.want)
			}
		})
	}
	t.Run("duplicates get numeric suffix", func(t *testing.T) {
		t.Parallel()
		seen := map[string]int{}
		if got := model.EnumConstantName("happy", 0, seen); got != "HAPPY" {
			t.Fatalf("first EnumConstantName = %q, want %q", got, "HAPPY")
		}
		if got := model.EnumConstantName("HAPPY", 1, seen); got != "HAPPY_2" {
			t.Fatalf("second EnumConstantName = %q, want %q", got, "HAPPY_2")
		}
	})
}

func TestDedupName(t *testing.T) {
	t.Parallel()
	seen := map[string]int{}
	// Order matters: later steps depend on the names taken by earlier ones.
	steps := []struct {
		in   string
		want string
	}{
		{in: "x", want: "x"},
		{in: "x", want: "x_2"},
		{in: "x", want: "x_3"},
		{in: "y_2", want: "y_2"},
		{in: "y", want: "y"},
		{in: "y", want: "y_3"},
	}
	for i, step := range steps {
		if got := model.DedupName(step.in, seen); got != step.want {
			t.Fatalf("step %d: DedupName(%q) = %q, want %q", i, step.in, got, step.want)
		}
	}
}

func TestDedupClassName(t *testing.T) {
	t.Parallel()
	seen := map[string]int{}
	steps := []struct {
		in   string
		want string
	}{
		{in: "Name", want: "Name"},
		{in: "Name", want: "Name2"},
		{in: "Foo2", want: "Foo2"},
		{in: "Foo", want: "Foo"},
		{in: "Foo", want: "Foo3"},
	}
	for i, step := range steps {
		if got := model.DedupClassName(step.in, seen); got != step.want {
			t.Fatalf("step %d: DedupClassName(%q) = %q, want %q", i, step.in, got, step.want)
		}
	}
}

func TestModelName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		conf   config.Config
		table  string
		schema string
		want   string
	}{
		{name: "singularized by default", table: "users", want: "User"},
		{name: "emit exact table names", conf: config.Config{EmitExactTableNames: true}, table: "users", want: "Users"},
		{
			name:  "bare exclusion skips singularization",
			conf:  config.Config{InflectionExcludeTableNames: []string{"users"}},
			table: "users",
			want:  "Users",
		},
		{
			name:  "exclusion matches case-insensitively",
			conf:  config.Config{InflectionExcludeTableNames: []string{"USERS"}},
			table: "users",
			want:  "Users",
		},
		{
			name:   "schema-qualified exclusion matches",
			conf:   config.Config{InflectionExcludeTableNames: []string{"analytics_events"}},
			table:  "events",
			schema: "analytics",
			want:   "AnalyticsEvents",
		},
		{name: "schema-qualified table singularized", table: "events", schema: "analytics", want: "AnalyticsEvent"},
		{
			name:  "non-matching exclusion still singularizes",
			conf:  config.Config{InflectionExcludeTableNames: []string{"orders"}},
			table: "users",
			want:  "User",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := model.ModelName(&tc.conf, tc.table, tc.schema); got != tc.want {
				t.Errorf("ModelName(%q, %q) = %q, want %q", tc.table, tc.schema, got, tc.want)
			}
		})
	}
}

func TestEnumName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		enum   string
		schema string
		want   string
	}{
		{name: "bare enum", enum: "mood", schema: "", want: "Mood"},
		{name: "schema-qualified enum", enum: "mood", schema: "analytics", want: "AnalyticsMood"},
		{name: "never singularized", enum: "moods", schema: "", want: "Moods"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := model.EnumName(&config.Config{}, tc.enum, tc.schema); got != tc.want {
				t.Errorf("EnumName(%q, %q) = %q, want %q", tc.enum, tc.schema, got, tc.want)
			}
		})
	}
}
