# Getting Started with LastBackend Toolkit

This guide will help you get up and running with the LastBackend Toolkit quickly.

## Prerequisites

### Required Software
- **Go 1.21+** - [Download and install Go](https://golang.org/dl/)
- **Protocol Buffers compiler (protoc)** - [Installation guide](https://grpc.io/docs/protoc-installation/)
- **Git** - For version control

### Required Tools Installation

Install the necessary protoc plugins:

```bash
# Protocol Buffers Go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# gRPC Go plugin
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Validation plugin
go install github.com/envoyproxy/protoc-gen-validate@latest

# LastBackend Toolkit plugin (main code generator)
go install github.com/lastbackend/toolkit/protoc-gen-toolkit@latest

# Toolkit CLI (optional but recommended)
go install github.com/lastbackend/toolkit/cli@latest
```

Verify your PATH includes the Go bin directory:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

## Quick Start

### Option 1: Using the CLI (Recommended)

Create a new microservice with plugins:

```bash
# Create a CRUD microservice with PostgreSQL and Redis
toolkit init service github.com/yourorg/user-service --postgres-gorm --redis

cd user-service

# Initialize dependencies and generate code
make init proto update tidy

# Run the service
go run main.go
```

### Option 2: Manual Setup

1. **Create project structure:**
   ```bash
   mkdir user-service && cd user-service
   mkdir -p {apis,config,internal/{server,repository,service},scripts,tests}
   ```

2. **Initialize Go module:**
   ```bash
   go mod init github.com/yourorg/user-service
   ```

3. **Add toolkit dependency:**
   ```bash
   go get github.com/lastbackend/toolkit@latest
   go get github.com/lastbackend/toolkit-plugins/postgres_gorm@latest
   go get github.com/lastbackend/toolkit-plugins/redis@latest
   ```

## Your First Service

### 1. Define Your Service in Proto

Create `apis/user-service.proto`:

```protobuf
syntax = "proto3";
package userservice;

option go_package = "github.com/yourorg/user-service/gen;servicepb";

import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";
import "google/api/annotations.proto";
import "validate/validate.proto";

// Configure plugins
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

option (toolkit.plugins) = {
  prefix: "cache"
  plugin: "redis"
};

service UserService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]  // Enable both gRPC and HTTP servers
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
}
```

### 2. Create Generation Script

Create `scripts/generate.sh`:

```bash
#!/bin/bash -e

SOURCE_PACKAGE=github.com/yourorg/user-service
ROOT_DIR=$GOPATH/src/$SOURCE_PACKAGE
PROTO_DIR=$ROOT_DIR/apis

# Clean previous generation
find $ROOT_DIR -type f \( -name '*.pb.go' -o -name '*.pb.*.go' \) -delete 2>/dev/null || true

# Setup proto dependencies
mkdir -p $PROTO_DIR/google/api
mkdir -p $PROTO_DIR/validate

# Download required proto files
curl -s -o $PROTO_DIR/google/api/annotations.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
curl -s -o $PROTO_DIR/google/api/http.proto -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto
curl -s -o $PROTO_DIR/validate/validate.proto -L https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/main/validate/validate.proto

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
rm -rf $PROTO_DIR/google $PROTO_DIR/validate

echo "Code generation completed successfully"
```

Make it executable:
```bash
chmod +x scripts/generate.sh
```

### 3. Create Go Generate File

Create `generate.go`:

```go
package main

//go:generate ./scripts/generate.sh
```

### 4. Generate Code

```bash
go generate ./...
```

This creates the following files:
- `gen/user-service.pb.go` - Protobuf messages
- `gen/user-service_grpc.pb.go` - gRPC service definitions
- `gen/user-service.pb.validate.go` - Input validation
- `gen/user-service_service.pb.toolkit.go` - Service infrastructure (main file)

### 5. Create Application Entry Point

Create `main.go`:

```go
package main

import (
    "context"
    "os"

    "github.com/yourorg/user-service/config"
    servicepb "github.com/yourorg/user-service/gen"
    "github.com/yourorg/user-service/internal/server"
    "github.com/lastbackend/toolkit/pkg/runtime"
)

func main() {
    // Create service with plugins automatically initialized
    app, err := servicepb.NewUserServiceService("user-service",
        runtime.WithVersion("1.0.0"),
        runtime.WithDescription("User management service"),
        runtime.WithEnvPrefix("USER_SERVICE"),
    )
    if err != nil {
        panic(err)
    }

    // Register application configuration
    cfg := config.New()
    if err := app.RegisterConfig(cfg); err != nil {
        app.Log().Error(err)
        return
    }

    // Register your business logic
    app.Server().GRPC().SetService(server.NewUserServer)

    // Start the service (both gRPC and HTTP servers)
    if err := app.Start(context.Background()); err != nil {
        app.Log().Errorf("failed to start service: %v", err)
        os.Exit(1)
    }
}
```

### 6. Create Configuration

Create `config/config.go`:

```go
package config

import "time"

// Application-specific configuration
// Plugin configs (database, redis) are handled automatically
type Config struct {
    AppName     string        `env:"APP_NAME" envDefault:"user-service"`
    Environment string        `env:"ENVIRONMENT" envDefault:"development"`
    LogLevel    string        `env:"LOG_LEVEL" envDefault:"info"`
    MaxRetries  int           `env:"MAX_RETRIES" envDefault:"3"`
    Timeout     time.Duration `env:"TIMEOUT" envDefault:"30s"`
}

func New() *Config {
    return &Config{}
}
```

### 7. Implement Business Logic

Create `internal/server/server.go`:

```go
package server

import (
    "context"

    servicepb "github.com/yourorg/user-service/gen"
    "github.com/lastbackend/toolkit"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type UserServer struct {
    servicepb.UserServiceRpcServer
    
    app   toolkit.Service
    pgsql servicepb.PgsqlPlugin  // Auto-injected database
    cache servicepb.CachePlugin  // Auto-injected cache
}

func NewUserServer(
    app toolkit.Service,
    pgsql servicepb.PgsqlPlugin,
    cache servicepb.CachePlugin,
) servicepb.UserServiceRpcServer {
    return &UserServer{
        app:   app,
        pgsql: pgsql,
        cache: cache,
    }
}

func (s *UserServer) CreateUser(ctx context.Context, req *servicepb.CreateUserRequest) (*servicepb.CreateUserResponse, error) {
    // Input validation (automatic via protoc-gen-validate)
    if err := req.Validate(); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }
    
    // Your business logic here
    // Use s.pgsql.DB() for database operations
    // Use s.cache.Client() for Redis operations
    
    return &servicepb.CreateUserResponse{
        UserId: "generated-user-id",
        Status: "created",
    }, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *servicepb.GetUserRequest) (*servicepb.GetUserResponse, error) {
    if err := req.Validate(); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }
    
    // Your business logic here
    
    return &servicepb.GetUserResponse{
        UserId: req.UserId,
        Name:   "John Doe",
        Email:  "john@example.com",
    }, nil
}
```

### 8. Set Environment Variables

Create `.env` file or set environment variables:

```bash
# Service configuration
USER_SERVICE_APP_NAME=user-service
USER_SERVICE_ENVIRONMENT=development
USER_SERVICE_LOG_LEVEL=info

# PostgreSQL plugin configuration (auto-parsed)
USER_SERVICE_PGSQL_HOST=localhost
USER_SERVICE_PGSQL_PORT=5432
USER_SERVICE_PGSQL_USERNAME=postgres
USER_SERVICE_PGSQL_PASSWORD=secret
USER_SERVICE_PGSQL_DATABASE=userservice
USER_SERVICE_PGSQL_SSL_MODE=disable

# Redis plugin configuration (auto-parsed)
USER_SERVICE_CACHE_HOST=localhost
USER_SERVICE_CACHE_PORT=6379
USER_SERVICE_CACHE_PASSWORD=
USER_SERVICE_CACHE_DATABASE=0
```

### 9. Run Your Service

Start required services:
```bash
# Start PostgreSQL and Redis (using Docker)
docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=secret postgres:15
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

Run your service:
```bash
go run main.go
```

Your service will start with:
- gRPC server on port 9090
- HTTP server on port 8080
- Automatic plugin initialization and health checks

### 10. Test Your Service

Test gRPC endpoint:
```bash
# Install grpcurl for testing
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Test CreateUser
grpcurl -plaintext -d '{"name":"John Doe","email":"john@example.com"}' \
  localhost:9090 userservice.UserService/CreateUser
```

Test HTTP endpoint:
```bash
# Test CreateUser via HTTP
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'

# Test GetUser via HTTP  
curl http://localhost:8080/users/123e4567-e89b-12d3-a456-426614174000
```

## Next Steps

1. **Add More Business Logic** - Implement your domain-specific logic in the server layer
2. **Add Layers** - Create repository, service, and controller layers for better organization
3. **Add Tests** - Write unit and integration tests using generated mocks
4. **Add More Plugins** - Integrate additional plugins like RabbitMQ, Sentry, etc.
5. **Deploy** - Use Docker and Kubernetes for production deployment

## Project Structure

Your completed project structure should look like:

```
user-service/
├── .env                        # Environment variables
├── .gitignore                  # Git ignore file
├── generate.go                 # Go generate directive
├── go.mod                      # Go module
├── go.sum                      # Go dependencies
├── main.go                     # Application entry point
├── apis/                       # Proto definitions
│   └── user-service.proto
├── config/                     # Application configuration
│   └── config.go
├── gen/                        # Generated code (auto-generated)
│   ├── *.pb.go
│   ├── *_grpc.pb.go
│   ├── *.pb.validate.go
│   └── *_service.pb.toolkit.go
├── internal/                   # Internal packages
│   ├── server/                 # gRPC/HTTP handlers
│   ├── service/                # Business logic
│   ├── repository/             # Data access
│   └── controller/             # Controllers
├── scripts/                    # Build scripts
│   └── generate.sh
└── tests/                      # Test files
```

## Common Issues

### Plugin Connection Issues
```
Error: plugin PreStart failed: connection refused
```
**Solution**: Ensure PostgreSQL and Redis are running and accessible with correct credentials.

### Environment Variables Not Found
```
Error: required environment variable not set: USER_SERVICE_PGSQL_USERNAME
```
**Solution**: Check environment variable naming follows pattern: `{SERVICE_PREFIX}_{PLUGIN_PREFIX}_{SETTING}`

### Code Generation Fails
```
Error: protoc-gen-toolkit: program not found
```
**Solution**: Ensure all protoc plugins are installed and `$GOPATH/bin` is in your `$PATH`.

## Getting Help

- **Documentation**: Check the [full documentation](../README.md)
- **Examples**: See [examples directory](../examples/) for real-world examples
- **Discord**: Join our [Discord community](https://discord.gg/WhK9ujvem9)
- **Issues**: Report issues on [GitHub](https://github.com/lastbackend/toolkit/issues)

You're now ready to build production-ready microservices with the LastBackend Toolkit!