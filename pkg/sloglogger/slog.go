package sloglogger

import (
	"context"
	"log/slog"
	"os"
	"unitly/internal/log"
)

type slogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger wraps a native slog.Logger into our custom Logger interface.
func NewSlogLogger(options *slog.HandlerOptions) log.Logger {
	slogHandler := slog.NewJSONHandler(os.Stdout, options)
	logger := slog.New(slogHandler)
	return &slogLogger{logger: logger}
}

func (s *slogLogger) Debug(msg string, kv ...any) {
	s.logger.Debug(msg, kv...)
}

func (s *slogLogger) Info(msg string, kv ...any) {
	s.logger.Info(msg, kv...)
}

func (s *slogLogger) Warn(msg string, kv ...any) {
	s.logger.Warn(msg, kv...)
}

func (s *slogLogger) Error(msg string, kv ...any) {
	s.logger.Error(msg, kv...)
}

func (s *slogLogger) DebugCtx(ctx context.Context, msg string, kv ...any) {
	s.logger.DebugContext(ctx, msg, kv...)
}

func (s *slogLogger) InfoCtx(ctx context.Context, msg string, kv ...any) {
	s.logger.InfoContext(ctx, msg, kv...)
}

func (s *slogLogger) WarnCtx(ctx context.Context, msg string, kv ...any) {
	s.logger.WarnContext(ctx, msg, kv...)
}

func (s *slogLogger) ErrorCtx(ctx context.Context, msg string, kv ...any) {
	s.logger.ErrorContext(ctx, msg, kv...)
}

func (s *slogLogger) With(kv ...any) log.Logger {
	return &slogLogger{logger: s.logger.With(kv...)}
}
