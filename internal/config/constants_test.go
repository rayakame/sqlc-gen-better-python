package config_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
)

func TestSQLDriverString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		driver config.SQLDriver
		want   string
	}{
		{name: "sqlite3", driver: config.SQLDriverSQLite, want: "sqlite3"},
		{name: "aiosqlite", driver: config.SQLDriverAioSQLite, want: "aiosqlite"},
		{name: "asyncpg", driver: config.SQLDriverAsyncpg, want: "asyncpg"},
		{name: "arbitrary value round-trips", driver: config.SQLDriver("mysql"), want: "mysql"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.driver.String(); got != tc.want {
				t.Errorf("SQLDriver.String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSQLDriverValidate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		driver  config.SQLDriver
		engine  string
		wantErr string
	}{
		{name: "sqlite3 with sqlite", driver: config.SQLDriverSQLite, engine: "sqlite"},
		{name: "aiosqlite with sqlite", driver: config.SQLDriverAioSQLite, engine: "sqlite"},
		{name: "asyncpg with postgresql", driver: config.SQLDriverAsyncpg, engine: "postgresql"},
		{name: "psycopg_async with postgresql", driver: config.SQLDriverPsycopgAsync, engine: "postgresql"},
		{name: "psycopg_sync with postgresql", driver: config.SQLDriverPsycopgSync, engine: "postgresql"},
		{
			name:    "psycopg_sync with sqlite",
			driver:  config.SQLDriverPsycopgSync,
			engine:  "sqlite",
			wantErr: "SQL driver psycopg_sync does not support sqlite",
		},
		{
			name:    "sqlite3 with postgresql",
			driver:  config.SQLDriverSQLite,
			engine:  "postgresql",
			wantErr: "SQL driver sqlite3 does not support postgresql",
		},
		{
			name:    "aiosqlite with postgresql",
			driver:  config.SQLDriverAioSQLite,
			engine:  "postgresql",
			wantErr: "SQL driver aiosqlite does not support postgresql",
		},
		{
			name:    "asyncpg with sqlite",
			driver:  config.SQLDriverAsyncpg,
			engine:  "sqlite",
			wantErr: "SQL driver asyncpg does not support sqlite",
		},
		{
			name:    "unknown driver",
			driver:  config.SQLDriver("mysql"),
			engine:  "postgresql",
			wantErr: "unknown SQL driver: mysql",
		},
		{
			name:    "empty driver",
			driver:  config.SQLDriver(""),
			engine:  "sqlite",
			wantErr: "unknown SQL driver: ",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.driver.Validate(tc.engine)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("Validate(%q) error = %v, want nil", tc.engine, err)
				}
			} else {
				if err == nil {
					t.Fatalf("Validate(%q) error = nil, want %q", tc.engine, tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Errorf("Validate(%q) error = %q, want %q", tc.engine, err.Error(), tc.wantErr)
				}
			}
		})
	}
}

func TestModelTypeValid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		modelType config.ModelType
		want      bool
	}{
		{name: "dataclass", modelType: config.ModelTypeDataclass, want: true},
		{name: "attrs", modelType: config.ModelTypeAttrs, want: true},
		{name: "msgspec", modelType: config.ModelTypeMsgspec, want: true},
		{name: "pydantic", modelType: config.ModelTypePydantic, want: true},
		{name: "empty", modelType: config.ModelType(""), want: false},
		{name: "unknown", modelType: config.ModelType("struct"), want: false},
		{name: "wrong case", modelType: config.ModelType("Dataclass"), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.modelType.Valid(); got != tc.want {
				t.Errorf("ModelType(%q).Valid() = %v, want %v", string(tc.modelType), got, tc.want)
			}
		})
	}
}

func TestDocstringConventionValid(t *testing.T) {
	t.Parallel()
	valid := []config.DocstringConvention{
		config.DocstringConventionNone,
		config.DocstringConventionGoogle,
		config.DocstringConventionNumpy,
		config.DocstringConventionPEP257,
	}
	for _, ds := range valid {
		if !ds.Valid() {
			t.Errorf("DocstringConvention(%q).Valid() = false, want true", string(ds))
		}
	}
	invalid := []config.DocstringConvention{"", "rst", "NONE", "google "}
	for _, ds := range invalid {
		if ds.Valid() {
			t.Errorf("DocstringConvention(%q).Valid() = true, want false", string(ds))
		}
	}
}

func TestSQLDriverIsPsycopg(t *testing.T) {
	t.Parallel()
	for driver, want := range map[config.SQLDriver]bool{
		config.SQLDriverPsycopgAsync: true,
		config.SQLDriverPsycopgSync:  true,
		config.SQLDriverAsyncpg:      false,
		config.SQLDriverAioSQLite:    false,
		config.SQLDriverSQLite:       false,
	} {
		if got := driver.IsPsycopg(); got != want {
			t.Errorf("IsPsycopg(%q) = %v, want %v", driver, got, want)
		}
	}
}
