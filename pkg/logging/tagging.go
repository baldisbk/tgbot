package logging

import (
	"context"

	"go.uber.org/zap"
)

type tagKey struct {
	Name string
}

func WithTag(ctx context.Context, name, value string) context.Context {
	newCtx := context.WithValue(ctx, tagKey{Name: name}, value)
	if logger, ok := LCheck(ctx); ok {
		newCtx = WithLogger(newCtx, logger.With(zap.String(name, value)))
	}
	return newCtx
}

func Tag(ctx context.Context, name string) string {
	if value := ctx.Value(tagKey{Name: name}); value == nil {
		return ""
	} else {
		return value.(string)
	}
}
