package http

import (
	"github.com/lastbackend/toolkit/pkg/server"
	"net/http"
)

const corsMiddlewareKind server.KindMiddleware = "corsMiddleware"

type corsMiddleware struct {
	handler http.HandlerFunc
}

func (corsMiddleware) Kind() server.KindMiddleware {
	return corsMiddlewareKind
}

func (s *corsMiddleware) Apply(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.handler(w, r)
		h.ServeHTTP(w, r)
	}
}
