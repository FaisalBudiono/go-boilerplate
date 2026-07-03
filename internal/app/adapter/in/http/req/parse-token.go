package req

import (
	"context"
	"errors"

	"komdigi-immigration/internal/app/domain"
)

type contextKey string

const (
	contextKeyUser contextKey = "api:http:user"
)

// ContextWithUser set authenticated user to context
func ContextWithUser(ctx context.Context, user *domain.Userinfo) context.Context {
	return context.WithValue(ctx, contextKeyUser, user)
}

// AuthUser get authenticated user from context. Need [http.AuthMiddleware]
// to be called in the chain.
func AuthUser(ctx context.Context) (*domain.Userinfo, error) {
	user, ok := ctx.Value(contextKeyUser).(*domain.Userinfo)
	if !ok || user == nil {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}
