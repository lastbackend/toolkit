# LastBackend Toolkit - AI Agent Instructions

This document provides comprehensive instructions for AI agents (like Claude Code) to effectively work with the LastBackend Toolkit framework.

## Overview for AI Agents

LastBackend Toolkit is a Go microservices framework that uses:
- **Protobuf-first development** - All services defined in `.proto` files
- **Automatic code generation** - `protoc-gen-toolkit` generates service infrastructure
- **Plugin-based architecture** - External services integrated as plugins
- **Multi-protocol support** - gRPC, HTTP, WebSocket from single definition
- **Dependency injection** - Built on Uber FX framework

## Key Principles for AI Generation

### 1. Always Start with Proto Definitions
Every service begins with a `.proto` file that declares:
- Service interfaces and methods
- Plugin requirements
- Server types needed
- Message validation rules

### 2. Plugin-Driven Configuration
Most configuration comes from plugins automatically. Don't create manual database or cache configs - declare plugins and they handle their own configuration via environment variables.

### 3. Generated Code is King
Never write boilerplate manually. The toolkit generates:
- Service constructors with plugin initialization
- Plugin interfaces for dependency injection
- Client libraries for inter-service communication
- Mock interfaces for testing

## Quick Reference

### Essential Imports for Proto Files
```protobuf
import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";
import "google/api/annotations.proto";
import "validate/validate.proto";
```

### Plugin Declaration Pattern
```protobuf
// Database plugin
option (toolkit.plugins) = {
  prefix: "pgsql"           // Environment prefix: MYSERVICE_PGSQL_*
  plugin: "postgres_gorm"   // Plugin type from toolkit-plugins
};

// Cache plugin
option (toolkit.plugins) = {
  prefix: "cache"
  plugin: "redis"
};
```

### Service Runtime Configuration
```protobuf
service MyService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]    // Multiple protocol support
  };
  option (toolkit.server) = {
    middlewares: ["request_id", "auth"]  // Global middleware
  };
}
```

### Environment Variable Pattern
```bash
# Service configuration (from runtime.WithEnvPrefix)
MYSERVICE_APP_NAME=my-service
MYSERVICE_ENVIRONMENT=production

# Plugin configurations (auto-parsed by plugins)
MYSERVICE_PGSQL_HOST=localhost
MYSERVICE_PGSQL_USERNAME=user
MYSERVICE_PGSQL_PASSWORD=secret

MYSERVICE_CACHE_HOST=redis.example.com
MYSERVICE_CACHE_PASSWORD=redispass
```

## Available Plugins Reference

| Plugin | Package | Environment Prefix | Key Interface Method |
|--------|---------|-------------------|----------------------|
| **postgres_gorm** | `postgres_gorm` | `{PREFIX}_PGSQL_*` | `Plugin.DB() *gorm.DB` |
| **postgres_pg** | `postgres_pg` | `{PREFIX}_PG_*` | `Plugin.DB() *pg.DB` |
| **redis** | `redis` | `{PREFIX}_REDIS_*` | `Plugin.Client() redis.Cmdable` |
| **rabbitmq** | `rabbitmq` | `{PREFIX}_QUEUE_*` | `Plugin.Publish/Subscribe` |
| **centrifuge** | `centrifuge` | `{PREFIX}_CENTRIFUGE_*` | `Plugin.Node() *centrifuge.Node` |
| **sentry** | `sentry` | `{PREFIX}_SENTRY_*` | Error tracking |

## Common Patterns

### 1. CRUD Microservice Pattern
```protobuf
// Required plugins
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
    servers: [GRPC, HTTP]
  };
  
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = { post: "/users" body: "*" };
  };
  
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = { get: "/users/{user_id}" };
  };
  
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option (google.api.http) = { put: "/users/{user_id}" body: "*" };
  };
  
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
    option (google.api.http) = { delete: "/users/{user_id}" };
  };
  
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {
    option (google.api.http) = { get: "/users" };
  };
}
```

### 2. Event-Driven Service Pattern
```protobuf
// Required plugins for event-driven architecture
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

option (toolkit.plugins) = {
  prefix: "queue"
  plugin: "rabbitmq"
};

service EventService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]
  };
  
  rpc PublishEvent(PublishEventRequest) returns (PublishEventResponse) {
    option (google.api.http) = { post: "/events" body: "*" };
  };
}
```

