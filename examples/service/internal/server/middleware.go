package server

import (
	"context"
	"fmt"
	"github.com/lastbackend/toolkit/pkg/server"
	"net/http"
)

type Data string

const MWAuthenticate server.KindMiddleware = "mwauthenticate"

func ExampleHTTPServerMiddleware1(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Call: ExampleMiddleware")

		// Set example data to request context
		ctx := context.WithValue(r.Context(), Data("test-data"), "example context data")

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ExampleHTTPServerMiddleware2(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Call: ExampleMiddleware")

		// Set example data to request context
		ctx := context.WithValue(r.Context(), Data("test-data"), "example context data")

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
