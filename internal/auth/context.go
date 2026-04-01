package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const userIDKey contextKey = "userID"

func WithUserID(ctx context.Context, userID pgtype.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) (pgtype.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(pgtype.UUID)
	return id, ok
}
