package reqctx

import "context"

type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyUser
	keySilent
)

type User struct {
	ID    string
	Email string
	Role  string
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(keyRequestID).(string); ok {
		return v
	}
	return ""
}

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, keyUser, u)
}

func UserOf(ctx context.Context) (User, bool) {
	if ctx == nil {
		return User{}, false
	}
	u, ok := ctx.Value(keyUser).(User)
	return u, ok
}

func WithSilent(ctx context.Context) context.Context {
	return context.WithValue(ctx, keySilent, true)
}

func IsSilent(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v, ok := ctx.Value(keySilent).(bool)
	return ok && v
}
