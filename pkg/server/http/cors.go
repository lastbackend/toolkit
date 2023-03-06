package http

import (
	"github.com/lastbackend/toolkit/pkg/server"
	"net/http"
)

type corsMiddleware struct {
	handler http.HandlerFunc
}

func (corsMiddleware) Kind() server.KindMiddleware {
	return "cors"
}

func (s *corsMiddleware) Apply(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.handler(w, r)
		h.ServeHTTP(w, r)
	}
}
