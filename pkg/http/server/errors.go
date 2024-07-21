package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"ctacampado/go-chi-service/pkg/logger"
)

// http errors
var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrInternalServerError = errors.New("internal server error")

	ErrBodyNotJSON = fmt.Errorf("body not valid json: %w", ErrBadRequest)
)

// compile time check to ensure HTTPError always implements StatusError
var _ StatusError = (*HTTPError)(nil)

// StatusError composes error and a method to return a status code
type StatusError interface {
	error
	Status() int
	JSON() string
	MarshalJSON() ([]byte, error)
}

// HTTPError implements StatusError, returns an error and a HTTP status code
type HTTPError struct {
	Code int
	Err  error
}

// Error implements the error interface for HTTPError
func (s HTTPError) Error() string {
	if s.Err != nil {
		err := s.Err

		if errors.Is(err, ErrBadRequest) ||
			errors.Is(err, ErrUnauthorized) ||
			errors.Is(err, ErrForbidden) ||
			errors.Is(err, ErrNotFound) ||
			errors.Is(err, ErrInternalServerError) {

			// instead of seeing "oops: internal server error", you would only see "oops".
			sections := strings.Split(err.Error(), ":")
			suffix := strings.TrimSpace(sections[len(sections)-1])

			// trim suffix from HTTP errors
			return strings.TrimSuffix(s.Err.Error(), ": "+suffix)
		}

		return s.Err.Error()

	}

	if status := s.Status(); status >= 400 {
		return http.StatusText(s.Status())
	}

	return ""
}

// Status implements the StatusError interface for HTTPError
func (s HTTPError) Status() int {
	return s.Code
}

type (
	// ErrorField contains details of an error
	ErrorField struct {
		Detail string `json:"detail"`
	}
	// ErrorResponse contains a set of ErrorFields
	ErrorResponse struct {
		Errors []ErrorField `json:"errors"`
	}
)

func (er *ErrorResponse) String() string {
	if len(er.Errors) == 0 {
		return ""
	}

	return er.Errors[0].Detail
}

// JSON view of HTTPError
func (s HTTPError) JSON() string {
	j, _ := s.MarshalJSON()
	return string(j)
}

// MarshalJSON implements JSON encoder
func (s HTTPError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&ErrorResponse{
		Errors: []ErrorField{
			{
				Detail: s.Error(),
			},
		},
	})
}

// ErrorHandler returns an error from a http handler
type ErrorHandler func(w http.ResponseWriter, r *http.Request) error

// ServerHTTP implements the http.Handler interface, checks for an error and parses it if it is a StatusError
func (h ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// response is always application/json
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := h(w, r); err != nil {
		status := getStatus(err)
		log(r.Context(), err, status)

		switch err.(type) {
		case StatusError:
			log(r.Context(), err, status)
		default:
			log(r.Context(), err, status)
			err = HTTPError{Code: status, Err: err}
		}

		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(err.(StatusError))
	}
}

func getStatus(err error) int {
	if e, ok := err.(HTTPError); ok {
		return e.Status()
	}

	if errors.Is(err, ErrBadRequest) {
		return http.StatusBadRequest
	}

	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}

	if errors.Is(err, ErrForbidden) {
		return http.StatusForbidden
	}

	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

func log(ctx context.Context, err error, status int) {
	l := logger.FromContext(ctx)

	if status >= http.StatusInternalServerError {
		l.Error().Err(err).Send()
	} else {
		l.Warn().Err(err).Send()
	}
}

func NewHTTPBadRequestError(err error) error {
	return HTTPError{
		Code: http.StatusBadRequest,
		Err:  err,
	}
}
