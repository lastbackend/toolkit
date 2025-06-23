# LastBackend Toolkit - Developer Guide

[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/lastbackend/toolkit?tab=doc)

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Getting Started](#getting-started)
4. [Code Generation](#code-generation)
5. [Service Configuration](#service-configuration)
6. [Plugin System](#plugin-system)
7. [Multi-Server Architecture](#multi-server-architecture)
8. [Client Generation](#client-generation)
9. [Testing Strategy](#testing-strategy)
10. [Best Practices](#best-practices)
11. [Examples Reference](#examples-reference)

## Overview

LastBackend Toolkit is a comprehensive Go framework for building microservices with automatic code generation, multi-protocol support (gRPC, HTTP, WebSocket), and a rich plugin system. The toolkit follows a protobuf-first approach where all services are defined in `.proto` files and the framework generates boilerplate code, client libraries, and infrastructure.

### Key Features

- **Protobuf-First Development** - All services defined in `.proto` files
- **Multi-Protocol Support** - gRPC, HTTP, WebSocket, and WebSocket Proxy servers
- **Automatic Code Generation** - Service interfaces, clients, and infrastructure code
- **Rich Plugin System** - Database connections, caching, message queues
- **Dependency Injection** - Constructor-based dependency injection
- **Configuration Management** - Environment-based configuration with validation
- **Testing Support** - Automatic mock generation and testing utilities
- **Middleware System** - Composable middleware for cross-cutting concerns

## Architecture

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Proto Files   │───▶│  Code Generator │───▶│  Generated Code │
│   (.proto)      │    │ (protoc-gen-    │    │  (.pb.go files) │
│                 │    │  toolkit)       │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                                              │
         ▼                                              ▼
┌─────────────────┐                            ┌─────────────────┐
│  Toolkit        │                            │  Application    │
│  Annotations    │                            │  Runtime        │
│                 │                            │                 │
└─────────────────┘                            └─────────────────┘
```

### Toolkit Annotations

The toolkit uses custom protobuf annotations to configure service behavior:

- **`toolkit.runtime`** - Configure server types and plugins
- **`toolkit.server`** - Define global middleware
- **`toolkit.route`** - Configure route-specific options
- **`toolkit.plugins`** - Plugin configuration
- **`toolkit.services`** - Client generation
- **`toolkit.tests_spec`** - Mock generation configuration

## Getting Started

### Prerequisites

1. **Go 1.21+**
2. **Protocol Buffers compiler (protoc)**
3. **Required protoc plugins:**
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   go install github.com/envoyproxy/protoc-gen-validate@latest
   go install github.com/lastbackend/toolkit/protoc-gen-toolkit@latest
   ```

### Project Initialization

#### Using the CLI Tool

```bash
# Install the toolkit CLI
go install github.com/lastbackend/toolkit/cli@latest

# Create a new service
toolkit init service github.com/yourorg/myservice --redis --postgres-gorm

# Navigate to the created directory
cd myservice

# Initialize and generate code
make init proto update tidy
```

#### Manual Setup

1. **Create project structure:**
   ```
   myservice/
   ├── apis/
   │   └── myservice.proto
   ├── config/
   │   └── config.go
   ├── internal/
   │   ├── server/
   │   ├── controller/
   │   └── repository/
   ├── scripts/
   │   └── generate.sh
   ├── main.go
   ├── generate.go
   └── go.mod
   ```

2. **Define your service in `apis/myservice.proto`:**
   ```protobuf
   syntax = "proto3";
   package myservice;
   
   option go_package = "github.com/yourorg/myservice/gen;servicepb";
   
   import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";
   
   // Configure plugins
   option (toolkit.plugins) = {
     prefix: "redis"
     plugin: "redis"
   };
   
   option (toolkit.plugins) = {
     prefix: "pgsql"
     plugin: "postgres_gorm"
   };
   
   service MyService {
     option (toolkit.runtime) = {
       servers: [GRPC, HTTP]
     };
     option (toolkit.server) = {
       middlewares: ["request_id"]
     };
     
     rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
     rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {}
   }
   
   message GetUserRequest {
     string user_id = 1;
   }
   
   message GetUserResponse {
     string user_id = 1;
     string name = 2;
     string email = 3;
   }
   
   message CreateUserRequest {
     string name = 1;
     string email = 2;
   }
   
   message CreateUserResponse {
     string user_id = 1;
   }
   ```

3. **Create main.go:**
   ```go
   package main
   
   import (
       "context"
       "fmt"
       "os"
       
       "github.com/yourorg/myservice/config"
       servicepb "github.com/yourorg/myservice/gen"
       "github.com/yourorg/myservice/internal/controller"
       "github.com/yourorg/myservice/internal/repository"
       "github.com/yourorg/myservice/internal/server"
       "github.com/lastbackend/toolkit/pkg/runtime"
   )
   
   func main() {
       app, err := servicepb.NewMyServiceService("myservice",
           runtime.WithVersion("1.0.0"),
           runtime.WithDescription("My Service"),
           runtime.WithEnvPrefix("MYSERVICE"),
       )
       if err != nil {
           fmt.Println(err)
           os.Exit(1)
       }
   
       // Register configuration
       cfg := config.New()
       if err := app.RegisterConfig(cfg); err != nil {
           app.Log().Error(err)
           return
       }
   
       // Register packages (dependency injection)
       app.RegisterPackage(repository.NewRepository, controller.NewController)
   
       // Configure servers
       app.Server().GRPC().SetService(server.NewServer)
   
       // Start the service
       if err := app.Start(context.Background()); err != nil {
           app.Log().Errorf("could not run the service %v", err)
           os.Exit(1)
       }
   }
   ```

## Code Generation

### Generation Process

The toolkit uses a multi-step code generation process:

1. **Protobuf compilation** - Standard `.pb.go` files
2. **gRPC compilation** - gRPC service definitions
3. **Validation compilation** - Input validation code
4. **Toolkit compilation** - Service bootstrap and infrastructure

### Generation Script

Create `scripts/generate.sh`:

```bash
#!/bin/bash -e

SOURCE_PACKAGE=github.com/yourorg/myservice
ROOT_DIR=$GOPATH/src/$SOURCE_PACKAGE
PROTO_DIR=$ROOT_DIR/apis

# Clean previous generation
find $ROOT_DIR -type f \( -name '*.pb.go' -o -name '*.pb.*.go' \) -delete

# Setup dependencies
mkdir -p $PROTO_DIR/google/api
mkdir -p $PROTO_DIR/validate

# Download required proto files
curl -s -f -o $PROTO_DIR/google/api/annotations.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
curl -s -f -o $PROTO_DIR/google/api/http.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto
curl -s -f -o $PROTO_DIR/validate/validate.proto -L https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/main/validate/validate.proto

# Find all proto files
PROTOS=$(find $PROTO_DIR -type f -name '*.proto' | grep -v $PROTO_DIR/google/api | grep -v $PROTO_DIR/validate)

# Generate code
for PROTO in $PROTOS; do
  protoc \
    -I. \
    -I$GOPATH/src \
    -I$PROTO_DIR \
    -I$(dirname $PROTO) \
    --validate_out=lang=go:$GOPATH/src \
    --go_out=:$GOPATH/src \
    --go-grpc_out=require_unimplemented_servers=false:$GOPATH/src \
    --toolkit_out=$GOPATH/src \
    $PROTO
done

# Cleanup
rm -r $PROTO_DIR/google
rm -r $PROTO_DIR/validate

echo "Generation completed successfully"
```

### Generated Code Structure

After generation, your project will have:

```
gen/
├── myservice.pb.go                    # Protobuf messages
├── myservice_grpc.pb.go              # gRPC service definitions
├── myservice.pb.validate.go          # Validation code
├── myservice_service.pb.toolkit.go   # Service bootstrap (main file)
├── client/                           # Generated clients
│   └── myservice.pb.toolkit.rpc.go
└── tests/                           # Generated mocks
    └── myservice.pb.toolkit.mockery.go
```

## Service Configuration

### Toolkit Runtime Options

Configure your service behavior using protobuf annotations:

```protobuf
service MyService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP, WEBSOCKET]  // Server types
    plugins: [{                       // Service-level plugins
      prefix: "cache"
      plugin: "redis"
    }]
  };
  
  option (toolkit.server) = {
    middlewares: [                    // Global middleware
      "request_id",
      "auth",
      "rate_limit"
    ]
  };
  
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (toolkit.route) = {
      middlewares: ["validate"]       // Route-specific middleware
      exclude_global_middlewares: ["auth"]  // Exclude global middleware
    };
    option (google.api.http) = {
      get: "/users/{user_id}"         // HTTP route
    };
  };
}
```

### Server Types

| Server Type | Description | Use Case |
|-------------|-------------|----------|
| `GRPC` | Standard gRPC server | High-performance RPC calls |
| `HTTP` | HTTP REST server | Web APIs, external integrations |
| `WEBSOCKET` | WebSocket server | Real-time communication |
| `WEBSOCKET_PROXY` | WebSocket proxy to gRPC | WebSocket frontend to gRPC backend |

### Configuration Management

**Important:** Most configuration in the toolkit comes from plugins. Plugin configurations are automatically parsed using environment variables with prefixes based on the plugin names declared in your proto files.

#### Plugin-Driven Configuration

When you declare plugins in proto files:
```protobuf
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

option (toolkit.plugins) = {
  prefix: "redis"
  plugin: "redis"
};
```

The toolkit automatically creates environment variable configuration with prefixes:
```bash
# Service prefix: MYSERVICE (from runtime.WithEnvPrefix)
# Plugin configurations are parsed automatically
MYSERVICE_PGSQL_HOST=localhost
MYSERVICE_PGSQL_PORT=5432
MYSERVICE_PGSQL_USERNAME=user
MYSERVICE_PGSQL_PASSWORD=password
MYSERVICE_PGSQL_DATABASE=myapp

MYSERVICE_REDIS_HOST=localhost
MYSERVICE_REDIS_PORT=6379
MYSERVICE_REDIS_PASSWORD=redispass
```

#### Application-Specific Configuration

For application-specific configuration, create a simple config struct:

```go
// config/config.go
package config

type Config struct {
    // Application-specific settings
    AppName     string `env:"APP_NAME" envDefault:"MyService"`
    Environment string `env:"ENVIRONMENT" envDefault:"development"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
    
    // Business logic configuration
    MaxRetries      int           `env:"MAX_RETRIES" envDefault:"3"`
    RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" envDefault:"30s"`
    EnableFeatureX  bool          `env:"ENABLE_FEATURE_X" envDefault:"false"`
}

func New() *Config {
    return &Config{}
    // No need to parse here - toolkit handles this automatically
}
```

#### Environment Variable Hierarchy

The toolkit builds environment variables with this pattern:
```
{SERVICE_PREFIX}_{PLUGIN_PREFIX}_{SETTING_NAME}
```

Example with service prefix "MYSERVICE":
- Plugin configs: `MYSERVICE_PGSQL_HOST`, `MYSERVICE_REDIS_PORT`
- App configs: `MYSERVICE_APP_NAME`, `MYSERVICE_LOG_LEVEL`
- Server configs: `MYSERVICE_SERVER_GRPC_PORT`

## Plugin System

### Plugin Architecture Overview

The toolkit's plugin system provides modular integration for external services with automatic dependency injection, lifecycle management, and configuration parsing.

**Key Features:**
- **Declarative Configuration** - Define plugins in proto files
- **Automatic Code Generation** - Generated interfaces and registration
- **Dependency Injection** - Built on Uber FX framework
- **Lifecycle Management** - PreStart, OnStart, OnStop hooks
- **Multi-Instance Support** - Multiple instances of same plugin type

### Available Plugins

From `github.com/lastbackend/toolkit-plugins`:

| Plugin | Package | Purpose |
|--------|---------|----------|
| **postgres_gorm** | `postgres_gorm` | PostgreSQL with GORM ORM |
| **postgres_pg** | `postgres_pg` | PostgreSQL with go-pg |
| **postgres_pgx** | `postgres_pgx` | PostgreSQL with pgx driver |
| **redis** | `redis` | Redis caching and pub/sub |
| **rabbitmq** | `rabbitmq` | RabbitMQ message queue |
| **centrifuge** | `centrifuge` | Real-time messaging |
| **sentry** | `sentry` | Error monitoring |
| **resolver_consul** | `resolver_consul` | Consul service discovery |

### Plugin Declaration in Proto Files

#### Global Plugins (File-level)
```protobuf
// Available to all services in this file
option (toolkit.plugins) = {
  prefix: "pgsql"           // Environment variable prefix
  plugin: "postgres_gorm"  // Plugin type from toolkit-plugins
};

option (toolkit.plugins) = {
  prefix: "cache"
  plugin: "redis"
};

option (toolkit.plugins) = {
  prefix: "queue"
  plugin: "rabbitmq"
};
```

#### Service-Specific Plugins
```protobuf
service MyService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]
    plugins: [{
      prefix: "session_store"
      plugin: "redis"
    }]
  };
}
```

#### Multi-Instance Plugin Pattern
```protobuf
// Multiple instances of same plugin type
option (toolkit.plugins) = {
  prefix: "primary_db"
  plugin: "postgres_gorm"
};

option (toolkit.plugins) = {
  prefix: "analytics_db"
  plugin: "postgres_gorm"
};

option (toolkit.plugins) = {
  prefix: "user_cache"
  plugin: "redis"
};

option (toolkit.plugins) = {
  prefix: "session_cache"
  plugin: "redis"
};
```

### Generated Code Integration

The toolkit generates plugin interfaces and registration:

```go
// Generated plugin interfaces
type PgsqlPlugin interface {
    postgres_gorm.Plugin
}

type CachePlugin interface {
    redis.Plugin
}

type QueuePlugin interface {
    rabbitmq.Plugin
}

// Generated service constructor with plugin initialization
func NewMyServiceService(name string, opts ...runtime.Option) (_ toolkit.Service, err error) {
    app := new(serviceMyService)
    
    app.runtime, err = controller.NewRuntime(context.Background(), name, opts...)
    if err != nil {
        return nil, err
    }

    // Plugin instances created with automatic configuration parsing
    plugin_pgsql := postgres_gorm.NewPlugin(app.runtime, &postgres_gorm.Options{Name: "pgsql"})
    plugin_cache := redis.NewPlugin(app.runtime, &redis.Options{Name: "cache"})
    plugin_queue := rabbitmq.NewPlugin(app.runtime, &rabbitmq.Options{Name: "queue"})

    // Plugin registration for dependency injection
    app.runtime.Plugin().Provide(func() PgsqlPlugin { return plugin_pgsql })
    app.runtime.Plugin().Provide(func() CachePlugin { return plugin_cache })
    app.runtime.Plugin().Provide(func() QueuePlugin { return plugin_queue })

    return app.runtime.Service(), nil
}
```

### Plugin Usage in Components

Plugins are automatically injected into components that declare them:

```go
// internal/repository/repository.go
type Repository struct {
    db  servicepb.PgsqlPlugin  // Automatically injected
    log toolkit.Logger
}

// Constructor with plugin dependency
func NewRepository(app toolkit.Service, db servicepb.PgsqlPlugin) *Repository {
    return &Repository{
        db:  db,
        log: app.Log(),
    }
}

// Use plugin interface methods
func (r *Repository) GetUser(ctx context.Context, userID string) (*User, error) {
    var user User
    err := r.db.DB().WithContext(ctx).Where("id = ?", userID).First(&user).Error
    return &user, err
}

// internal/service/service.go
type Service struct {
    repo  *repository.Repository
    cache servicepb.CachePlugin
    queue servicepb.QueuePlugin
}

func NewService(
    repo *repository.Repository,
    cache servicepb.CachePlugin,
    queue servicepb.QueuePlugin,
) *Service {
    return &Service{
        repo:  repo,
        cache: cache,
        queue: queue,
    }
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    user := &User{Name: req.Name, Email: req.Email}
    
    // Save to database
    if err := s.repo.CreateUser(ctx, user); err != nil {
        return err
    }
    
    // Cache user data
    userData, _ := json.Marshal(user)
    s.cache.Client().Set(ctx, fmt.Sprintf("user:%s", user.ID), userData, time.Hour)
    
    // Publish event
    eventData, _ := json.Marshal(map[string]interface{}{
        "type": "user_created",
        "user_id": user.ID,
    })
    s.queue.Publish("user.created", eventData)
    
    return nil
}
```

### Plugin Configuration

Plugin configurations are automatically parsed based on prefixes:

```bash
# PostgreSQL plugin (prefix: "pgsql")
MYSERVICE_PGSQL_HOST=localhost
MYSERVICE_PGSQL_PORT=5432
MYSERVICE_PGSQL_USERNAME=user
MYSERVICE_PGSQL_PASSWORD=secret
MYSERVICE_PGSQL_DATABASE=myapp
MYSERVICE_PGSQL_SSL_MODE=disable
MYSERVICE_PGSQL_TIMEZONE=UTC
MYSERVICE_PGSQL_DEBUG=false

# Redis plugin (prefix: "cache")
MYSERVICE_CACHE_HOST=localhost
MYSERVICE_CACHE_PORT=6379
MYSERVICE_CACHE_PASSWORD=redispass
MYSERVICE_CACHE_DATABASE=0
MYSERVICE_CACHE_POOL_SIZE=10

# RabbitMQ plugin (prefix: "queue")
MYSERVICE_QUEUE_HOST=localhost
MYSERVICE_QUEUE_PORT=5672
MYSERVICE_QUEUE_USERNAME=guest
MYSERVICE_QUEUE_PASSWORD=guest
MYSERVICE_QUEUE_VHOST=/
```

### Plugin Lifecycle

Plugins implement optional lifecycle hooks:

```go
// Plugin lifecycle methods (called automatically)
type Plugin interface {
    // PreStart: Synchronous initialization (connections, etc.)
    PreStart(ctx context.Context) error
    
    // OnStart: Asynchronous startup (background workers)
    OnStart(ctx context.Context) error
    
    // OnStop: Graceful shutdown
    OnStop(ctx context.Context) error
}
```

Lifecycle execution order:
1. **PreStart** - All plugins initialized synchronously
2. **OnStart** - Background processes started asynchronously
3. **Service runs**
4. **OnStop** - Graceful shutdown when service stops

### Custom Plugin Development

Create custom plugins by implementing the plugin interface:

```go
package myplugin

import (
    "context"
    "github.com/lastbackend/toolkit/pkg/runtime"
)

type Plugin interface {
    Init(ctx context.Context) error
    Name() string
    Client() MyClient  // Your custom client
}

type plugin struct {
    runtime runtime.Runtime
    options *Options
    client  MyClient
}

type Options struct {
    Name     string
    Host     string
    Port     int
}

func NewPlugin(runtime runtime.Runtime, opts *Options) Plugin {
    return &plugin{
        runtime: runtime,
        options: opts,
    }
}

func (p *plugin) Init(ctx context.Context) error {
    // Initialize your plugin
    client, err := NewMyClient(p.options.Host, p.options.Port)
    if err != nil {
        return err
    }
    p.client = client
    return nil
}

func (p *plugin) Name() string {
    return p.options.Name
}

func (p *plugin) Client() MyClient {
    return p.client
}
```

## Multi-Server Architecture

### Concurrent Server Operation

Services can run multiple server types simultaneously:

```protobuf
service MyService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP, WEBSOCKET]
  };
  
  // This method is available on all server types
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/users/{user_id}"
    };
  };
  
  // WebSocket-specific method
  rpc Subscribe(SubscribeRequest) returns (SubscribeResponse) {
    option (toolkit.route).websocket = true;
    option (google.api.http) = {
      get: "/subscribe"
    };
  };
}
```

### Server Configuration

Configure each server type:

```go
func main() {
    app, err := servicepb.NewMyServiceService("myservice")
    if err != nil {
        panic(err)
    }
    
    // gRPC server configuration
    app.Server().GRPC().SetService(server.NewGRPCServer)
    app.Server().GRPC().SetInterceptor(server.NewGRPCInterceptor)
    
    // HTTP server configuration  
    app.Server().HTTP().SetMiddleware(server.NewHTTPMiddleware)
    app.Server().HTTP().AddHandler(http.MethodGet, "/health", server.HealthCheck)
    
    // WebSocket configuration
    app.Server().HTTP().Subscribe("user:update", server.HandleUserUpdate)
    
    if err := app.Start(context.Background()); err != nil {
        panic(err)
    }
}
```

### Proxy Configuration

Configure HTTP-to-gRPC and WebSocket-to-gRPC proxying:

```protobuf
service Gateway {
  option (toolkit.runtime) = {
    servers: [HTTP, WEBSOCKET_PROXY]
  };
  
  // HTTP proxy to external gRPC service
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (toolkit.route) = {
      http_proxy: {
        service: "user-service"
        method: "/userservice.UserService/GetUser"
      }
    };
    option (google.api.http) = {
      get: "/users/{user_id}"
    };
  };
  
  // WebSocket proxy to gRPC
  rpc StreamUsers(StreamUsersRequest) returns (StreamUsersResponse) {
    option (toolkit.route).websocket_proxy = {
      service: "user-service"
      method: "/userservice.UserService/StreamUsers"
    };
  };
}
```

## Client Generation 

### Service Client Generation

Generate clients for external services:

```protobuf
// Configure client generation
option (toolkit.services) = {
  service: "user-service",
  package: "github.com/yourorg/myservice/gen/client"
};

