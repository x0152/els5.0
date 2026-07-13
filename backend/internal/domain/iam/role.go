package iam

import "fmt"

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleExpert   Role = "expert"
	RoleCustomer Role = "customer"
)

func ParseRole(s string) (Role, error) {
	r := Role(s)
	if !r.IsValid() {
		return "", fmt.Errorf("invalid role: %q", s)
	}
	return r, nil
}

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleExpert, RoleCustomer:
		return true
	}
	return false
}

func (r Role) String() string { return string(r) }
