package server

import (
	"encoding/json"
	"net/http"
	"time"

	"ctacampado/go-chi-service/pkg/http/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func health(msgs ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		for _, msg := range msgs {
			_, _ = w.Write([]byte(msg))
		}
	}
}

func version(version string, now time.Time) http.HandlerFunc {
	rsp := struct {
		Version   string    `json:"version,omitempty"`
		StartedAt time.Time `json:"started_at"`
	}{
		Version:   version,
		StartedAt: now,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(rsp)
	}
}

func (s *Server) newRouter(now time.Time) chi.Router {
	router := chi.NewRouter()

	if s.cors != nil {
		router.Use(s.cors.Handler)
	}

	compressionLevel := 5 // recommended by chi

	router.Use(
		middleware.RequestID,
		middleware.StripSlashes,
		middleware.RealIP,
		bodyCloser,
		middleware.Compress(compressionLevel),
		logger.RequestLogger,
	)

	router.Method(
		http.MethodGet,
		"/",
		version(s.version, now.Truncate(time.Second)),
	)
	router.Method(http.MethodGet, "/health", health("OK"))
	router.Method(http.MethodHead, "/health", health())

	return router
}