option (toolkit.services) = {
  service: "notification-service",
  package: "github.com/yourorg/myservice/gen/client"
};
```

### Using Generated Clients

Clients are automatically injected and available through the services interface:

```go
// Generated services interface
type MyServiceServices interface {
    UserService() userservice.UserServiceRPCClient
    NotificationService() notificationservice.NotificationServiceRPCClient
}

// Usage in your controller
func NewController(
    app toolkit.Service,
    services servicepb.MyServiceServices,
) *Controller {
    return &Controller{
        app:      app,
        services: services,
    }
}

func (c *Controller) GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
    // Call external user service
    user, err := c.services.UserService().GetUser(ctx, &userservice.GetUserRequest{
        UserId: userID,
    })
    if err != nil {
        return nil, err
    }
    
    // Call notification service
    _, err = c.services.NotificationService().SendNotification(ctx, &notificationservice.SendNotificationRequest{
        UserId:  userID,
        Message: "Profile accessed",
    })
    if err != nil {
        c.app.Log().Warnf("Failed to send notification: %v", err)
    }
    
    return &UserProfile{
        ID:    user.UserId,
        Name:  user.Name,
        Email: user.Email,
    }, nil
}
```

### Client Configuration

Configure client behavior:

```go
// Configure gRPC client options
runtime.WithGRPCClientOptions(
    runtime.GRPCClientDialTimeout(5*time.Second),
    runtime.GRPCClientKeepAlive(30*time.Second),
    runtime.GRPCClientMaxRecvMsgSize(4*1024*1024),
)
```

## Testing Strategy

### Mock Generation

Configure automatic mock generation:

```protobuf
option (toolkit.tests_spec) = {
  mockery: {
    package: "github.com/yourorg/myservice/gen/tests"
  }
};
```

### Generated Mocks

The toolkit generates mocks for all service interfaces:

```go
// gen/tests/myservice.pb.toolkit.mockery.go
type MockMyServiceRpcServer struct {
    mock.Mock
}

