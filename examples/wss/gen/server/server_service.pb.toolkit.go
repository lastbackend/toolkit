// Code generated by protoc-gen-toolkit. DO NOT EDIT.
// source: github.com/lastbackend/toolkit/examples/wss/apis/server.proto

package serverpb

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	toolkit "github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/helloworld/gen"
	grpc "github.com/lastbackend/toolkit/pkg/client/grpc"
	logger "github.com/lastbackend/toolkit/pkg/logger"
	router "github.com/lastbackend/toolkit/pkg/router"
	errors "github.com/lastbackend/toolkit/pkg/router/errors"
	ws "github.com/lastbackend/toolkit/pkg/router/ws"
	server "github.com/lastbackend/toolkit/pkg/server"
	fx "go.uber.org/fx"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// Suppress "imported and not used" errors
var _ context.Context
var _ logger.Logger
var _ emptypb.Empty
var _ server.Server
var _ grpc.Client
var _ http.Handler
var _ errors.Err
var _ io.Reader
var _ json.Marshaler
var _ ws.Client

type middleware map[string][]func(h http.Handler) http.Handler

func (m middleware) getMiddleware(name string) router.Middleware {
	middleware := router.Middleware{}
	if m[name] != nil {
		for _, mdw := range m[name] {
			middleware.Add(mdw)
		}
	}
	return middleware
}

var middlewares = make(middleware, 0)

type Service interface {
	Logger() logger.Logger
	Meta() toolkit.Meta
	CLI() toolkit.CLI
	Client() grpc.Client
	Router() router.Server
	Run(ctx context.Context) error

	SetConfig(cfg interface{})

	AddPackage(pkg interface{})
	AddMiddleware(mdw interface{})
	Invoke(fn interface{})
}

type RPC struct {
	Grpc grpc.RPCClient
}

func NewService(name string) Service {
	return &service{
		toolkit: toolkit.NewService(name),
		pkg:     make([]interface{}, 0),
		inv:     make([]interface{}, 0),
		rpc:     new(RPC),
	}
}

type service struct {
	toolkit toolkit.Service
	rpc     *RPC
	pkg     []interface{}
	inv     []interface{}
	cfg     interface{}
	mdw     interface{}
}

func (s *service) Meta() toolkit.Meta {
	return s.toolkit.Meta()
}

func (s *service) CLI() toolkit.CLI {
	return s.toolkit.CLI()
}

func (s *service) Logger() logger.Logger {
	return s.toolkit.Logger()
}

func (s *service) Router() router.Server {
	return s.toolkit.Router()
}

func (s *service) Client() grpc.Client {
	return s.toolkit.Client()
}

func (s *service) SetConfig(cfg interface{}) {
	s.cfg = cfg
}

func (s *service) AddPackage(pkg interface{}) {
	if pkg == nil {
		return
	}
	s.pkg = append(s.pkg, pkg)
}

func (s *service) AddMiddleware(mdw interface{}) {
	s.mdw = mdw
}

func (s *service) Invoke(fn interface{}) {
	if fn == nil {
		return
	}
	s.inv = append(s.inv, fn)
}

func (s *service) Run(ctx context.Context) error {

	provide := make([]interface{}, 0)
	provide = append(provide,
		fx.Annotate(
			func() toolkit.Service {
				return s.toolkit
			},
		),
		func() Service {
			return s
		},
		func() *RPC {
			return s.rpc
		},
	)

	provide = append(provide, s.pkg...)

	opts := make([]fx.Option, 0)

	if s.cfg != nil {
		opts = append(opts, fx.Supply(s.cfg))
	}

	opts = append(opts, fx.Provide(provide...))
	opts = append(opts, fx.Invoke(s.registerClients))
	opts = append(opts, fx.Invoke(s.inv...))
	opts = append(opts, fx.Invoke(s.mdw))
	opts = append(opts, fx.Invoke(s.registerRouter))
	opts = append(opts, fx.Invoke(s.runService))

	app := fx.New(
		fx.Options(opts...),
		fx.NopLogger,
	)

	if err := app.Start(ctx); err != nil {
		return err
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, shutdownSignals...)

	select {
	// wait on kill signal
	case <-signalCh:
	// wait on context cancel
	case <-ctx.Done():
	}

	return app.Stop(ctx)
}

