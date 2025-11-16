package middleware

import "context"

type contextKey string

const (
	claimsKey    contextKey = "claims"
	validatorKey contextKey = "validator"
)

func contextWithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	val, ok := ctx.Value(claimsKey).(Claims)
	return val, ok
}

func contextWithValidator(ctx context.Context, v Validator) context.Context {
	return context.WithValue(ctx, validatorKey, v)
}

func ValidatorFromContext(ctx context.Context) (Validator, bool) {
	val, ok := ctx.Value(validatorKey).(Validator)
	return val, ok
}
