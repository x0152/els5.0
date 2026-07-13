package admin

import (
	"github.com/els/backend/internal/domain/iam"
)

// Filter limits the administrator query. The base template has no
// multi-tenancy (clients), so visibility is all-or-nothing:
// only a global admin may view administrators.
type Filter struct {
	deny bool
}

func (f Filter) IsDeny() bool { return f.deny }

func VisibilityFor(actor *iam.Actor) Filter {
	if actor == nil {
		return Filter{deny: true}
	}
	if actor.IsGlobalAdmin() {
		return Filter{}
	}
	return Filter{deny: true}
}
