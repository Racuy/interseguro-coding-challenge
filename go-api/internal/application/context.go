package application

import "context"

// avoids colliding with context keys from other packages
type contextKey int

const tokenKey contextKey = 0

// attaches a bearer token to ctx, for the gateway to read later
func ContextWithToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, tokenKey, token)
}

// reads the token back out, if any
func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey).(string)
	return token, ok
}
