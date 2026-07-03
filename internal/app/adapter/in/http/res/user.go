package res

import (
	"time"

	"FaisalBudiono/go-boilerplate/internal/app/domain"
)

type user struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

func newUser(u domain.UserEagerLoad) user {
	roles := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		roles[i] = string(role)
	}

	return user{
		ID:        u.User.ID,
		Name:      u.User.Name,
		Email:     u.User.Email,
		Roles:     roles,
		CreatedAt: u.User.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.User.UpdatedAt.Format(time.RFC3339),
	}
}

func User(u domain.UserEagerLoad) response[user] {
	return response[user]{
		Data: newUser(u),
	}
}

func Users(
	users []domain.UserEagerLoad, pg domain.Pagination,
) responsePaginated[user] {
	res := make([]user, len(users))
	for i, u := range users {
		res[i] = newUser(u)
	}

	return responsePaginated[user]{
		response: response[[]user]{
			Data: res,
		},
		Meta: newPaginatedMeta(pg),
	}
}
