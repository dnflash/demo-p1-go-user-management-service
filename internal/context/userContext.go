package context

import (
	"context"
	"errors"
)

type userContextKey struct{}
type UserContext struct {
	UserID string
	Role   string
}

func SetUserContext(ctx context.Context, uc UserContext) context.Context {
	return context.WithValue(ctx, userContextKey{}, uc)
}

func GetUserContext(ctx context.Context) (UserContext, error) {
	uc, ok := ctx.Value(userContextKey{}).(UserContext)
	if !ok {
		return uc, errors.New("failed to get UserContext")
	}
	return uc, nil
}
