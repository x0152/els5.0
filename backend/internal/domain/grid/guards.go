package grid

import (
	"fmt"

	"github.com/els/backend/internal/domain/shared"
)

func EnforceSchemaVersion(got, want string) error {
	if got == "" {
		return shared.Validation(fmt.Errorf("schema_version: must not be empty"))
	}
	if got != want {
		return fmt.Errorf("%w: schema_version mismatch (got %s, want %s)", shared.ErrConflict, got, want)
	}
	return nil
}

func EnforceRowVersion(got, current int64) error {
	if got != current {
		return fmt.Errorf("%w: row base_version mismatch (got %d, want %d)", shared.ErrConflict, got, current)
	}
	return nil
}

func ValidateOp(op Op, schema Schema) error {
	if !op.Kind.IsValid() {
		return shared.Validation(fmt.Errorf("op.kind: invalid %q", op.Kind))
	}
	switch op.Kind {
	case OpCreate:
		if op.TempID == "" {
			return shared.Validation(fmt.Errorf("op.temp_id: must not be empty for create"))
		}
		if err := validatePatch(op.Data, schema, true); err != nil {
			return err
		}
	case OpUpdate:
		if op.ID == "" {
			return shared.Validation(fmt.Errorf("op.id: must not be empty for update"))
		}
		if len(op.Data) == 0 {
			return shared.Validation(fmt.Errorf("op.data: must not be empty for update"))
		}
		if err := validatePatch(op.Data, schema, false); err != nil {
			return err
		}
	case OpDelete:
		if op.ID == "" {
			return shared.Validation(fmt.Errorf("op.id: must not be empty for delete"))
		}
	}
	return nil
}

func validatePatch(data map[ColumnID]any, schema Schema, isCreate bool) error {
	for id := range data {
		c, ok := schema.Column(id)
		if !ok {
			return shared.Validation(fmt.Errorf("column %q: unknown", id))
		}
		if c.Readonly {
			return shared.Validation(fmt.Errorf("column %q: readonly", id))
		}
	}
	if isCreate {
		for _, c := range schema.Columns {
			if !c.Required || c.Readonly {
				continue
			}
			if _, ok := data[c.ID]; !ok {
				return shared.Validation(fmt.Errorf("column %q: required", c.ID))
			}
		}
	}
	return nil
}
