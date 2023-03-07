package server

import (
	"context"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	typespb "github.com/lastbackend/toolkit/examples/service/gen/ptypes"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/server"
	"net/http"
)

type Data string

const MWAuthenticate server.KindMiddleware = "mwauthenticate"

type ExampleHTTPServerMiddleware struct {
	server.DefaultHttpServerMiddleware
	name    server.KindMiddleware
	runtime runtime.Runtime
	service servicepb.ExampleServices
}

func (e *ExampleHTTPServerMiddleware) Apply(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e.runtime.Log().Info("Call: ExampleMiddleware")
		resp, _ := e.service.Example().HelloWorld(context.Background(), &typespb.HelloWorldRequest{
			Name: "middleware",
			Type: "test",
			Data: nil,
		})

		e.runtime.Log().Info("middleware resp response:> ", resp.Name)
		// Set example data to request context
		ctx := context.WithValue(r.Context(), Data("test-data"), "example context data")

		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (e *ExampleHTTPServerMiddleware) Kind() server.KindMiddleware {
	return e.name
}

func RegisterExampleHTTPServerMiddleware(runtime runtime.Runtime, service servicepb.ExampleServices) server.HttpServerMiddleware {
	return &ExampleHTTPServerMiddleware{name: MWAuthenticate, runtime: runtime, service: service}
}
