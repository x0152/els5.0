package grid_test

import (
	"errors"
	"testing"

	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
)

func sampleSchema() grid.Schema {
	return grid.Schema{
		Columns: []grid.Column{
			{ID: "name", Type: grid.TypeText, Required: true},
			{ID: "email", Type: grid.TypeEmail, Required: true},
			{ID: "id", Type: grid.TypeText, Required: true, Readonly: true},
			{ID: "notes", Type: grid.TypeText},
		},
		Version: "v1",
	}
}

func TestEnforceSchemaVersion(t *testing.T) {
	if err := grid.EnforceSchemaVersion("", "v1"); !errors.Is(err, shared.ErrValidation) {
		t.Errorf("empty got: expected ErrValidation, got %v", err)
	}
	if err := grid.EnforceSchemaVersion("v0", "v1"); !errors.Is(err, shared.ErrConflict) {
		t.Errorf("mismatch: expected ErrConflict, got %v", err)
	}
	if err := grid.EnforceSchemaVersion("v1", "v1"); err != nil {
		t.Errorf("match: expected nil, got %v", err)
	}
}

func TestEnforceRowVersion(t *testing.T) {
	if err := grid.EnforceRowVersion(0, 5); !errors.Is(err, shared.ErrConflict) {
		t.Errorf("mismatch: expected ErrConflict, got %v", err)
	}
	test.NoErr(t, grid.EnforceRowVersion(5, 5))
}

func TestValidateOp(t *testing.T) {
	schema := sampleSchema()
	cases := []struct {
		name    string
		op      grid.Op
		wantErr bool
	}{
		{
			name:    "create_missing_temp_id",
			op:      grid.Op{Kind: grid.OpCreate, Data: map[grid.ColumnID]any{"name": "x", "email": "x@y"}},
			wantErr: true,
		},
		{
			name:    "create_unknown_column",
			op:      grid.Op{Kind: grid.OpCreate, TempID: "t1", Data: map[grid.ColumnID]any{"name": "x", "email": "x@y", "garbage": "y"}},
			wantErr: true,
		},
		{
			name:    "create_missing_required",
			op:      grid.Op{Kind: grid.OpCreate, TempID: "t1", Data: map[grid.ColumnID]any{"name": "x"}},
			wantErr: true,
		},
		{
			name: "create_ok",
			op: grid.Op{Kind: grid.OpCreate, TempID: "t1", Data: map[grid.ColumnID]any{
				"name": "x", "email": "x@y",
			}},
		},
		{
			name:    "update_missing_id",
			op:      grid.Op{Kind: grid.OpUpdate, Data: map[grid.ColumnID]any{"name": "x"}},
			wantErr: true,
		},
		{
			name:    "update_empty_data",
			op:      grid.Op{Kind: grid.OpUpdate, ID: "id-1"},
			wantErr: true,
		},
		{
			name:    "update_readonly_column",
			op:      grid.Op{Kind: grid.OpUpdate, ID: "id-1", Data: map[grid.ColumnID]any{"id": "new"}},
			wantErr: true,
		},
		{
			name: "update_ok",
			op:   grid.Op{Kind: grid.OpUpdate, ID: "id-1", Data: map[grid.ColumnID]any{"name": "y"}},
		},
		{
			name:    "delete_missing_id",
			op:      grid.Op{Kind: grid.OpDelete},
			wantErr: true,
		},
		{
			name: "delete_ok",
			op:   grid.Op{Kind: grid.OpDelete, ID: "id-1"},
		},
		{
			name:    "invalid_kind",
			op:      grid.Op{Kind: "garbage"},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := grid.ValidateOp(tc.op, schema)
			if tc.wantErr {
				test.ErrIs(t, err, shared.ErrValidation)
				return
			}
			test.NoErr(t, err)
		})
	}
}
