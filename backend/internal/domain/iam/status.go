package iam

import "fmt"

type AccountStatus string

const (
	AccountStatusPendingPassword AccountStatus = "pending_password"
	AccountStatusActive          AccountStatus = "active"
	AccountStatusBlocked         AccountStatus = "blocked"
	AccountStatusNoAuth          AccountStatus = "no_auth"
)

func ParseAccountStatus(s string) (AccountStatus, error) {
	st := AccountStatus(s)
	if !st.IsValid() {
		return "", fmt.Errorf("invalid account status: %q", s)
	}
	return st, nil
}

func (s AccountStatus) IsValid() bool {
	switch s {
	case AccountStatusPendingPassword, AccountStatusActive, AccountStatusBlocked, AccountStatusNoAuth:
		return true
	}
	return false
}

func (s AccountStatus) String() string { return string(s) }
