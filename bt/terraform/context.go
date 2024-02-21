package terraform

import "context"

type invalidateCache string

const invalidateCacheKey invalidateCache = "invalidateCache"

func invalidateCacheContext(ctx context.Context, value bool) context.Context {
	return context.WithValue(ctx, invalidateCacheKey, value)
}

func invalidateCacheFromContext(ctx context.Context) bool {
	v, ok := ctx.Value(invalidateCacheKey).(bool)
	if ok {
		return v
	}
	return false
}

type dirContextKey string

const dirKey dirContextKey = "dir"

func NewDirContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, dirKey, value)
}

func DirFromContext(ctx context.Context) string {
	v, ok := ctx.Value(dirKey).(string)
	if ok {
		return v
	}
	return "."
}
