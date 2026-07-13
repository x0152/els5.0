package apps

import "strings"

type AccessPolicy struct {
	AllowAll    bool
	AllowRoles  []string
	AllowEmails []string
}

func (p AccessPolicy) Allows(email string, roles []string) bool {
	if p.AllowAll {
		return true
	}
	for _, e := range p.AllowEmails {
		if strings.EqualFold(strings.TrimSpace(e), email) {
			return true
		}
	}
	for _, r := range roles {
		for _, ar := range p.AllowRoles {
			if ar == r {
				return true
			}
		}
	}
	return false
}

type App struct {
	ID          string
	Name        string
	URI         string
	Description string
	Group       string
	Disabled    bool
	Access      AccessPolicy
}
