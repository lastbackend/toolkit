# Last.Backend Toolkit 
[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/lastbackend/toolkit?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/lastbackend/toolkit)](https://goreportcard.com/report/lastbackend/toolkit)
[![Sourcegraph](https://sourcegraph.com/github.com/lastbackend/toolkit/-/badge.svg)](https://sourcegraph.com/github.com/lastbackend/toolkit?badge)

**lastbackend:toolkit** is a **programming toolkit** for building microservices, services
(or elegant monoliths) in Go. We provide you base application modular bolierpate with plugins so you can focus on delivering
business value.

## Communication

[![](https://dcbadge.vercel.app/api/server/WhK9ujvem9)](https://discord.gg/WhK9ujvem9)

## Overview

Last.Backend provides the modular boilerplate with plugins and package management, developed for distributed systems development including RPC and Event driven communication.
We provide basics to get you started as quickliy as possible.

## Features

Toolkit abstracts away the details of distributed systems. Here are the main features.

- **Configs** - Register your own config in main.go and use it from anywhere. The config interface provides a way to load application level config from env vars.

- **Plugins** - Rich plugin system to reuse same code. Plugins are designed as connectors to user services like databases, brokers, caches, etc...

- **Packages** - Automatic packages system for custom packages logic. Like repositories, controllers and other stuff. Can be powerupped with Hooks: <PreStart, OnStart, OnStop> to customize applciation logic.

- **Services/Clients** - Define another services clients and use it anywhere.

- **Logging** - One loggins system available across all applications with different log leves.

- **RPC/HTTP Server** - GRPC and HTTP servers in the box. Enabled by request in proto file.

- **Usefull help** - Print rich usefull helm information how to operate with application

## Preparation

Run go mod tidy to resolve the versions. Install all dependencies by running
[Configure](https://github.com/lastbackend/toolkit/tree/master/hack/hack/bootstrap.sh)

Download annotations and place them to: 
`$GOPATH/grpc/annotations`

Annotations: [toolkit](https://github.com/lastbackend/toolkit/tree/master/protoc-gen-toolkit/toolkit/options/annotations.proto)

## Getting Started

Start with define your apis/<service name>.proto file:

```proto
syntax = "proto3";

package lastbackend.example;

option go_package = "github.com/lastbackend/toolkit/examples/service/gen;servicepb";
import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";

message HelloWorldRequest {
  string name = 1;
}

message HelloWorldResponse {
  string id = 1;
}

service Example {
  //  Example methods
  rpc HelloWorld(HelloWorldRequest) returns (HelloWorldResponse) {}
};
```

Generate toolkit service with command

```bash
protoc \
    -I. \
    -I$GOPATH/src \
    -I$GOPATH/grpc/annotations \
    -I. \
    --go_out=:$GOPATH/src \
    --go-grpc_out=require_unimplemented_servers=false:$GOPATH/src \
    ./apis/<service name>.proto
```

Import generated serive in main.go

```go
import servicepb "<app path>/gen"
```

Define your application

```go
  app, err := servicepb.NewExampleService("example")
  if err != nil {
    fmt.Println(err)
  }
```

And now you can start it:

```go
  if err := app.Start(context.Background()); err != nil {
    app.Log().Errorf("could not run the service %v", err)
    os.Exit(1)
    return
  }

  // <- code goes here after application stops
```

Start application with:

```bash
go run main.go
```

Get help

```bash
go run main.go -h
```

Now you can add options/servers/plugins and packages. After modifing .proto file you need to re-run code generation command from step above.

Add some servers:

only GRPC
```proto
  option (toolkit.runtime) = {
    servers: [GRPC]
  };
```

only HTTP
```proto
  option (toolkit.runtime) = {
    servers: [HTTP]
  };
```

only GRPC, HTTP
```proto
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]
  };
```

Add plugin:

```proto
option (toolkit.plugins) = {
  prefix: "pgsql"    // prefix used if you need to add multiple same plugin instances
  plugin: "postgres" // check available plugins in plugins directory
};
```

Add anoter service client:

```proto
option (toolkit.services) = {
  service: "example",
  package: "<package path>"
};
```

Register your config

```go
  app, err := servicepb.NewExampleService("example")
  if err != nil {
    fmt.Println(err)
  }

  // Config management
  cfg := config.New()

  if err := app.RegisterConfig(cfg); err != nil {
    app.Log().Error(err)
    return
  }
```

Register custom package

```go
  app, err := servicepb.NewExampleService("example")
  if err != nil {
    fmt.Println(err)
  }

  // Config management
  cfg := config.New()

  if err := app.RegisterConfig(cfg); err != nil {
    app.Log().Error(err)
    return
  }

  // Add packages
  app.RegisterPackage(repository.NewRepository, controller.NewController)
```

Define GRPC service descriptor

```go
// exmaple services
type Handlers struct {
  servicepb.ExampleRpcServer

  app  toolkit.Service
  cfg  *config.Config
  repo *repository.Repository
}

func (h Handlers) HelloWorld(ctx context.Context, req *typespb.HelloWorldRequest) (*typespb.HelloWorldResponse, error) {
  h.app.Log().Info("ExamplseRpcServer: HelloWorld: call")

  md, ok := metadata.FromIncomingContext(ctx)
  if !ok {
    return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
  }

  demo := h.repo.Get(ctx)

  resp := typespb.HelloWorldResponse{
    Id:   fmt.Sprintf("%d", demo.Id),
    Name: fmt.Sprintf("%s: %d", req.Name, demo.Count),
    Type: req.Type,
  }

  if len(md["x-req-id"]) > 0 {
    header := metadata.New(map[string]string{"x-response-id": md["x-req-id"][0]})
    grpc.SendHeader(ctx, header)
  }

  return &resp, nil
}

func NewServer(app toolkit.Service, cfg *config.Config, repo *repository.Repository) servicepb.ExampleRpcServer {
  return &Handlers{
    repo: repo,
    app:  app,
    cfg:  cfg,
  }
}
```

```go
// main.go
  app, err := servicepb.NewExampleService("example")
  if err != nil {
    fmt.Println(err)
  }

  // Config management
  cfg := config.New()

  if err := app.RegisterConfig(cfg); err != nil {
    app.Log().Error(err)
    return
  }

  // Add packages
  app.RegisterPackage(repository.NewRepository, controller.NewController)

  // GRPC service descriptor
  app.Server().GRPC().SetService(server.NewServer)
```

## More examples

See here [See examples](https://github.com/lastbackend/toolkit/tree/master/examples) for more examples.

## Changelog

See [CHANGELOG](https://github.com/lastbackend/toolkit/tree/master/CHANGELOG.md) for release history.

## Contributing

Want to help develop? Check out our [contributing documentation](https://github.com/lastbackend/toolkit/tree/master/CONTRIBUTING.md).

## License

Last.Backend Toolkit is Apache 2.0 licensed.
