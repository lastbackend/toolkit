/*
Copyright [2014] - [2023] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	servicepb "github.com/lastbackend/toolkit/examples/wss/gen/server"
	"github.com/lastbackend/toolkit/examples/wss/middleware"
	"github.com/lastbackend/toolkit/pkg/runtime"
	tk_http "github.com/lastbackend/toolkit/pkg/server/http"
	"github.com/lastbackend/toolkit/pkg/server/http/websockets"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, `{"alive": true}`); err != nil {
		fmt.Println(err)
	}
}

func TestWSHandler(ctx context.Context, event websockets.Event, c *websockets.Client) error {
	fmt.Println("Event:", event.Type, string(event.Payload))
	fmt.Println("Context:", ctx.Value("test-data"))
	return c.WriteMessage(websockets.TextMessage, event.Payload)
}

func main() {

	// define service with name and options
	app, err := servicepb.NewRouterService("wss",
		runtime.WithVersion("0.1.0"),
		runtime.WithDescription("Example router microservice"),
		runtime.WithEnvPrefix("WSS"),
	)
	if err != nil {
		fmt.Println(err)
	}

	// Logger settings
	app.Log().Info("Run microservice")

	// Add middleware
	app.Server().HTTP().SetMiddleware(middleware.RegisterExampleMiddleware)
	app.Server().HTTP().SetMiddleware(middleware.RegisterRequestID)

	// set middleware as global middleware
	app.Server().HTTP().UseMiddleware("request_id")
	app.Server().HTTP().Subscribe("event:name", TestWSHandler)

	// add handler to default http server
	app.Server().HTTP().
		AddHandler(http.MethodGet, "/health", HealthCheckHandler, tk_http.WithMiddleware("example"))

	// Service run
	if err := app.Start(context.Background()); err != nil {
		app.Log().Errorf("could not run the service %v", err)
		os.Exit(1)
		return
	}

	app.Log().Info("graceful stop")
}