func (m *MockMyServiceRpcServer) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*GetUserResponse), args.Error(1)
}
```

### Unit Testing

Example unit test using generated mocks:

```go
// internal/server/server_test.go
package server

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    servicepb "github.com/yourorg/myservice/gen"
    "github.com/yourorg/myservice/gen/tests"
)

func TestGetUser(t *testing.T) {
    // Setup
    mockRepo := new(tests.MockRepository)
    mockServices := new(tests.MockMyServiceServices)
    
    server := NewServer(nil, nil, mockRepo, mockServices)
    
    // Mock expectations
    expectedUser := &User{
        ID:    "123",
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    mockRepo.On("GetUser", mock.Anything, "123").Return(expectedUser, nil)
    
    // Execute
    resp, err := server.GetUser(context.Background(), &servicepb.GetUserRequest{
        UserId: "123",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "123", resp.UserId)
    assert.Equal(t, "John Doe", resp.Name)
    assert.Equal(t, "john@example.com", resp.Email)
    
    mockRepo.AssertExpectations(t)
}
```

### Integration Testing

Example integration test:

```go
// integration_test.go
package main

import (
    "context"
    "testing"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/test/bufconn"
    
    servicepb "github.com/yourorg/myservice/gen"
)

func TestMyServiceIntegration(t *testing.T) {
    // Setup test server
    lis := bufconn.Listen(1024 * 1024)
    s := grpc.NewServer()
    
    // Register your service
    servicepb.RegisterMyServiceServer(s, NewTestServer())
    
    go func() {
        if err := s.Serve(lis); err != nil {
            t.Errorf("Server exited with error: %v", err)
        }
    }()
    
    defer s.Stop()
    
    // Create client
    conn, err := grpc.DialContext(context.Background(), "bufnet",
        grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
            return lis.Dial()
        }),
        grpc.WithInsecure(),
    )
    if err != nil {
        t.Fatalf("Failed to dial bufnet: %v", err)
    }
    defer conn.Close()
    
    client := servicepb.NewMyServiceClient(conn)
    
    // Test your service
    resp, err := client.GetUser(context.Background(), &servicepb.GetUserRequest{
        UserId: "test-user",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, resp)
}
```

## Best Practices

### 1. Project Structure

Follow the recommended project structure:

```
myservice/
├── apis/                    # Protobuf definitions
│   ├── myservice.proto
│   └── ptypes/             # Shared message types
├── config/                 # Configuration management
│   └── config.go
├── gen/                    # Generated code (auto-generated)
├── internal/               # Internal packages
│   ├── controller/         # Business logic
│   ├── repository/         # Data access layer
│   └── server/            # Server implementations
├── scripts/               # Build and generation scripts
│   ├── bootstrap.sh
│   └── generate.sh
├── tests/                 # Test files
├── main.go               # Application entry point
├── generate.go           # Go generate directive
├── go.mod
└── Makefile
```

### 2. Error Handling

Implement structured error handling:

```go
// internal/errors/errors.go
package errors

import (
    "fmt"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

var (
    ErrUserNotFound     = status.Error(codes.NotFound, "user not found")
    ErrInvalidInput     = status.Error(codes.InvalidArgument, "invalid input")
    ErrInternalError    = status.Error(codes.Internal, "internal server error")
)

func UserNotFound(userID string) error {
    return status.Errorf(codes.NotFound, "user %s not found", userID)
}

func ValidationError(field string, message string) error {
    return status.Errorf(codes.InvalidArgument, "validation failed for %s: %s", field, message)
}
```

### 3. Logging

Use structured logging throughout your application:

```go
func (s *Server) GetUser(ctx context.Context, req *servicepb.GetUserRequest) (*servicepb.GetUserResponse, error) {
    logger := s.app.Log().WithField("user_id", req.UserId)
    logger.Info("GetUser request received")
    
    user, err := s.repo.GetUser(ctx, req.UserId)
    if err != nil {
        logger.WithError(err).Error("Failed to get user from repository")
        return nil, status.Error(codes.Internal, "failed to get user")
    }
    
    logger.Info("GetUser request completed successfully")
    return &servicepb.GetUserResponse{
        UserId: user.ID,
        Name:   user.Name,
        Email:  user.Email,
    }, nil
}
```

### 4. Configuration Validation

Validate configuration at startup:

```go
func (c *Config) Validate() error {
    if c.Database.Username == "" {
        return fmt.Errorf("database username is required")
    }
    if c.Database.Password == "" {
        return fmt.Errorf("database password is required")
    }
    if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
        return fmt.Errorf("invalid gRPC port: %d", c.Server.GRPCPort)
    }
    return nil
}

func main() {
    cfg := config.New()
    if err := cfg.Validate(); err != nil {
        log.Fatalf("Configuration validation failed: %v", err)
    }
    // ... rest of initialization
}
```

### 5. Graceful Shutdown

Implement graceful shutdown:

```go
func main() {
    app, err := servicepb.NewMyServiceService("myservice")
    if err != nil {
        log.Fatal(err)
    }
    
    // Setup graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        app.Log().Info("Shutdown signal received")
        cancel()
        app.Stop(ctx)
    }()
    
    // Start the service
    if err := app.Start(ctx); err != nil {
        app.Log().Errorf("Service failed: %v", err)
        os.Exit(1)
    }
    
    app.Log().Info("Service shutdown completed")
}
```

### 6. Health Checks

Implement health checks:

```go
// internal/health/health.go
package health

