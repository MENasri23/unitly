package log

import "context"

// Logger defines the behavior your application expects for structured logging.
type Logger interface {
	Debug(msg string, keysVals ...any)
	Info(msg string, keysVals ...any)
	Warn(msg string, keysVals ...any)
	Error(msg string, keysVals ...any)

	// Context-aware logging (highly recommended for tracing/OpenTelemetry)
	DebugCtx(ctx context.Context, msg string, keysVals ...any)
	InfoCtx(ctx context.Context, msg string, keysVals ...any)
	WarnCtx(ctx context.Context, msg string, keysVals ...any)
	ErrorCtx(ctx context.Context, msg string, keysVals ...any)

	// With returns a new logger with the given attributes permanently attached
	With(keysVals ...any) Logger
}
