package spanattr

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

func Actor(prefix string, u domain.User) []attribute.KeyValue {
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = string(r)
	}

	return []attribute.KeyValue{
		attribute.String(fmt.Sprintf("%sactor.id", prefix), string(u.ID)),
		attribute.StringSlice(fmt.Sprintf("%sactor.roles", prefix), rolesToString(u.Roles)),
	}
}

func User(prefix string, u domain.User) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(fmt.Sprintf("%suser.id", prefix), string(u.ID)),
		attribute.StringSlice(fmt.Sprintf("%suser.roles", prefix), rolesToString(u.Roles)),
	}
}

func rolesToString(raws []domain.Role) []string {
	roles := make([]string, len(raws))
	for i, r := range raws {
		roles[i] = string(r)
	}

	return roles
}
