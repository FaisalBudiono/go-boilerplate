package domain

import "slices"

type Role string

func (r Role) String() string  { return string(r) }
func (r Role) IsValid() bool   { return slices.Contains(roles, r) }
func (r Role) Options() []Role { return roles }

var roles = []Role{
	RoleAdmin,
}

const (
	RoleAdmin Role = "admin"
)
