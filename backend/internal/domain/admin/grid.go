package admin

import (
	"fmt"

	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

func NewGrid(actor *iam.Actor) grid.Grid[*Administrator] {
	columns := iam.AccountColumns()

	return grid.New(grid.Spec[*Administrator]{
		Columns: columns,
		Row: func(a *Administrator) grid.Row {
			cells := iam.AccountCells(a.AccountSide)
			return grid.Row{
				ID:          a.ID().String(),
				BaseVersion: a.Version(),
				Cells:       cells,
			}
		},
		ApplyPatch: func(a *Administrator, data map[grid.ColumnID]any) error {
			if err := iam.RequireGlobalAdmin(actor); err != nil {
				return err
			}
			handled, err := iam.ApplyAccountPatch(a.AccountSide, data)
			if err != nil {
				return err
			}
			for id := range data {
				if _, ok := handled[id]; ok {
					continue
				}
				return shared.Validation(fmt.Errorf("column %q: not editable", id))
			}
			return nil
		},
	})
}
