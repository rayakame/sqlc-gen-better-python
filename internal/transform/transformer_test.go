package transform_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/transform"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestNewTransformer(t *testing.T) {
	t.Parallel()
	req := &plugin.GenerateRequest{Catalog: &plugin.Catalog{DefaultSchema: "public"}}

	tf := transform.NewTransformer(&config.Config{}, req, types.PostgresTypeToPython)

	if tf == nil {
		t.Fatal("NewTransformer() = nil, want a transformer")
	}
	// An empty catalog yields empty (not nil) model slices.
	if enums := tf.BuildEnums(); enums == nil || len(enums) != 0 {
		t.Errorf("BuildEnums() = %#v, want empty slice", enums)
	}
	if tables := tf.BuildTables(); tables == nil || len(tables) != 0 {
		t.Errorf("BuildTables() = %#v, want empty slice", tables)
	}
}
