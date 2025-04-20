package zapx

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(cfg zap.Config) *Logger {
	l, err := cfg.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return NewCustomCore(core)
	}))
	if err != nil {
		return nil
	}
	zap.RedirectStdLog(l)

	l = l.WithOptions(zap.AddCallerSkip(1))

	return &Logger{Logger: l}
}

func (l *Logger) Error(ctx context.Context, name string, msg string, fields ...zap.Field) {
	tid := extraceTraceId(ctx)

	fields = append([]zap.Field{zap.String("name", name), zap.String("trace_id", tid)}, fields...)

	l.Logger.Error(msg, fields...)
}

func (l *Logger) Info(ctx context.Context, name string, msg string, fields ...zap.Field) {
	tid := extraceTraceId(ctx)

	fields = append([]zap.Field{zap.String("name", name), zap.String("trace_id", tid)}, fields...)

	l.Logger.Info(msg, fields...)
}

func (l *Logger) Debug(ctx context.Context, name string, msg string, fields ...zap.Field) {
	tid := extraceTraceId(ctx)

	fields = append([]zap.Field{zap.String("name", name), zap.String("trace_id", tid)}, fields...)

	l.Logger.Debug(msg, fields...)
}

func extraceTraceId(ctx context.Context) string {
	spanCtx := oteltrace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return ""
	}

	return spanCtx.TraceID().String()
}

type CustomCore struct {
	zapcore.Core
}

func NewCustomCore(core zapcore.Core) *CustomCore {
	return &CustomCore{
		Core: core,
	}
}

func (z *CustomCore) Write(en zapcore.Entry, fields []zapcore.Field) error {
	for i, fd := range fields {
		if fd.Key == "phone" {
			phone := fd.String
			fields[i].String = phone[:3] + "****" + phone[7:]
		}
	}

	return z.Core.Write(en, fields)
}

func (z *CustomCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if z.Enabled(ent.Level) {
		return ce.AddCore(ent, z)
	}
	return ce
}
