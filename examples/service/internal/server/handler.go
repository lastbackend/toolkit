package server

import "net/http"

// ExampleHTTPServerHandler server definitions
func ExampleHTTPServerHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
