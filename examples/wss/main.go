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
	logger2 "github.com/lastbackend/toolkit/pkg/runtime/logger"
	"io"
	"net/http"
	"os"

	pb "github.com/lastbackend/toolkit/examples/wss/gen/server"
	"github.com/lastbackend/toolkit/examples/wss/middleware"
	"github.com/lastbackend/toolkit/pkg/http"
	"github.com/lastbackend/toolkit/pkg/http/ws"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, `{"alive": true}`); err != nil {
		fmt.Println(err)
	}
}

func TestWSHandler(ctx context.Context, event ws.Event, c *ws.Client) error {
	fmt.Println("Event:", event.Type, string(event.Payload))
	fmt.Println("Context:", ctx.Value("test-data"))
	return c.WriteMessage(ws.TextMessage, event.Payload)
}

func main() {
	log := logger2.DefaultLogger
	opts := log.Options()
	opts.Level = logger2.DebugLevel
	opts.VerboseLevel = logger2.DebugLevel
	log.Init(opts)
	log = log.WithFields(logger2.Fields{
		"service": "wss",
	})

	log.Infof("Start process")

	svc := pb.NewService("wss")
	svc.Meta().SetEnvPrefix("WSS")
	svc.AddMiddleware(middleware.New)

	svc.Router().Subscribe("event:name", TestWSHandler)

	svc.Router().Handle(http.MethodGet, "/health", HealthCheckHandler, http.HandleOptions{})

	if err := svc.Run(context.Background()); err != nil {
		os.Exit(1)
		return
	}
}
