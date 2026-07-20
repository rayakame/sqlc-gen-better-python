package internal_test

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

const (
	sqliteOptions      = `{"package":"foo","sql_driver":"sqlite3","emit_init_file":false}`
	sqliteInitOptions  = `{"package":"foo","sql_driver":"sqlite3","emit_init_file":true}`
	sqliteDebugOptions = `{"package":"foo","sql_driver":"sqlite3","emit_init_file":false,"debug":true}`
	asyncpgOptions     = `{"package":"foo","sql_driver":"asyncpg","emit_init_file":false}`
	asyncpgOmitOptions = `{"package":"foo","sql_driver":"asyncpg","emit_init_file":false,"omit_unused_models":true}`
)

func sqliteRequest(options string, queries ...*plugin.Query) *plugin.GenerateRequest {
	return &plugin.GenerateRequest{
		Settings:      &plugin.Settings{Engine: "sqlite"},
		Catalog:       &plugin.Catalog{DefaultSchema: "main", Schemas: []*plugin.Schema{{Name: "main"}}},
		Queries:       queries,
		PluginOptions: []byte(options),
	}
}

func postgresRequest(options string, queries ...*plugin.Query) *plugin.GenerateRequest {
	return &plugin.GenerateRequest{
		Settings: &plugin.Settings{Engine: "postgresql"},
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{{
				Name:  "public",
				Enums: []*plugin.Enum{{Name: "mood", Vals: []string{"happy", "sad"}}},
				Tables: []*plugin.Table{{
					Rel: &plugin.Identifier{Name: "users"},
					Columns: []*plugin.Column{
						{Name: "id", NotNull: true, Type: &plugin.Identifier{Name: "serial"}},
						{Name: "name", NotNull: true, Type: &plugin.Identifier{Name: "text"}},
					},
				}},
			}},
		},
		Queries:       queries,
		PluginOptions: []byte(options),
	}
}

func execQuery(name, filename string) *plugin.Query {
	return &plugin.Query{
		Name:     name,
		Cmd:      metadata.CmdExec,
		Text:     "DELETE FROM users",
		Filename: filename,
	}
}

func fileNames(files []*plugin.File) []string {
	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, file.Name)
	}
	slices.Sort(names)

	return names
}

func TestHandlerErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		req         *plugin.GenerateRequest
		wantContain string
	}{
		{
			"invalid plugin options json",
			sqliteRequest(`{`),
			"error trying to parse config",
		},
		{
			"failed options validation",
			sqliteRequest(`{"package":"foo","sql_driver":"sqlite3"}`),
			"error trying to parse config: invalid options: you need to specify emit_init_file",
		},
		{
			"unsupported query command for driver",
			sqliteRequest(sqliteOptions, &plugin.Query{
				Name:     "CopyUsers",
				Cmd:      metadata.CmdCopyFrom,
				Text:     "INSERT INTO users (id) VALUES (?)",
				Filename: "copy.sql",
			}),
			"error building queries",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp, err := internal.Handler(context.Background(), tc.req)
			if resp != nil {
				t.Errorf("Handler returned a response alongside an error: %v", resp)
			}
			if err == nil {
				t.Fatal("Handler returned nil error, want error")
			}
			if !strings.Contains(err.Error(), tc.wantContain) {
				t.Errorf("Handler error = %q, want it to contain %q", err, tc.wantContain)
			}
		})
	}
}

func TestHandlerFiles(t *testing.T) {
	t.Parallel()
	moodQuery := &plugin.Query{
		Name:     "GetMood",
		Cmd:      metadata.CmdOne,
		Text:     "SELECT mood FROM users LIMIT 1",
		Filename: "moods.sql",
		Columns:  []*plugin.Column{{Name: "mood", NotNull: true, Type: &plugin.Identifier{Name: "mood"}}},
	}
	cases := []struct {
		name string
		req  *plugin.GenerateRequest
		want []string
	}{
		{
			"empty catalog emits only models",
			sqliteRequest(sqliteOptions),
			[]string{"models.py"},
		},
		{
			"emit_init_file emits package init",
			sqliteRequest(sqliteInitOptions),
			[]string{"__init__.py", "models.py"},
		},
		{
			"enums and query module are emitted",
			postgresRequest(asyncpgOptions, execQuery("DeleteUsers", "user_queries.sql")),
			[]string{"enums.py", "models.py", "user_queries.py"},
		},
		{
			"queries split into one module per file",
			sqliteRequest(sqliteOptions, execQuery("DeleteA", "a.sql"), execQuery("DeleteB", "b.sql")),
			[]string{"a.py", "b.py", "models.py"},
		},
		{
			"omit_unused_models drops unused enums and tables",
			postgresRequest(asyncpgOmitOptions),
			[]string{"models.py"},
		},
		{
			"omit_unused_models keeps enums used by queries",
			postgresRequest(asyncpgOmitOptions, moodQuery),
			[]string{"enums.py", "models.py", "moods.py"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp, err := internal.Handler(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}
			if got := fileNames(resp.Files); !slices.Equal(got, tc.want) {
				t.Errorf("Handler emitted files %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHandlerDebugEmitsLog(t *testing.T) {
	t.Parallel()
	resp, err := internal.Handler(context.Background(), sqliteRequest(sqliteDebugOptions))
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	var logFile *plugin.File
	for _, file := range resp.Files {
		if file.Name == "log.json" {
			logFile = file
		}
	}
	if logFile == nil {
		t.Fatalf("Handler emitted files %v, want log.json among them", fileNames(resp.Files))
	}
	// The log package is a process-global singleton accumulating across
	// tests, so only verify the export is a valid JSON array.
	var entries []json.RawMessage
	if err := json.Unmarshal(logFile.Contents, &entries); err != nil {
		t.Errorf("log.json contents are not a JSON array: %v", err)
	}
}
