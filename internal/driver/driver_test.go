package driver_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/driver"
)

func TestNew(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		sqlDriver config.SQLDriver
		wantName  string
		wantAsync bool
	}{
		{name: "asyncpg", sqlDriver: config.SQLDriverAsyncpg, wantName: "asyncpg", wantAsync: true},
		{name: "psycopg_async", sqlDriver: config.SQLDriverPsycopgAsync, wantName: "psycopg", wantAsync: true},
		{name: "psycopg_sync", sqlDriver: config.SQLDriverPsycopgSync, wantName: "psycopg", wantAsync: false},
		{name: "aiosqlite", sqlDriver: config.SQLDriverAioSQLite, wantName: "aiosqlite", wantAsync: true},
		{name: "sqlite3", sqlDriver: config.SQLDriverSQLite, wantName: "sqlite3", wantAsync: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d, err := driver.New(&config.Config{SqlDriver: tc.sqlDriver})
			if err != nil {
				t.Fatalf("New(%q) error = %v, want nil", tc.sqlDriver, err)
			}
			if got := d.Name(); got != tc.wantName {
				t.Errorf("Name() = %q, want %q", got, tc.wantName)
			}
			if got := d.IsAsync(); got != tc.wantAsync {
				t.Errorf("IsAsync() = %v, want %v", got, tc.wantAsync)
			}
		})
	}
}

func TestNewUnsupportedDriver(t *testing.T) {
	t.Parallel()
	d, err := driver.New(&config.Config{SqlDriver: "mysql"})
	if d != nil {
		t.Errorf("New(\"mysql\") driver = %v, want nil", d)
	}
	if err == nil {
		t.Fatal("New(\"mysql\") error = nil, want non-nil")
	}
	if got, want := err.Error(), "unsupported driver: mysql"; got != want {
		t.Errorf("New(\"mysql\") error = %q, want %q", got, want)
	}
}
