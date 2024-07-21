package logger

import (
	"context"
	"net/http"
	"sync"
	"time"

	"ctacampado/go-chi-service/pkg/logger"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func RequestLogger(h http.Handler) http.Handler {
	return middleware.RequestLogger(&formatter{})(h)
}

type formatter struct {
}

func (f *formatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	l := logger.FromContext(r.Context())

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.
			Str("requestMethod", r.Method).
			Str("host", r.Host).
			Str("requestUrl", r.RequestURI).
			Str("remoteIp", r.RemoteAddr)
	})

	return &entry{req: r, logger: l}
}

type entry struct {
	req    *http.Request
	logger *zerolog.Logger
	mu     sync.Mutex
}

func (e *entry) Write(
	status int,
	bytes int,
	header http.Header,
	elapsed time.Duration,
	extra interface{},
) {

	// status might not be explicitly set
	if status == 0 {
		status = http.StatusOK
	}

	loggingCtx := e.logger.With()

	l := loggingCtx.Logger()
	l.Debug().
		Int("status", status).
		Dur("latency", elapsed).
		Int("responseSize(bytes)", bytes).
		Send()
}

func (e *entry) Panic(v interface{}, stack []byte) {
	e.logger.Error().Caller().Msgf("panic: %+v", v)
}

func (e *entry) AddToRequestLoggingContext(key, value string) {
	// in most cases entry would be associated with a single goroutine,
	// however there's nothing preventing access by multiple goroutines, so using mutex for safety.
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str(key, value)
	})
}

func AddToRequestLoggingContext(r *http.Request, key, value string) {
	e, ok := middleware.GetLogEntry(r).(*entry)
	if ok {
		e.AddToRequestLoggingContext(key, value)
	}
}

// FromContext is a convenience function to be used when the package is already imported.
func FromContext(ctx context.Context) *zerolog.Logger {
	return logger.FromContext(ctx)
}
