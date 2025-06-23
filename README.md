# LastBackend Toolkit

[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/lastbackend/toolkit?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/lastbackend/toolkit)](https://goreportcard.com/report/github.com/lastbackend/toolkit)
[![Discord](https://dcbadge.vercel.app/api/server/WhK9ujvem9)](https://discord.gg/WhK9ujvem9)

**A comprehensive Go framework for building production-ready microservices with automatic code generation, multi-protocol support, and modular plugin architecture.**

LastBackend Toolkit eliminates microservice boilerplate through protobuf-first development, enabling developers to focus on business logic while the framework handles infrastructure concerns.

---

## ✨ Key Features

- 🔧 **Protobuf-First Development** - Define services in `.proto` files, get everything else generated
- 🌐 **Multi-Protocol Support** - gRPC, HTTP REST, WebSocket from single definition  
- 🔌 **Rich Plugin Ecosystem** - PostgreSQL, Redis, RabbitMQ, and [more plugins](https://github.com/lastbackend/toolkit-plugins)
- 🏗️ **Automatic Code Generation** - Service interfaces, clients, mocks, and infrastructure
- 💉 **Dependency Injection** - Built on Uber FX with automatic plugin registration
- ⚙️ **Environment Configuration** - Hierarchical configuration with validation
- 🧪 **Testing Ready** - Generated mocks and testing utilities
- 📊 **Production Features** - Metrics, tracing, health checks, graceful shutdown

## 🚀 Quick Start

### Installation

```bash
# Install the toolkit CLI
go install github.com/lastbackend/toolkit/cli@latest

# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/envoyproxy/protoc-gen-validate@latest
go install github.com/lastbackend/toolkit/protoc-gen-toolkit@latest
```

### Create Your First Service

```bash
# Generate a new microservice with PostgreSQL and Redis
toolkit init service github.com/yourorg/user-service --postgres-gorm --redis

cd user-service

# Initialize and generate code
make init proto update tidy

# Run the service
go run main.go
```

### Define Your Service

**`apis/user-service.proto`:**
```protobuf
syntax = "proto3";
package userservice;

option go_package = "github.com/yourorg/user-service/gen;servicepb";

import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";
import "google/api/annotations.proto";
import "validate/validate.proto";

// Database plugin
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

// Cache plugin  
option (toolkit.plugins) = {
  prefix: "cache"
  plugin: "redis"
};

service UserService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]  // Multi-protocol support
  };
  
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/users"
      body: "*"
    };
  };
  
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/users/{user_id}"
    };
  };
}

message CreateUserRequest {
  string name = 1 [(validate.rules).string.min_len = 1];
  string email = 2 [(validate.rules).string.pattern = "^[^@]+@[^@]+\\.[^@]+$"];
}

message CreateUserResponse {
  string user_id = 1;
  string status = 2;
}

message GetUserRequest {
  string user_id = 1 [(validate.rules).string.uuid = true];
}

message GetUserResponse {
  string user_id = 1;
  string name = 2;
  string email = 3;
  google.protobuf.Timestamp created_at = 4;
}
```

### Generated Code Usage

The toolkit generates everything you need:

**`main.go`:**
```go
package main

import (
    "context"
    servicepb "github.com/yourorg/user-service/gen"
    "github.com/yourorg/user-service/internal/server"
    "github.com/lastbackend/toolkit/pkg/runtime"
)

func main() {
    // Service with automatic plugin initialization
    app, err := servicepb.NewUserServiceService("user-service",
        runtime.WithVersion("1.0.0"),
        runtime.WithEnvPrefix("USER_SERVICE"),
    )
    if err != nil {
        panic(err)
    }

    // Register your business logic
    app.Server().GRPC().SetService(server.NewUserServer)

    // Start with gRPC and HTTP servers
    if err := app.Start(context.Background()); err != nil {
        panic(err)
    }
}
```

**`internal/server/server.go`:**
```go
package server

import (
    "context"
    servicepb "github.com/yourorg/user-service/gen"
)

type UserServer struct {
    servicepb.UserServiceRpcServer
    
    app    toolkit.Service
    pgsql  servicepb.PgsqlPlugin  // Auto-injected PostgreSQL
    cache  servicepb.CachePlugin  // Auto-injected Redis
}

// Constructor with automatic dependency injection
func NewUserServer(
    app toolkit.Service,
    pgsql servicepb.PgsqlPlugin,
    cache servicepb.CachePlugin,
) servicepb.UserServiceRpcServer {
    return &UserServer{app: app, pgsql: pgsql, cache: cache}
}

func (s *UserServer) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    // Use plugins directly - they're ready to go
    user := &User{Name: req.Name, Email: req.Email}
    
    // Save to PostgreSQL
    if err := s.pgsql.DB().WithContext(ctx).Create(user).Error; err != nil {
        return nil, err
    }
    
    // Cache the user
    s.cache.Client().Set(ctx, fmt.Sprintf("user:%s", user.ID), user, time.Hour)
    
    return &CreateUserResponse{
        UserId: user.ID,
        Status: "created",
    }, nil
}
```

### Configuration via Environment Variables

```bash
# Service configuration
USER_SERVICE_APP_NAME=user-service
USER_SERVICE_ENVIRONMENT=production

# PostgreSQL plugin configuration (auto-parsed)
USER_SERVICE_PGSQL_HOST=localhost
USER_SERVICE_PGSQL_PORT=5432
USER_SERVICE_PGSQL_USERNAME=user
USER_SERVICE_PGSQL_PASSWORD=secret
USER_SERVICE_PGSQL_DATABASE=users_db

# Redis plugin configuration (auto-parsed)
USER_SERVICE_CACHE_HOST=localhost
USER_SERVICE_CACHE_PORT=6379
USER_SERVICE_CACHE_PASSWORD=redispass
```

## 📚 Documentation

- **[Getting Started](docs/getting-started.md)** - Detailed setup and first steps
- **[Architecture Guide](docs/architecture.md)** - Framework concepts and design
- **[Plugin System](docs/plugins/)** - Available plugins and development guide
- **[Configuration](docs/configuration.md)** - Environment-based configuration
- **[Testing](docs/testing.md)** - Testing strategies and generated mocks
- **[Deployment](docs/deployment.md)** - Production deployment guide

### For AI Agents
- **[AI Instructions](.ai/INSTRUCTIONS.md)** - Complete guide for AI code generation
- **[Usage Patterns](.ai/toolkit-usage.md)** - Common patterns and examples
- **[Troubleshooting](.ai/troubleshooting.md)** - Common issues and solutions

## 🔌 Plugin Ecosystem

LastBackend Toolkit includes a rich set of plugins from [`toolkit-plugins`](https://github.com/lastbackend/toolkit-plugins):

| Plugin | Purpose | Interface |
|--------|---------|-----------|
| **postgres_gorm** | PostgreSQL with GORM | `Plugin.DB() *gorm.DB` |
| **postgres_pg** | PostgreSQL with go-pg | `Plugin.DB() *pg.DB` |
| **redis** | Redis cache/pub-sub | `Plugin.Client() redis.Cmdable` |
| **rabbitmq** | Message queue | `Plugin.Publish/Subscribe` |
| **centrifuge** | Real-time messaging | `Plugin.Node() *centrifuge.Node` |
| **sentry** | Error monitoring | Error tracking integration |

### Using Plugins

```protobuf
// Declare plugins in your .proto file
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

option (toolkit.plugins) = {
  prefix: "cache"  
  plugin: "redis"
};
```

Plugins are automatically:
- ✅ Initialized with environment configuration
- ✅ Registered for dependency injection
- ✅ Managed through lifecycle hooks
- ✅ Available as typed interfaces in your code

## 🏗️ Examples

Explore real-world examples in the [`examples/`](examples/) directory:

- **[Basic Service](examples/helloworld/)** - Simple gRPC service
- **[Full Microservice](examples/service/)** - Complete service with plugins
- **[HTTP Gateway](examples/gateway/)** - HTTP-to-gRPC proxy
- **[WebSocket Service](examples/wss/)** - Real-time WebSocket service

## 🔄 Multi-Protocol Support

Define once, serve everywhere:

```protobuf
service UserService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP, WEBSOCKET]  // All protocols from one definition
  };
  
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    // Available as gRPC call
    option (google.api.http) = {
      get: "/users/{user_id}"  // Also as HTTP GET
    };
  };
  
  rpc StreamUsers(StreamRequest) returns (StreamResponse) {
    option (toolkit.route).websocket = true;  // And WebSocket
  };
}
```

Your service automatically provides:
- 🔧 **gRPC** - High-performance RPC
- 🌐 **HTTP REST** - Web-friendly APIs  
- ⚡ **WebSocket** - Real-time communication
- 🔀 **Proxying** - HTTP-to-gRPC and WebSocket-to-gRPC

## 🧪 Testing Support

Generated testing utilities:

```go
// Generated mocks for all interfaces
mockPlugin := new(tests.MockPgsqlPlugin)
mockPlugin.On("DB").Return(mockDB)

// Generated client for integration tests
client := servicepb.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &GetUserRequest{UserId: "123"})
```

## 🏢 Production Ready

Built-in production features:
- 📊 **Metrics** - Prometheus integration
- 🔍 **Tracing** - Distributed tracing support  
- 🏥 **Health Checks** - Kubernetes-ready health endpoints
- 🛡️ **Graceful Shutdown** - Clean resource cleanup
- ⚙️ **Configuration Validation** - Environment variable validation
- 📝 **Structured Logging** - Contextual logging throughout

## 🤝 Contributing

We welcome contributions! Please see:
- [Contributing Guidelines](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Development Setup](docs/development.md)

### Development

```bash
# Clone the repository
git clone https://github.com/lastbackend/toolkit.git
cd toolkit

# Install dependencies
make init

# Run tests
make test

# Generate examples
make examples
```

## 📄 License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## 🌟 Why LastBackend Toolkit?

**Before:**
```go
// Hundreds of lines of boilerplate
server := grpc.NewServer()
mux := http.NewServeMux()
db, _ := gorm.Open(postgres.Open(dsn))
redis := redis.NewClient(&redis.Options{})
// ... setup middleware, routing, health checks, etc.
```

**After:**
```protobuf
// Just describe what you want
option (toolkit.plugins) = {
  prefix: "pgsql" 
  plugin: "postgres_gorm"
};

service UserService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]
  };
  // ... your business logic
}
```

**Focus on what matters:** Your business logic, not infrastructure setup.

---

<div align="center">

**[Get Started](docs/getting-started.md)** • **[Examples](examples/)** • **[Discord](https://discord.gg/WhK9ujvem9)** • **[Documentation](docs/)**

Made with ❤️ by the [LastBackend](https://github.com/lastbackend) team

</div>