package types_test

import (
	"reflect"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/types"
)

func TestGetTypeConversionFunc(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		engine  string
		want    types.TypeConversionFunc
		wantErr string
	}{
		{name: "postgresql", engine: "postgresql", want: types.PostgresTypeToPython},
		{name: "sqlite", engine: "sqlite", want: types.SqliteTypeToPython},
		{name: "unsupported engine", engine: "mysql", wantErr: `engine "mysql" is not supported`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := types.GetTypeConversionFunc(tc.engine)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("GetTypeConversionFunc(%q) error = %v, want %q", tc.engine, err, tc.wantErr)
				}
				if got != nil {
					t.Errorf("GetTypeConversionFunc(%q) = non-nil func, want nil", tc.engine)
				}

				return
			}
			if err != nil {
				t.Fatalf("GetTypeConversionFunc(%q) unexpected error: %v", tc.engine, err)
			}
			// Functions are not comparable; compare their code pointers.
			if reflect.ValueOf(got).Pointer() != reflect.ValueOf(tc.want).Pointer() {
				t.Errorf("GetTypeConversionFunc(%q) returned the wrong conversion func", tc.engine)
			}
		})
	}
}