func (s *service) registerClients() error {

	// Register clients

	s.rpc.Grpc = grpc.NewClient(s.toolkit.CLI(), grpc.ClientOptions{Name: "client-grpc"})

	if err := s.toolkit.ClientRegister(s.rpc.Grpc); err != nil {
		return err
	}

	return nil
}

func (s *service) registerRouter() {

	s.toolkit.Router().Handle(http.MethodGet, "/events", s.Router().ServerWS,
		router.HandleOptions{Middlewares: middlewares.getMiddleware("Subscribe")})

	s.toolkit.Router().Subscribe("HelloWorld", func(ctx context.Context, event ws.Event, c *ws.Client) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var protoRequest servicepb.HelloRequest
		var protoResponse servicepb.HelloReply

		if err := json.Unmarshal(event.Payload, &protoRequest); err != nil {
			return err
		}

		callOpts := make([]grpc.CallOption, 0)

		if headers := ctx.Value(ws.RequestHeaders); headers != nil {
			if v, ok := headers.(map[string]string); ok {
				callOpts = append(callOpts, grpc.Headers(v))
			}
		}

		if err := s.toolkit.Client().Call(ctx, "helloworld", "/helloworld.Greeter/SayHello", &protoRequest, &protoResponse, callOpts...); err != nil {
			return err
		}

		return c.WriteJSON(protoResponse)
	})

	s.toolkit.Router().Handle(http.MethodPost, "/hello", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		var protoRequest servicepb.HelloRequest
		var protoResponse servicepb.HelloReply

		im, om := router.GetMarshaler(s.toolkit.Router(), r)

		reader, err := router.NewReader(r.Body)
		if err != nil {
			errors.HTTP.InternalServerError(w)
			return
		}

		if err := im.NewDecoder(reader).Decode(&protoRequest); err != nil && err != io.EOF {
			errors.HTTP.InternalServerError(w)
			return
		}

		headers, err := router.PrepareHeaderFromRequest(r)
		if err != nil {
			errors.HTTP.InternalServerError(w)
			return
		}

		callOpts := make([]grpc.CallOption, 0)
		callOpts = append(callOpts, grpc.Headers(headers))

		if err := s.toolkit.Client().Call(ctx, "helloworld", "/helloworld.Greeter/SayHello", &protoRequest, &protoResponse, callOpts...); err != nil {
			errors.GrpcErrorHandlerFunc(w, err)
			return
		}

		buf, err := om.Marshal(protoResponse)
		if err != nil {
			errors.HTTP.InternalServerError(w)
			return
		}

		w.Header().Set("Content-Type", om.ContentType())

		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(buf); err != nil {
			s.toolkit.Logger().Infof("Failed to write response: %v", err)
			return
		}

	}, router.HandleOptions{Middlewares: middlewares.getMiddleware("HelloWorld")})

}

func registerMiddleware(name string, mdw ...func(h http.Handler) http.Handler) {
	if middlewares[name] == nil {
		middlewares[name] = make([]func(h http.Handler) http.Handler, 0)
	}
	for _, h := range mdw {
		middlewares[name] = append(middlewares[name], h)
	}
}

func SubscribeMiddlewareAdd(mdw ...func(h http.Handler) http.Handler) {
	registerMiddleware("Subscribe", mdw...)
}

func HelloWorldMiddlewareAdd(mdw ...func(h http.Handler) http.Handler) {
	registerMiddleware("HelloWorld", mdw...)
}

func (s *service) runService(lc fx.Lifecycle) error {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return s.toolkit.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return s.toolkit.Stop(ctx)
		},
	})
	return nil
}

var shutdownSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGKILL,
}
