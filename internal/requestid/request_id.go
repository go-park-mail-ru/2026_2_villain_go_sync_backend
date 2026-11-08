package requestid

import "context"

type contextKey struct{}

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

func FromContext(ctx context.Context) string {
	id, ok := ctx.Value(contextKey{}).(string)
	if !ok {
		return ""
	}
	return id
}
