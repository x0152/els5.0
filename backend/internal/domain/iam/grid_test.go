package iam_test

import (
	"testing"

	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
)

func newSide(t *testing.T) iam.AccountSide {
	t.Helper()
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	test.NoErr(t, acc.ChangePictureURL("https://cdn/pic.png"))
	side, err := iam.NewAccountSide(acc)
	test.NoErr(t, err)
	return side
}

func TestAccountCells(t *testing.T) {
	// arrange
	side := newSide(t)

	// act
	cells := iam.AccountCells(side)

	// assert
	if cells[iam.ColAccountEmail] != side.Email().String() {
		t.Fatalf("expected email cell")
	}
	if cells[iam.ColAccountFirstName] != side.FirstName() || cells[iam.ColAccountLastName] != side.LastName() {
		t.Fatalf("expected first/last names")
	}
	if cells[iam.ColAccountStatus] != string(side.Status()) {
		t.Fatalf("expected status")
	}
	if cells[iam.ColAccountID] != side.AccountID().String() {
		t.Fatalf("expected account_id")
	}
	if cells[iam.ColAccountPictureURL] != "https://cdn/pic.png" {
		t.Fatalf("expected picture url")
	}
}

func TestAccountCells_NilPictureURL(t *testing.T) {
	// arrange
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	side, _ := iam.NewAccountSide(acc)

	// act
	cells := iam.AccountCells(side)

	// assert
	if v, ok := cells[iam.ColAccountPictureURL]; !ok || v != nil {
		t.Fatalf("expected nil picture URL value, got %v", v)
	}
}

func TestApplyAccountPatch_HappyPath(t *testing.T) {
	// arrange
	side := newSide(t)

	// act
	handled, err := iam.ApplyAccountPatch(side, map[grid.ColumnID]any{
		iam.ColAccountEmail:     "new@example.com",
		iam.ColAccountFirstName: "Jane",
		iam.ColAccountLastName:  "Smith",
		iam.ColAccountStatus:    string(iam.AccountStatusBlocked),
	})

	// assert
	test.NoErr(t, err)
	if len(handled) != 4 {
		t.Fatalf("expected 4 columns handled, got %d", len(handled))
	}
	if side.Email().String() != "new@example.com" {
		t.Fatalf("expected email updated")
	}
	if side.Name().Full() != "Jane Smith" {
		t.Fatalf("expected name updated")
	}
	if side.Status() != iam.AccountStatusBlocked {
		t.Fatalf("expected status updated")
	}
}

func TestApplyAccountPatch_ValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		data map[grid.ColumnID]any
	}{
		{name: "email_not_string", data: map[grid.ColumnID]any{iam.ColAccountEmail: 42}},
		{name: "email_invalid", data: map[grid.ColumnID]any{iam.ColAccountEmail: "broken"}},
		{name: "first_name_nil", data: map[grid.ColumnID]any{iam.ColAccountFirstName: nil}},
		{name: "status_invalid", data: map[grid.ColumnID]any{iam.ColAccountStatus: "garbage"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			side := newSide(t)
			_, err := iam.ApplyAccountPatch(side, tc.data)
			test.ErrIs(t, err, shared.ErrValidation)
		})
	}
}

func TestAccountColumns_Stable(t *testing.T) {
	// act
	cols := iam.AccountColumns()

	// assert
	if len(cols) != 6 {
		t.Fatalf("expected 6 columns, got %d", len(cols))
	}
	wantOrder := []grid.ColumnID{
		iam.ColAccountEmail,
		iam.ColAccountFirstName,
		iam.ColAccountLastName,
		iam.ColAccountStatus,
		iam.ColAccountPictureURL,
		iam.ColAccountID,
	}
	for i, w := range wantOrder {
		if cols[i].ID != w {
			t.Fatalf("expected column[%d]=%s, got %s", i, w, cols[i].ID)
		}
	}
}
