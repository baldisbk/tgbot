package logging

import (
	"context"

	"go.uber.org/zap"
)

type logKey struct{}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, logKey{}, logger)
}

func L(ctx context.Context) *zap.Logger {
	if value := ctx.Value(logKey{}); value == nil {
		return zap.NewNop()
	} else {
		return value.(*zap.Logger)
	}
}

func LCheck(ctx context.Context) (*zap.Logger, bool) {
	if value := ctx.Value(logKey{}); value == nil {
		return nil, false
	} else {
		return value.(*zap.Logger), true
	}
}

func S(ctx context.Context) *zap.SugaredLogger {
	return L(ctx).Sugar()
}
