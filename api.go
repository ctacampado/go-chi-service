package main

import (
	"context"
	"ctacampado/go-chi-service/pkg/http/server"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	*server.Server
}

func NewServer(ctx context.Context) (*Server, error) {
	// TODO: set allowedOrigins from config
	allowedOrigins := []string{}

	s := server.New(svcName, ver, server.WithCors(allowedOrigins...))

	s.Route(func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// TODO: set corsOriginHost (ex: *.example.com) and corsOriginScheme (ex: https) from config
			r.Use(server.CORSHeaders(allowedOrigins, "", ""))

			r.Method(http.MethodGet, "/", Greeter())
		})
	})

	return &Server{Server: s}, nil
}

func Greeter() server.ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
		return nil
	}
}