### 3. Gateway Service Pattern
```protobuf
// No plugins needed for pure gateway
service ApiGateway {
  option (toolkit.runtime) = {
    servers: [HTTP, WEBSOCKET_PROXY]
  };
  
  rpc ProxyToUserService(userservice.GetUserRequest) returns (userservice.GetUserResponse) {
    option (toolkit.route) = {
      http_proxy: {
        service: "user-service"
        method: "/userservice.UserService/GetUser"
      }
    };
    option (google.api.http) = { get: "/api/v1/users/{user_id}" };
  };
}
```

### 4. Real-time Service Pattern
```protobuf
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

option (toolkit.plugins) = {
  prefix: "realtime"
  plugin: "centrifuge"
};

service RealtimeService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP, WEBSOCKET]
  };
  
  rpc Subscribe(SubscribeRequest) returns (SubscribeResponse) {
    option (toolkit.route).websocket = true;
    option (google.api.http) = { get: "/subscribe" };
  };
}
```

## Code Generation Guidelines

### 1. Service Constructor Pattern
```go
// main.go
func main() {
    app, err := servicepb.NewMyServiceService("my-service",
        runtime.WithVersion("1.0.0"),
        runtime.WithDescription("My microservice"),
        runtime.WithEnvPrefix("MYSERVICE"),  // Important: sets env prefix
    )
    if err != nil {
        panic(err)
    }

    // Register application config (business logic only)
    cfg := config.New()
    if err := app.RegisterConfig(cfg); err != nil {
        app.Log().Error(err)
        return
    }

    // Register packages (dependency injection)
    app.RegisterPackage(repository.NewRepository, controller.NewController)

    // Configure servers
    app.Server().GRPC().SetService(server.NewMyServiceServer)

    // Start service
    if err := app.Start(context.Background()); err != nil {
        app.Log().Errorf("failed to start service: %v", err)
        os.Exit(1)
    }
}
```

### 2. Plugin Usage Pattern
```go
// internal/repository/repository.go
type Repository struct {
    db  servicepb.PgsqlPlugin  // Auto-injected database plugin
    log toolkit.Logger
}

// Constructor with plugin dependency injection
func NewRepository(app toolkit.Service, db servicepb.PgsqlPlugin) *Repository {
    return &Repository{
        db:  db,
        log: app.Log(),
    }
}

// Use plugin interface methods
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    return r.db.DB().WithContext(ctx).Create(user).Error
}

func (r *Repository) GetUser(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.DB().WithContext(ctx).Where("id = ?", id).First(&user).Error
    return &user, err
}
```

### 3. Multi-Plugin Service Pattern
```go
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
    
    // Cache user (best effort)
    if userData, err := json.Marshal(user); err == nil {
        s.cache.Client().Set(ctx, fmt.Sprintf("user:%s", user.ID), userData, time.Hour)
    }
    
    // Publish event (best effort)
    if eventData, err := json.Marshal(map[string]interface{}{
        "type": "user_created",
        "user_id": user.ID,
    }); err == nil {
        s.queue.Publish("user.created", eventData)
    }
    
    return nil
}
```

### 4. Configuration Pattern
```go
// config/config.go - Only application-specific config
type Config struct {
    // Application settings
    AppName     string `env:"APP_NAME" envDefault:"my-service"`
    Environment string `env:"ENVIRONMENT" envDefault:"development"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
    
    // Business logic settings
    MaxRetries      int           `env:"MAX_RETRIES" envDefault:"3"`
    RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" envDefault:"30s"`
    EnableFeatureX  bool          `env:"ENABLE_FEATURE_X" envDefault:"false"`
}

