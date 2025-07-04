package logutil

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"fmt"
	"log/slog"
)

func SlogActor(u domain.User) []any {
	logRoles := make([]any, len(u.Roles))
	for i, r := range u.Roles {
		logRoles[i] = slog.String(
			fmt.Sprintf("actor.roles[%d]", i),
			string(r),
		)
	}

	vals := make([]any, 0)

	vals = append(vals, slog.String("actor.id", string(u.ID)))
	vals = append(vals, logRoles...)

	return vals
}

func SlogUser(u domain.User) []any {
	logRoles := make([]any, len(u.Roles))
	for i, r := range u.Roles {
		logRoles[i] = slog.String(
			fmt.Sprintf("user.roles[%d]", i),
			string(r),
		)
	}

	vals := make([]any, 0)

	vals = append(vals, slog.String("user.id", string(u.ID)))
	vals = append(vals, logRoles...)

	return vals
}
