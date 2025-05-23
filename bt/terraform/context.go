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

type componentContextKey string

const componentKey componentContextKey = "component"

func NewComponentContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, componentKey, value)
}

func ComponentFromContext(ctx context.Context) string {
	v, ok := ctx.Value(componentKey).(string)
	if ok {
		return v
	}
	return "."
}

type stackContextKey string

const stackKey stackContextKey = "stack"

func NewStackContext(ctx context.Context, value bool) context.Context {
	return context.WithValue(ctx, stackKey, value)
}

func StackFromContext(ctx context.Context) bool {
	v, ok := ctx.Value(stackKey).(bool)
	if ok {
		return v
	}
	return false
}
