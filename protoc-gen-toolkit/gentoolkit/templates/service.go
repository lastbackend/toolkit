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

package templates

// ServiceTpl is the service template used for new services.
var ServiceTpl = `
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
	{{- if not .HasNotServer }}
	SetServer(srv interface{})
	{{ end }}

	AddPackage(pkg interface{})
	AddMiddleware(mdw interface{})
	Invoke(fn interface{})
}

type RPC struct {
	Grpc grpc.Client
	{{- range $key, $value := .Clients }}
		{{ $value.Service | ToCamel }} {{ $key }}.{{ $value.Service | ToCamel }}RPCClient
	{{- end }}
}

func NewService(name string) Service {
	return &service{
		toolkit: toolkit.NewService(name),
		{{- if not .HasNotServer }}
		srv:     make([]interface{}, 0),
		{{- end }}
		pkg:     make([]interface{}, 0),
		inv:     make([]interface{}, 0),
		rpc:     new(RPC),
	}
}

type service struct {
	toolkit toolkit.Service
	rpc     *RPC
	{{- if not .HasNotServer }}
	srv    []interface{}
	{{- end }}
	pkg    []interface{}
	inv    []interface{}
	cfg    interface{}
	mdw    interface{}
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

{{ if not .HasNotServer }}
func (s *service) SetServer(srv interface{}) {
	if srv == nil {
		return 
	}
	s.srv = append(s.srv, srv)
}
{{ end }}

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

{{- range $type, $plugins := .Plugins }}
{{- range $index, $plugin := $plugins }}
type {{ $plugin.Prefix | ToCamel }}Plugin interface {
	{{ $type }}.Plugin
}
{{ end }}
{{ end }}

func (s *service) Run(ctx context.Context) error {
	
	{{ $count := 0 }}

	{{ range $type, $plugins := .Plugins }}
		{{- range $index, $plugin := $plugins }}
			plugin_{{ $count }} := {{ $plugin.Plugin }}.NewPlugin(s.toolkit, &{{ $plugin.Plugin }}.Options{Name: "{{ $plugin.Prefix | ToLower }}"})
			{{ $count = inc $count }}
		{{- end }}
	{{- end }}

	{{ $count = 0 }}

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
		{{- range $type, $plugins := .Plugins }}
			{{- range $index, $plugin := $plugins }}
				fx.Annotate(
					func() {{ $plugin.Prefix | ToCamel }}Plugin {
						return plugin_{{ $count }}
					},
				),
				{{ $count = inc $count }}	
			{{- end }}
		{{- end }}
	)

	provide = append(provide, s.pkg...)
	{{- if not .HasNotServer }}
	provide = append(provide, s.srv...)
	{{- end }}

	opts := make([]fx.Option, 0)
	
	if s.cfg != nil {
		opts = append(opts, fx.Supply(s.cfg))	
	}

	opts = append(opts, fx.Provide(provide...))
	opts = append(opts, fx.Invoke(s.registerClients))
	{{- if not $.HasNotServer }}
		{{- range $svc := .Services }}			
			opts = append(opts, fx.Invoke(s.register{{ $svc.GetName }}Server))
		{{- end }}
	{{- end }}
	opts = append(opts, fx.Invoke(s.inv...))
	if s.mdw != nil {
		opts = append(opts, fx.Invoke(s.mdw))
	}
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
	
	s.rpc.Grpc = s.toolkit.Client()

	{{ range $key, $value := .Clients }}
		s.rpc.{{ $value.Service | ToCamel }} = {{ $value.Service | ToLower }}.New{{ $value.Service | ToCamel }}RPCClient("{{ $value.Service | ToLower }}", s.rpc.Grpc)
	{{ end }}

	return nil
}

{{ if not .HasNotServer }}
	{{ $lengthService := len .Services }} 
	{{ range $svc := .Services }}
	func (s *service) register{{ $svc.GetName }}Server(srv {{ $svc.GetName }}RpcServer) error {
	
		// Register servers
	
		type {{ $svc.GetName }}GrpcRpcServer struct {
				{{ $svc.GetName }}Server
		}
	
		h := &{{ $svc.GetName | ToLower }}GrpcRpcServer{srv.({{ $svc.GetName }}RpcServer)}

		{{ if eq $lengthService 1 }}
		grpc{{ $svc.GetName | ToLower }}Server := server.NewServer(s.toolkit, &server.ServerOptions{Name: "grpc"})
		{{ else }}
		grpc{{ $svc.GetName | ToLower }}Server := server.NewServer(s.toolkit, &server.ServerOptions{Name: "{{ $svc.GetName | ToLower }}-grpc"})
		{{ end }}
		if err := grpc{{ $svc.GetName | ToLower }}Server.Register(&{{ $svc.GetName }}_ServiceDesc, &{{ $svc.GetName }}GrpcRpcServer{h}); err != nil {
			return err
		}
	
		if err := s.toolkit.ServerRegister(grpc{{ $svc.GetName | ToLower }}Server); err != nil {
			return err
		}
	
		return nil
	}
	{{ end }}
{{ end }}

func (s *service) registerRouter() {
	{{ range $svc := .Services }}
		{{ range $m := $svc.Methods }}
			{{ range $binding := $m.Bindings }}
				
				{{ if $binding.Websocket }}
				s.toolkit.Router().Handle(http.MethodGet, "{{ $binding.HttpPath }}", s.Router().ServerWS,
					router.HandleOptions{Middlewares: middlewares.getMiddleware("{{ $binding.RpcMethod }}")})
				{{ end }}
			
				{{ if or (not $binding.Websocket) ($binding.WebsocketProxy) }}
				s.toolkit.Router().Subscribe("{{ $binding.RpcMethod }}", func(ctx context.Context, event ws.Event, c *ws.Client) error {
					ctx, cancel := context.WithCancel(ctx)
					defer cancel()
			
					var protoRequest {{ $binding.RequestType.GoType $binding.Method.Service.File.GoPkg.Path }}
					var protoResponse {{ $binding.ResponseType.GoType $binding.Method.Service.File.GoPkg.Path }}
			
					if err := json.Unmarshal(event.Payload, &protoRequest); err != nil {
						return err
					}
			
					callOpts := make([]grpc.CallOption, 0)
		
					if headers := ctx.Value(ws.RequestHeaders); headers != nil {
						if v, ok := headers.(map[string]string); ok {
							callOpts = append(callOpts, grpc.Headers(v))
						}
					}
			
					if err := s.toolkit.Client().Call(ctx, "{{ $binding.Service }}", "{{ $binding.RpcPath }}", &protoRequest, &protoResponse, callOpts...); err != nil {
						return err	
					}
			
					return c.WriteJSON(protoResponse)
				})
		
				{{ if not $binding.WebsocketProxy }}
					s.toolkit.Router().Handle("{{ $binding.HttpMethod }}", "{{ $binding.HttpPath }}", func(w http.ResponseWriter, r *http.Request) {
						ctx, cancel := context.WithCancel(r.Context())
						defer cancel()
			
						var protoRequest {{ $binding.RequestType.GoType $binding.Method.Service.File.GoPkg.Path }}
						var protoResponse {{ $binding.ResponseType.GoType $binding.Method.Service.File.GoPkg.Path }}
				
						{{ if or (eq $binding.HttpMethod "http.MethodPost") (eq $binding.HttpMethod "http.MethodPut") (eq $binding.HttpMethod "http.MethodPatch") }}
							{{ if eq $binding.RawBody "*" }}
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
							{{ else }}
								_, om := router.GetMarshaler(s.toolkit.Router(), r)
			
								if err := router.SetRawBodyToProto(r, &protoRequest, "{{ $binding.RawBody }}"); err != nil {
									errors.HTTP.InternalServerError(w)
									return
								}
			
							{{ end }}
						{{ else }}
							_, om := router.GetMarshaler(s.toolkit.Router(), r)
			
							if err := r.ParseForm(); err != nil {
								errors.HTTP.InternalServerError(w)
								return
							}
				
							if err := router.ParseRequestQueryParametersToProto(&protoRequest, r.Form); err != nil {
								errors.HTTP.InternalServerError(w)
								return
							}
						{{ end }}
			
						{{ range $param := $binding.HttpParams }}
						if err := router.ParseRequestUrlParametersToProto(r, &protoRequest, "{{ $param | ToTrimRegexFromQueryParameter }}"); err != nil {
							errors.HTTP.InternalServerError(w)
							return
						}
						{{ end }}
				
						headers, err := router.PrepareHeaderFromRequest(r)
						if err != nil {
							errors.HTTP.InternalServerError(w)
							return
						}
			
						callOpts := make([]grpc.CallOption, 0)
						callOpts = append(callOpts, grpc.Headers(headers))
			
						if err := s.toolkit.Client().Call(ctx, "{{ $binding.Service }}", "{{ $binding.RpcPath }}", &protoRequest, &protoResponse, callOpts...); err != nil {
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
			
					}, router.HandleOptions{Middlewares: middlewares.getMiddleware("{{ $binding.RpcMethod }}")})
					{{ end }}
				{{ end }} 
			{{ end }} 
		{{ end }}
	{{ end }}
}

func registerMiddleware(name string, mdw ...func(h http.Handler) http.Handler) {
	if middlewares[name] == nil {
		middlewares[name] = make([]func(h http.Handler) http.Handler, 0)
	}
	for _, h := range mdw {
		middlewares[name] = append(middlewares[name], h)
	}
}

{{ range $svc := .Services }}
	{{ range $m := $svc.Methods }}
	func {{ $m.GetName }}MiddlewareAdd(mdw ...func(h http.Handler) http.Handler) {
		registerMiddleware("{{ $m.GetName }}", mdw...)
	}
	{{ end }}
{{ end }}

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
`