func New() *Config {
    return &Config{}
    // No manual parsing - toolkit handles this automatically
}
```

## Error Handling Best Practices

### 1. Plugin Error Handling
```go
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    // Try cache first (non-critical)
    if cached, err := s.cache.Get(ctx, fmt.Sprintf("user:%s", id)); err == nil {
        var user User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            return &user, nil
        }
    }
    
    // Fallback to database (critical)
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user from database: %w", err)
    }
    
    // Update cache (best effort)
    if userData, err := json.Marshal(user); err == nil {
        s.cache.Set(ctx, fmt.Sprintf("user:%s", id), userData, time.Hour)
    }
    
    return user, nil
}
```

### 2. gRPC Error Responses
```go
func (s *Server) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    // Input validation
    if err := req.Validate(); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }
    
    user, err := s.service.GetUser(ctx, req.UserId)
    if err != nil {
        if errors.Is(err, repository.ErrUserNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        s.app.Log().WithError(err).Error("failed to get user")
        return nil, status.Error(codes.Internal, "internal server error")
    }
    
    return &GetUserResponse{
        UserId:    user.ID,
        Name:      user.Name,
        Email:     user.Email,
        CreatedAt: timestamppb.New(user.CreatedAt),
    }, nil
}
```

## Testing Patterns

### 1. Unit Testing with Mocks
```go
func TestService_CreateUser(t *testing.T) {
    // Setup mocks
    mockRepo := new(tests.MockRepository)
    mockCache := new(tests.MockCachePlugin)
    mockQueue := new(tests.MockQueuePlugin)
    
    service := NewService(mockRepo, mockCache, mockQueue)
    
    // Setup expectations
    user := &User{Name: "John", Email: "john@example.com"}
    mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *User) bool {
        return u.Name == "John" && u.Email == "john@example.com"
    })).Return(nil)
    
    mockCache.On("Client").Return(&tests.MockRedisClient{})
    mockQueue.On("Publish", "user.created", mock.Anything).Return(nil)
    
    // Execute
    err := service.CreateUser(context.Background(), &CreateUserRequest{
        Name:  "John",
        Email: "john@example.com",
    })
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

## Project Structure Template

```
myservice/
├── .ai/                        # AI agent instructions (copy from toolkit)
├── apis/                       # Proto definitions
│   ├── myservice.proto
│   └── ptypes/                 # Shared message types
├── cmd/                        # Application entry points
│   └── server/
│       └── main.go
├── config/                     # Application config
│   └── config.go
├── gen/                        # Generated code (auto-generated)
├── internal/                   # Internal packages
│   ├── controller/             # Business logic
│   ├── repository/             # Data access
│   ├── server/                 # gRPC/HTTP handlers
│   └── service/               # Service layer
├── scripts/                    # Build scripts
│   └── generate.sh
├── tests/                      # Test files
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Command Reference

### Code Generation
```bash
# Generate all code
go generate ./...

# Or run generation script directly
./scripts/generate.sh
```

### Environment Setup
```bash
# Required tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/envoyproxy/protoc-gen-validate@latest
go install github.com/lastbackend/toolkit/protoc-gen-toolkit@latest
```

## Common Mistakes to Avoid

1. **Don't create manual database configs** - Use plugins
2. **Don't hardcode connection strings** - Use environment variables
3. **Don't skip validation** - Always add validate annotations
4. **Don't ignore plugin errors in critical paths** - Handle gracefully
5. **Don't forget environment prefixes** - Use `runtime.WithEnvPrefix`
6. **Don't mix business logic with infrastructure** - Keep clean separation

## AI Decision Matrix

| Use Case | Plugins Needed | Server Types | Focus Areas |
|----------|----------------|--------------|-------------|
| **Simple API** | None | `[HTTP]` | REST endpoints, validation |
| **CRUD Service** | `postgres_gorm`, `redis` | `[GRPC, HTTP]` | Database operations, caching |
| **Event Service** | `postgres_gorm`, `rabbitmq` | `[GRPC, HTTP]` | Event publishing, queuing |
| **Real-time** | `postgres_gorm`, `centrifuge` | `[GRPC, HTTP, WEBSOCKET]` | WebSocket handling, real-time |
| **Gateway** | None | `[HTTP, WEBSOCKET_PROXY]` | Proxy configuration, routing |
| **Worker** | `postgres_gorm`, `rabbitmq` | `[GRPC]` | Background processing |

Use this guide to generate consistent, well-structured LastBackend Toolkit projects that follow framework conventions and best practices.