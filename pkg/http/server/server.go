package server

import (
	"context"
	"ctacampado/go-chi-service/pkg/logger"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

var (
	ErrMissingAddress = errors.New("Missing address")
	ErrMissingPort    = errors.New("Missing port")
)

// Server struct
type Server struct {
	name            string
	version         string
	projectID       string
	cors            *cors.Cors
	router          chi.Router
	shutdownTimeout time.Duration
}

type Option func(s *Server)

// New returns a new server
func New(name, version string, options ...Option) *Server {
	s := &Server{
		name:            name,
		version:         version,
		shutdownTimeout: 5 * time.Second,
	}

	for _, option := range options {
		option(s)
	}

	if s.router == nil {
		s.router = s.newRouter(time.Now())
	}

	return s
}

// Route allows setting up of routes
func (s *Server) Route(f func(chi.Router)) {
	f(s.router)
}

func (s *Server) String() string {
	name := s.name
	if name == "" {
		name = "unknown"
	}

	version := s.version
	if version == "" || version == "VERSION" {
		version = "v0.0.0"
	}

	return name + " (" + version + ")"
}

func (s *Server) Run(address string) error {
	if address == "" {
		return ErrMissingAddress
	}

	if !strings.Contains(address, ":") || strings.HasSuffix(address, ":") {
		return ErrMissingPort
	}

	l := logger.Default()

	svr := &http.Server{
		Addr:    address,
		Handler: s.router,
	}

	go func() {
		if err := svr.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				l.Fatal().Err(err).Send()
			}
		}
	}()

	l.Info().Msgf("running %s on %s", s.String(), address)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	l.Info().Msg("gracefully shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := svr.Shutdown(ctx); err != nil {
		l.Fatal().Err(err).Msg("cannot shutdown server")
	}

	l.Info().Msg("shut down")

	return nil
}
