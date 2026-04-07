// Package apiv1 wires all proxy.core.v1 service handlers into a single
// http.Handler ready to be served over a Unix domain socket.
package apiv1

import (
	"net/http"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/ryfineZ/weave/gen/go/proxy/core/v1/corev1connect"
	"github.com/ryfineZ/weave/internal/log"
)

// Server holds all service handlers and exposes an http.Handler.
type Server struct {
	handler http.Handler
}

// NewServer creates a Server wiring all service handlers.
// Additional dependencies (store, orchestrator, etc.) are injected here
// as the implementation matures.
func NewServer(l *log.Logger, opts ...connect.HandlerOption) *Server {
	mux := http.NewServeMux()

	mux.Handle(corev1connect.NewRuntimeServiceHandler(NewRuntimeHandler(l), opts...))
	mux.Handle(corev1connect.NewNodeServiceHandler(NewNodeHandler(l), opts...))
	mux.Handle(corev1connect.NewChainServiceHandler(NewChainHandler(l), opts...))
	mux.Handle(corev1connect.NewRuleServiceHandler(NewRuleHandler(l), opts...))
	mux.Handle(corev1connect.NewSubscriptionServiceHandler(NewSubscriptionHandler(l), opts...))

	// h2c allows HTTP/2 without TLS, which is safe over a Unix domain socket.
	return &Server{handler: h2c.NewHandler(mux, &http2.Server{})}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
