package perm

import "FaisalBudiono/go-boilerplate/internal/domain"

func IsAdmin(u domain.User) bool {
	for _, r := range u.Roles {
		if r == domain.RoleAdmin {
			return true
		}
	}

	return false
}
