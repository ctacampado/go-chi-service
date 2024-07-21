package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type ctxLogCtx int

const logCtxKey ctxLogCtx = 0

// FromContext creates a logger with logging context with values from ctx.
func FromContext(ctx context.Context) *zerolog.Logger {
	kvPairs, ok := ctx.Value(logCtxKey).(map[string]string)
	if !ok {
		return Default()
	}

	loggingCtx := Default().With()
	for k, v := range kvPairs {
		loggingCtx = loggingCtx.Str(k, v)
	}

	l := loggingCtx.Logger()
	return &l
}

// AddToLoggingContext adds key-value pair to logging context.
func AddToLoggingContext(
	ctx context.Context,
	key, value string,
) context.Context {
	kvPairs, ok := ctx.Value(logCtxKey).(map[string]string)
	if !ok {
		kvPairs = make(map[string]string)
	}
	kvPairs[key] = value

	return context.WithValue(ctx, logCtxKey, kvPairs)
}
