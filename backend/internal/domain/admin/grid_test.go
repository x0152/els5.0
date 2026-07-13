package admin_test

import (
	"testing"

	admindom "github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/grid"
	iamdom "github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
	admintest "github.com/els/backend/internal/utils/test/admin"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestAdminGrid_Row(t *testing.T) {
	t.Run("renders_account_cells", func(t *testing.T) {
		a := admintest.New(t).Build(t)
		g := admindom.NewGrid(iamtest.Admin(t))

		row := g.RowOf(a)

		if row.ID != a.ID().String() {
			t.Errorf("expected row ID=%s, got %s", a.ID(), row.ID)
		}
		if row.Cells[iamdom.ColAccountEmail] == nil {
			t.Errorf("expected account email to be rendered, got nil")
		}
	})
}

func TestAdminGrid_ApplyPatch(t *testing.T) {
	t.Run("non_admin_forbidden", func(t *testing.T) {
		a := admintest.New(t).Build(t)
		g := admindom.NewGrid(iamtest.Expert(t, vo.NewID()))

		err := g.ApplyPatch(a, map[grid.ColumnID]any{
			iamdom.ColAccountFirstName: "Alex",
		})

		test.ErrIs(t, err, shared.ErrForbidden)
	})

	t.Run("admin_renames_via_account_patch", func(t *testing.T) {
		a := admintest.New(t).Build(t)
		g := admindom.NewGrid(iamtest.Admin(t))

		err := g.ApplyPatch(a, map[grid.ColumnID]any{
			iamdom.ColAccountFirstName: "Alex",
			iamdom.ColAccountLastName:  "Smith",
		})

		test.NoErr(t, err)
		if a.FirstName() != "Alex" || a.LastName() != "Smith" {
			t.Errorf("expected name=Alex Smith, got %s %s", a.FirstName(), a.LastName())
		}
	})

	t.Run("unknown_column_returns_validation", func(t *testing.T) {
		a := admintest.New(t).Build(t)
		g := admindom.NewGrid(iamtest.Admin(t))

		err := g.ApplyPatch(a, map[grid.ColumnID]any{
			grid.ColumnID("not_a_real_column"): "x",
		})

		test.ErrIs(t, err, shared.ErrValidation)
	})
}
