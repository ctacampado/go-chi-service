package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/cors"
)

// WithCors sets cors origins
func WithCors(allowedOrigins ...string) Option {
	return func(s *Server) {
		s.cors = cors.New(cors.Options{
			AllowedOrigins: allowedOrigins,
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPatch,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
				http.MethodOptions,
			},
			AllowedHeaders: []string{
				"Accept",
				"Authorization",
				"Content-Type",
			},
			AllowCredentials: true,
			MaxAge:           300,
		})
	}
}

func isOriginAllowed(allowedOrigins []string, origin string) bool {
	origin = strings.TrimPrefix(origin, "https://")
	for _, val := range allowedOrigins {
		if val == origin {
			return true
		}
	}
	return false
}

func contains(a []string, b string) bool {
	for _, val := range a {
		if val == b {
			return true
		}
	}
	return false
}

func CORSHeaders(
	allowedOrigins []string,
	originHost, originScheme string,
) func(http.Handler) http.Handler {
	allowedOriginsAll := contains(allowedOrigins, "*")
	allowedOriginsCount := len(allowedOrigins)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && originHost != "" && originScheme != "" {
				u, err := url.Parse(origin)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).
						Encode(HTTPError{Code: http.StatusBadRequest, Err: err})
					return
				}

				if u.Scheme != originScheme {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).
						Encode(HTTPError{Code: http.StatusBadRequest, Err: fmt.Errorf("only https origins are allowed")})
					return
				}

				if !strings.HasSuffix(u.Host, originHost) {
					w.WriteHeader(http.StatusUnauthorized)
					_ = json.NewEncoder(w).
						Encode(HTTPError{Code: http.StatusUnauthorized, Err: fmt.Errorf("unauthorized http origin header")})
					return
				}
			}

			if allowedOriginsCount == 0 || origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if r.Method == http.MethodOptions &&
				(allowedOriginsAll || isOriginAllowed(allowedOrigins, origin)) {
				if allowedOriginsAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().
					Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
				w.Header().
					Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
				w.Header().Set("Access-Control-Expose-Headers", "Authorization")
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bodyCloser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		next.ServeHTTP(w, r)
	})
}
