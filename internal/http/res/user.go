package res

import (
	"FaisalBudiono/go-boilerplate/internal/domain"
)

type user struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	PhoneNumber string   `json:"phoneNumber"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
}

func ToUser(u domain.User) response[user] {
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = string(r)
	}

	return response[user]{
		Data: user{
            ID: u.ID,
            Name: u.Name,
            PhoneNumber: u.PhoneNumber,
            Email: u.Email,
            Roles: roles,
        },
	}
}
