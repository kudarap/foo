package telemetry

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

type Server struct {
	service string
}

func (s *Server) Middleware() func(http.Handler) http.Handler {
	return otelmux.Middleware(s.service)
}

// NewServerInstrumentation automatic instrumentation for server.
func NewServerInstrumentation(service string) *Server {
	return &Server{service: fmt.Sprintf("%s-server", service)}
}