import (
    "context"
    "github.com/lastbackend/toolkit"
)

type HealthChecker struct {
    app      toolkit.Service
    database DatabaseChecker
    redis    RedisChecker
}

func NewHealthChecker(app toolkit.Service, db DatabaseChecker, redis RedisChecker) *HealthChecker {
    return &HealthChecker{
        app:      app,
        database: db,
        redis:    redis,
    }
}

func (h *HealthChecker) Check(ctx context.Context) error {
    // Check database connectivity
    if err := h.database.Ping(ctx); err != nil {
        return fmt.Errorf("database health check failed: %w", err)
    }
    
    // Check Redis connectivity
    if err := h.redis.Ping(ctx); err != nil {
        return fmt.Errorf("redis health check failed: %w", err)
    }
    
    return nil
}
```

### 7. Middleware Development

Create reusable middleware:

```go
// internal/middleware/auth.go
package middleware

import (
    "context"
    "strings"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // Skip auth for health checks
    if strings.HasSuffix(info.FullMethod, "/Health") {
        return handler(ctx, req)
    }
    
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "metadata not found")
    }
    
    tokens := md.Get("authorization")
    if len(tokens) == 0 {
        return nil, status.Error(codes.Unauthenticated, "authorization token required")
    }
    
    token := strings.TrimPrefix(tokens[0], "Bearer ")
    if !isValidToken(token) {
        return nil, status.Error(codes.Unauthenticated, "invalid token")
    }
    
    // Add user info to context
    userID := getUserIDFromToken(token)
    ctx = context.WithValue(ctx, "user_id", userID)
    
    return handler(ctx, req)
}

