package middleware

import "context"

type ctxKey int

const apiKeyCtxKey ctxKey = 1

func WithAPIKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, apiKeyCtxKey, key)
}

func APIKeyFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(apiKeyCtxKey).(string); ok {
		return v
	}
	return ""
}