func isValidToken(token string) bool {
    // Implement your token validation logic
    return true
}

func getUserIDFromToken(token string) string {
    // Extract user ID from token
    return "user-123"
}
```

## Examples Reference

### Basic Service Example

**File: `examples/helloworld/`**
- Simple gRPC service without toolkit annotations
- Manual server setup
- Basic request/response pattern

### Full Microservice Example

**File: `examples/service/`**
- Complete microservice with plugins
- Multi-server setup (gRPC + HTTP)
- Dependency injection
- Testing with mocks
- Configuration management

### Gateway Service Example

**File: `examples/gateway/`**
- HTTP-to-gRPC proxy
- Service discovery
- REST API endpoints

### HTTP-Only Service Example

**File: `examples/http/`**
- HTTP-only server
- JSON marshaling
- REST endpoints with Google API annotations

### WebSocket Service Example

**File: `examples/wss/`**
- WebSocket server
- WebSocket proxy to gRPC
- Real-time communication
- Swagger documentation

---

## Next Steps

1. **Explore Examples** - Study the provided examples to understand different patterns
2. **Build Your First Service** - Create a simple service using the patterns shown
3. **Add Plugins** - Integrate database and caching plugins
4. **Implement Testing** - Add unit and integration tests
5. **Deploy** - Use Docker and Kubernetes for deployment

For more detailed information, refer to the specific example implementations in the `examples/` directory.