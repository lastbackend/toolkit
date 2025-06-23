# LastBackend Toolkit - Examples for AI Agents

This document provides specific examples and code patterns for AI agents to understand and replicate when working with the LastBackend Toolkit.

## Complete Project Examples

### 1. User Management Microservice

**Use Case**: CRUD operations for user management with caching

**Proto Definition** (`apis/user-service.proto`):
```protobuf
syntax = "proto3";
package userservice;

option go_package = "github.com/example/user-service/gen;servicepb";

import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";
import "google/api/annotations.proto";
import "validate/validate.proto";
import "google/protobuf/timestamp.proto";

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

// Mock generation
option (toolkit.tests_spec) = {
  mockery: {
    package: "github.com/example/user-service/gen/tests"
  }
};

service UserService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]
  };
  option (toolkit.server) = {
    middlewares: ["request_id", "logging"]
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

  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option (google.api.http) = {
      put: "/users/{user_id}"
      body: "*"
    };
  };

  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
    option (google.api.http) = {
      delete: "/users/{user_id}"
    };
  };

  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {
    option (google.api.http) = {
      get: "/users"
    };
  };
}

message User {
  string user_id = 1;
  string name = 2;
  string email = 3;
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp updated_at = 5;
}

message CreateUserRequest {
  string name = 1 [(validate.rules).string.min_len = 1, (validate.rules).string.max_len = 100];
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
  User user = 1;
}

message UpdateUserRequest {
  string user_id = 1 [(validate.rules).string.uuid = true];
  optional string name = 2 [(validate.rules).string.min_len = 1, (validate.rules).string.max_len = 100];
  optional string email = 3 [(validate.rules).string.pattern = "^[^@]+@[^@]+\\.[^@]+$"];
}

message UpdateUserResponse {
  User user = 1;
}

message DeleteUserRequest {
  string user_id = 1 [(validate.rules).string.uuid = true];
}

message DeleteUserResponse {
  string status = 1;
}

message ListUsersRequest {
  int32 limit = 1 [(validate.rules).int32.gte = 1, (validate.rules).int32.lte = 100];
  int32 offset = 2 [(validate.rules).int32.gte = 0];
  optional string search = 3;
}

message ListUsersResponse {
  repeated User users = 1;
  int32 total = 2;
}
```

**Main Application** (`main.go`):
```go
package main

import (
    "context"
    "os"

    "github.com/example/user-service/config"
    servicepb "github.com/example/user-service/gen"
    "github.com/example/user-service/internal/controller"
    "github.com/example/user-service/internal/repository"
    "github.com/example/user-service/internal/server"
    "github.com/example/user-service/internal/service"
    "github.com/lastbackend/toolkit/pkg/runtime"
)

func main() {
    app, err := servicepb.NewUserServiceService("user-service",
        runtime.WithVersion("1.0.0"),
        runtime.WithDescription("User management microservice"),
        runtime.WithEnvPrefix("USER_SERVICE"),
    )
    if err != nil {
        panic(err)
    }

    // Application configuration
    cfg := config.New()
    if err := app.RegisterConfig(cfg); err != nil {
        app.Log().Error(err)
        return
    }

    // Register packages (dependency injection chain)
    app.RegisterPackage(
        repository.NewUserRepository,
        service.NewUserService,
        controller.NewUserController,
    )

    // Configure gRPC server
    app.Server().GRPC().SetService(server.NewUserServer)

    // Start service
    if err := app.Start(context.Background()); err != nil {
        app.Log().Errorf("failed to start service: %v", err)
        os.Exit(1)
    }
}
```

**Repository Layer** (`internal/repository/user_repository.go`):
```go
package repository

import (
    "context"
    "fmt"
    "time"

    servicepb "github.com/example/user-service/gen"
    "github.com/google/uuid"
    "github.com/lastbackend/toolkit"
    "gorm.io/gorm"
)

type User struct {
    ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Name      string    `gorm:"not null"`
    Email     string    `gorm:"uniqueIndex;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type UserRepository struct {
    db  servicepb.PgsqlPlugin
    log toolkit.Logger
}

func NewUserRepository(app toolkit.Service, db servicepb.PgsqlPlugin) *UserRepository {
    repo := &UserRepository{
        db:  db,
        log: app.Log(),
    }
    
    // Auto-migrate database schema
    if err := db.DB().AutoMigrate(&User{}); err != nil {
        app.Log().WithError(err).Error("failed to migrate user table")
    }
    
    return repo
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    user.ID = uuid.New().String()
    return r.db.DB().WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.DB().WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err == gorm.ErrRecordNotFound {
        return nil, fmt.Errorf("user not found")
    }
    return &user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    err := r.db.DB().WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err == gorm.ErrRecordNotFound {
        return nil, nil // Not an error, just not found
    }
    return &user, err
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
    return r.db.DB().WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
    return r.db.DB().WithContext(ctx).Delete(&User{}, "id = ?", id).Error
}

func (r *UserRepository) List(ctx context.Context, limit, offset int, search string) ([]*User, int64, error) {
    var users []*User
    var total int64
    
    query := r.db.DB().WithContext(ctx).Model(&User{})
    
    if search != "" {
        query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
    }
    
    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Get paginated results
    if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
        return nil, 0, err
    }
    
    return users, total, nil
}
```

**Service Layer** (`internal/service/user_service.go`):
```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    servicepb "github.com/example/user-service/gen"
    "github.com/example/user-service/internal/repository"
    "github.com/lastbackend/toolkit"
)

type UserService struct {
    repo  *repository.UserRepository
    cache servicepb.CachePlugin
    log   toolkit.Logger
}

func NewUserService(
    app toolkit.Service,
    repo *repository.UserRepository,
    cache servicepb.CachePlugin,
) *UserService {
    return &UserService{
        repo:  repo,
        cache: cache,
        log:   app.Log(),
    }
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*repository.User, error) {
    // Check if user already exists
    if existing, _ := s.repo.GetByEmail(ctx, email); existing != nil {
        return nil, fmt.Errorf("user with email %s already exists", email)
    }

    user := &repository.User{
        Name:  name,
        Email: email,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    // Cache the user
    s.cacheUser(ctx, user)

    s.log.WithField("user_id", user.ID).Info("user created successfully")
    return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*repository.User, error) {
    // Try cache first
    if user := s.getCachedUser(ctx, id); user != nil {
        return user, nil
    }

    // Fallback to database
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    // Cache for future requests
    s.cacheUser(ctx, user)

    return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, name, email *string) (*repository.User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    // Update fields if provided
    if name != nil {
        user.Name = *name
    }
    if email != nil {
        // Check if email is already taken by another user
        if existing, _ := s.repo.GetByEmail(ctx, *email); existing != nil && existing.ID != id {
            return nil, fmt.Errorf("email %s is already taken", *email)
        }
        user.Email = *email
    }

    if err := s.repo.Update(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to update user: %w", err)
    }

    // Update cache
    s.cacheUser(ctx, user)

    s.log.WithField("user_id", user.ID).Info("user updated successfully")
    return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }

    // Remove from cache
    s.cache.Client().Del(ctx, fmt.Sprintf("user:%s", id))

    s.log.WithField("user_id", id).Info("user deleted successfully")
    return nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int, search string) ([]*repository.User, int64, error) {
    users, total, err := s.repo.List(ctx, limit, offset, search)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list users: %w", err)
    }

    return users, total, nil
}

// Helper methods for caching
func (s *UserService) cacheUser(ctx context.Context, user *repository.User) {
    data, err := json.Marshal(user)
    if err != nil {
        s.log.WithError(err).Warn("failed to marshal user for caching")
        return
    }

    if err := s.cache.Client().Set(ctx, fmt.Sprintf("user:%s", user.ID), data, time.Hour).Err(); err != nil {
        s.log.WithError(err).Warn("failed to cache user")
    }
}

func (s *UserService) getCachedUser(ctx context.Context, id string) *repository.User {
    data, err := s.cache.Client().Get(ctx, fmt.Sprintf("user:%s", id)).Result()
    if err != nil {
        return nil // Cache miss or error
    }

    var user repository.User
    if err := json.Unmarshal([]byte(data), &user); err != nil {
        s.log.WithError(err).Warn("failed to unmarshal cached user")
        return nil
    }

    return &user
}
```

**Server Layer** (`internal/server/user_server.go`):
```go
package server

import (
    "context"
    "errors"

    servicepb "github.com/example/user-service/gen"
    "github.com/example/user-service/internal/service"
    "github.com/lastbackend/toolkit"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type UserServer struct {
    servicepb.UserServiceRpcServer

    app     toolkit.Service
    service *service.UserService
}

func NewUserServer(
    app toolkit.Service,
    service *service.UserService,
) servicepb.UserServiceRpcServer {
    return &UserServer{
        app:     app,
        service: service,
    }
}

func (s *UserServer) CreateUser(ctx context.Context, req *servicepb.CreateUserRequest) (*servicepb.CreateUserResponse, error) {
    if err := req.Validate(); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }

    user, err := s.service.CreateUser(ctx, req.Name, req.Email)
    if err != nil {
        if errors.Is(err, service.ErrUserAlreadyExists) {
            return nil, status.Error(codes.AlreadyExists, err.Error())
        }
        s.app.Log().WithError(err).Error("failed to create user")
        return nil, status.Error(codes.Internal, "failed to create user")
    }

    return &servicepb.CreateUserResponse{
        UserId: user.ID,
        Status: "created",
    }, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *servicepb.GetUserRequest) (*servicepb.GetUserResponse, error) {
    if err := req.Validate(); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }

    user, err := s.service.GetUser(ctx, req.UserId)
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        s.app.Log().WithError(err).Error("failed to get user")
        return nil, status.Error(codes.Internal, "failed to get user")
    }

    return &servicepb.GetUserResponse{
        User: &servicepb.User{
            UserId:    user.ID,
            Name:      user.Name,
            Email:     user.Email,
            CreatedAt: timestamppb.New(user.CreatedAt),
            UpdatedAt: timestamppb.New(user.UpdatedAt),
        },
    }, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *servicepb.UpdateUserRequest) (*servicepb.UpdateUserResponse, error) {
    if err := req.Validate(); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }

    var name, email *string
    if req.Name != nil {
        name = req.Name
    }
    if req.Email != nil {
        email = req.Email
    }

    user, err := s.service.UpdateUser(ctx, req.UserId, name, email)
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        if errors.Is(err, service.ErrUserAlreadyExists) {
            return nil, status.Error(codes.AlreadyExists, err.Error())
        }
        s.app.Log().WithError(err).Error("failed to update user")
        return nil, status.Error(codes.Internal, "failed to update user")
    }

    return &servicepb.UpdateUserResponse{
        User: &servicepb.User{
            UserId:    user.ID,
            Name:      user.Name,
            Email:     user.Email,
            CreatedAt: timestamppb.New(user.CreatedAt),
            UpdatedAt: timestamppb.New(user.UpdatedAt),
        },
    }, nil
}

func (s *UserServer) DeleteUser(ctx context.Context, req *servicepb.DeleteUserRequest) (*servicepb.DeleteUserResponse, error) {
    if err := req.Validate(); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }

    if err := s.service.DeleteUser(ctx, req.UserId); err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        s.app.Log().WithError(err).Error("failed to delete user")
        return nil, status.Error(codes.Internal, "failed to delete user")
    }

    return &servicepb.DeleteUserResponse{
        Status: "deleted",
    }, nil
}

func (s *UserServer) ListUsers(ctx context.Context, req *servicepb.ListUsersRequest) (*servicepb.ListUsersResponse, error) {
    if err := req.Validate(); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }

    search := ""
    if req.Search != nil {
        search = *req.Search
    }

    users, total, err := s.service.ListUsers(ctx, int(req.Limit), int(req.Offset), search)
    if err != nil {
        s.app.Log().WithError(err).Error("failed to list users")
        return nil, status.Error(codes.Internal, "failed to list users")
    }

    var protoUsers []*servicepb.User
    for _, user := range users {
        protoUsers = append(protoUsers, &servicepb.User{
            UserId:    user.ID,
            Name:      user.Name,
            Email:     user.Email,
            CreatedAt: timestamppb.New(user.CreatedAt),
            UpdatedAt: timestamppb.New(user.UpdatedAt),
        })
    }

    return &servicepb.ListUsersResponse{
        Users: protoUsers,
        Total: int32(total),
    }, nil
}
```

**Configuration** (`config/config.go`):
```go
package config

import "time"

type Config struct {
    // Application settings
    AppName     string `env:"APP_NAME" envDefault:"user-service"`
    Environment string `env:"ENVIRONMENT" envDefault:"development"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

    // Business logic settings
    MaxRetries      int           `env:"MAX_RETRIES" envDefault:"3"`
    RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" envDefault:"30s"`
    CacheTimeout    time.Duration `env:"CACHE_TIMEOUT" envDefault:"1h"`
    
    // Feature flags
    EnableUserCache bool `env:"ENABLE_USER_CACHE" envDefault:"true"`
    EnableMetrics   bool `env:"ENABLE_METRICS" envDefault:"true"`
}

func New() *Config {
    return &Config{}
}
```

**Environment Variables** (`.env`):
```bash
# Service configuration
USER_SERVICE_APP_NAME=user-service
USER_SERVICE_ENVIRONMENT=development
USER_SERVICE_LOG_LEVEL=info

# PostgreSQL configuration (auto-parsed by plugin)
USER_SERVICE_PGSQL_HOST=localhost
USER_SERVICE_PGSQL_PORT=5432
USER_SERVICE_PGSQL_USERNAME=postgres
USER_SERVICE_PGSQL_PASSWORD=secret
USER_SERVICE_PGSQL_DATABASE=userservice
USER_SERVICE_PGSQL_SSL_MODE=disable
USER_SERVICE_PGSQL_DEBUG=true

# Redis configuration (auto-parsed by plugin)
USER_SERVICE_CACHE_HOST=localhost
USER_SERVICE_CACHE_PORT=6379
USER_SERVICE_CACHE_PASSWORD=
USER_SERVICE_CACHE_DATABASE=0

# Application settings
USER_SERVICE_MAX_RETRIES=3
USER_SERVICE_REQUEST_TIMEOUT=30s
USER_SERVICE_CACHE_TIMEOUT=1h
USER_SERVICE_ENABLE_USER_CACHE=true
```

### 2. Event-Driven Notification Service

**Use Case**: Microservice that processes events and sends notifications

**Proto Definition** (`apis/notification-service.proto`):
```protobuf
syntax = "proto3";
package notificationservice;

option go_package = "github.com/example/notification-service/gen;servicepb";

import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";
import "google/api/annotations.proto";
import "validate/validate.proto";
import "google/protobuf/timestamp.proto";

// Database for notification history
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

// Message queue for event processing
option (toolkit.plugins) = {
  prefix: "queue"
  plugin: "rabbitmq"
};

// Cache for user preferences
option (toolkit.plugins) = {
  prefix: "cache"
  plugin: "redis"
};

service NotificationService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]
  };

  rpc SendNotification(SendNotificationRequest) returns (SendNotificationResponse) {
    option (google.api.http) = {
      post: "/notifications"
      body: "*"
    };
  };

  rpc GetNotificationHistory(GetNotificationHistoryRequest) returns (GetNotificationHistoryResponse) {
    option (google.api.http) = {
      get: "/notifications/history/{user_id}"
    };
  };

  rpc UpdateNotificationPreferences(UpdateNotificationPreferencesRequest) returns (UpdateNotificationPreferencesResponse) {
    option (google.api.http) = {
      put: "/notifications/preferences/{user_id}"
      body: "*"
    };
  };
}

message SendNotificationRequest {
  string user_id = 1 [(validate.rules).string.uuid = true];
  string type = 2 [(validate.rules).string.in = "email", (validate.rules).string.in = "sms", (validate.rules).string.in = "push"];
  string subject = 3 [(validate.rules).string.min_len = 1, (validate.rules).string.max_len = 200];
  string message = 4 [(validate.rules).string.min_len = 1, (validate.rules).string.max_len = 2000];
  map<string, string> metadata = 5;
}

message SendNotificationResponse {
  string notification_id = 1;
  string status = 2;
}

message GetNotificationHistoryRequest {
  string user_id = 1 [(validate.rules).string.uuid = true];
  int32 limit = 2 [(validate.rules).int32.gte = 1, (validate.rules).int32.lte = 100];
  int32 offset = 3 [(validate.rules).int32.gte = 0];
}

message GetNotificationHistoryResponse {
  repeated Notification notifications = 1;
  int32 total = 2;
}

message UpdateNotificationPreferencesRequest {
  string user_id = 1 [(validate.rules).string.uuid = true];
  NotificationPreferences preferences = 2;
}

message UpdateNotificationPreferencesResponse {
  string status = 1;
}

message Notification {
  string notification_id = 1;
  string user_id = 2;
  string type = 3;
  string subject = 4;
  string message = 5;
  string status = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp sent_at = 8;
}

message NotificationPreferences {
  bool email_enabled = 1;
  bool sms_enabled = 2;
  bool push_enabled = 3;
  repeated string muted_types = 4;
}
```

**Event Processing** (`internal/service/event_processor.go`):
```go
package service

import (
    "context"
    "encoding/json"
    "fmt"

    servicepb "github.com/example/notification-service/gen"
    "github.com/lastbackend/toolkit"
)

type EventProcessor struct {
    queue   servicepb.QueuePlugin
    service *NotificationService
    log     toolkit.Logger
}

func NewEventProcessor(
    app toolkit.Service,
    queue servicepb.QueuePlugin,
    service *NotificationService,
) *EventProcessor {
    processor := &EventProcessor{
        queue:   queue,
        service: service,
        log:     app.Log(),
    }

    // Subscribe to events
    processor.subscribeToEvents()

    return processor
}

func (p *EventProcessor) subscribeToEvents() {
    // Subscribe to user events
    p.queue.Subscribe("notification-service", "user.created", p.handleUserCreated)
    p.queue.Subscribe("notification-service", "user.updated", p.handleUserUpdated)
    p.queue.Subscribe("notification-service", "order.completed", p.handleOrderCompleted)
}

func (p *EventProcessor) handleUserCreated(data []byte) error {
    var event struct {
        Type   string `json:"type"`
        UserID string `json:"user_id"`
        Name   string `json:"name"`
        Email  string `json:"email"`
    }

    if err := json.Unmarshal(data, &event); err != nil {
        return fmt.Errorf("failed to unmarshal user created event: %w", err)
    }

    // Send welcome notification
    return p.service.SendNotification(context.Background(), &SendNotificationRequest{
        UserId:  event.UserID,
        Type:    "email",
        Subject: "Welcome to our platform!",
        Message: fmt.Sprintf("Hi %s, welcome to our platform!", event.Name),
        Metadata: map[string]string{
            "event_type": "user_created",
            "email":      event.Email,
        },
    })
}

func (p *EventProcessor) handleUserUpdated(data []byte) error {
    // Handle user update events
    p.log.WithField("event", "user.updated").Info("processing user updated event")
    return nil
}

func (p *EventProcessor) handleOrderCompleted(data []byte) error {
    var event struct {
        Type     string  `json:"type"`
        UserID   string  `json:"user_id"`
        OrderID  string  `json:"order_id"`
        Amount   float64 `json:"amount"`
        Currency string  `json:"currency"`
    }

    if err := json.Unmarshal(data, &event); err != nil {
        return fmt.Errorf("failed to unmarshal order completed event: %w", err)
    }

    // Send order confirmation notification
    return p.service.SendNotification(context.Background(), &SendNotificationRequest{
        UserId:  event.UserID,
        Type:    "email",
        Subject: "Order Confirmation",
        Message: fmt.Sprintf("Your order #%s for %.2f %s has been completed!", event.OrderID, event.Amount, event.Currency),
        Metadata: map[string]string{
            "event_type": "order_completed",
            "order_id":   event.OrderID,
        },
    })
}
```

### 3. Real-time Chat Service

**Use Case**: WebSocket-based real-time chat with message persistence

**Proto Definition** (`apis/chat-service.proto`):
```protobuf
syntax = "proto3";
package chatservice;

option go_package = "github.com/example/chat-service/gen;servicepb";

import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";
import "google/api/annotations.proto";
import "validate/validate.proto";
import "google/protobuf/timestamp.proto";

// Database for message persistence
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

// Real-time messaging
option (toolkit.plugins) = {
  prefix: "realtime"
  plugin: "centrifuge"
};

// Cache for active users
option (toolkit.plugins) = {
  prefix: "cache"
  plugin: "redis"
};

service ChatService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP, WEBSOCKET]
  };

  rpc JoinRoom(JoinRoomRequest) returns (JoinRoomResponse) {
    option (toolkit.route).websocket = true;
    option (google.api.http) = {
      get: "/chat/rooms/{room_id}/join"
    };
  };

  rpc SendMessage(SendMessageRequest) returns (SendMessageResponse) {
    option (google.api.http) = {
      post: "/chat/rooms/{room_id}/messages"
      body: "*"
    };
  };

  rpc GetMessageHistory(GetMessageHistoryRequest) returns (GetMessageHistoryResponse) {
    option (google.api.http) = {
      get: "/chat/rooms/{room_id}/messages"
    };
  };

  rpc CreateRoom(CreateRoomRequest) returns (CreateRoomResponse) {
    option (google.api.http) = {
      post: "/chat/rooms"
      body: "*"
    };
  };
}

message JoinRoomRequest {
  string room_id = 1 [(validate.rules).string.uuid = true];
  string user_id = 2 [(validate.rules).string.uuid = true];
}

message JoinRoomResponse {
  string status = 1;
  string connection_id = 2;
}

message SendMessageRequest {
  string room_id = 1 [(validate.rules).string.uuid = true];
  string user_id = 2 [(validate.rules).string.uuid = true];
  string message = 3 [(validate.rules).string.min_len = 1, (validate.rules).string.max_len = 1000];
  string message_type = 4 [(validate.rules).string.in = "text", (validate.rules).string.in = "image", (validate.rules).string.in = "file"];
}

message SendMessageResponse {
  string message_id = 1;
  google.protobuf.Timestamp timestamp = 2;
}

message GetMessageHistoryRequest {
  string room_id = 1 [(validate.rules).string.uuid = true];
  int32 limit = 2 [(validate.rules).int32.gte = 1, (validate.rules).int32.lte = 100];
  int32 offset = 3 [(validate.rules).int32.gte = 0];
}

message GetMessageHistoryResponse {
  repeated ChatMessage messages = 1;
  int32 total = 2;
}

message CreateRoomRequest {
  string name = 1 [(validate.rules).string.min_len = 1, (validate.rules).string.max_len = 100];
  string creator_id = 2 [(validate.rules).string.uuid = true];
  repeated string participant_ids = 3;
}

message CreateRoomResponse {
  string room_id = 1;
  string status = 2;
}

message ChatMessage {
  string message_id = 1;
  string room_id = 2;
  string user_id = 3;
  string message = 4;
  string message_type = 5;
  google.protobuf.Timestamp timestamp = 6;
}
```

### 4. API Gateway Service

**Use Case**: HTTP gateway that routes requests to multiple backend services

**Proto Definition** (`apis/api-gateway.proto`):
```protobuf
syntax = "proto3";
package apigateway;

option go_package = "github.com/example/api-gateway/gen;servicepb";

import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";
import "google/api/annotations.proto";
import "validate/validate.proto";

// Import external service definitions
import "github.com/example/user-service/apis/user-service.proto";
import "github.com/example/notification-service/apis/notification-service.proto";

// Optional: Redis for API rate limiting
option (toolkit.plugins) = {
  prefix: "cache"
  plugin: "redis"
};

service ApiGateway {
  option (toolkit.runtime) = {
    servers: [HTTP, WEBSOCKET_PROXY]
  };
  option (toolkit.server) = {
    middlewares: ["cors", "rate_limit", "auth"]
  };

  // User service proxies
  rpc GetUser(userservice.GetUserRequest) returns (userservice.GetUserResponse) {
    option (toolkit.route) = {
      http_proxy: {
        service: "user-service"
        method: "/userservice.UserService/GetUser"
      }
      middlewares: ["validate_token"]
    };
    option (google.api.http) = {
      get: "/api/v1/users/{user_id}"
    };
  };

  rpc CreateUser(userservice.CreateUserRequest) returns (userservice.CreateUserResponse) {
    option (toolkit.route) = {
      http_proxy: {
        service: "user-service"
        method: "/userservice.UserService/CreateUser"
      }
      exclude_global_middlewares: ["auth"]  // Public endpoint
    };
    option (google.api.http) = {
      post: "/api/v1/users"
      body: "*"
    };
  };

  rpc ListUsers(userservice.ListUsersRequest) returns (userservice.ListUsersResponse) {
    option (toolkit.route) = {
      http_proxy: {
        service: "user-service"
        method: "/userservice.UserService/ListUsers"
      }
      middlewares: ["admin_only"]
    };
    option (google.api.http) = {
      get: "/api/v1/users"
    };
  };

  // Notification service proxies
  rpc SendNotification(notificationservice.SendNotificationRequest) returns (notificationservice.SendNotificationResponse) {
    option (toolkit.route) = {
      http_proxy: {
        service: "notification-service"
        method: "/notificationservice.NotificationService/SendNotification"
      }
    };
    option (google.api.http) = {
      post: "/api/v1/notifications"
      body: "*"
    };
  };

  // Health check endpoint (local, not proxied)
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse) {
    option (google.api.http) = {
      get: "/health"
    };
  };
}

message HealthCheckRequest {}

message HealthCheckResponse {
  string status = 1;
  map<string, string> services = 2;
}
```

## Testing Examples

### Unit Test with Mocks
```go
package service

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    servicepb "github.com/example/user-service/gen"
    "github.com/example/user-service/gen/tests"
    "github.com/example/user-service/internal/repository"
)

func TestUserService_CreateUser(t *testing.T) {
    // Setup mocks
    mockRepo := new(tests.MockUserRepository)
    mockCache := new(tests.MockCachePlugin)
    mockApp := new(tests.MockService)

    // Setup logger mock
    mockLogger := new(tests.MockLogger)
    mockApp.On("Log").Return(mockLogger)

    service := NewUserService(mockApp, mockRepo, mockCache)

    // Test data
    name := "John Doe"
    email := "john@example.com"

    // Mock expectations
    mockRepo.On("GetByEmail", mock.Anything, email).Return(nil, nil) // User doesn't exist
    mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *repository.User) bool {
        return user.Name == name && user.Email == email
    })).Return(nil).Run(func(args mock.Arguments) {
        user := args.Get(1).(*repository.User)
        user.ID = "test-user-id"
    })

    // Mock cache
    mockRedisClient := new(tests.MockRedisClient)
    mockCache.On("Client").Return(mockRedisClient)
    mockRedisClient.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

    // Mock logger
    mockLogger.On("WithField", "user_id", "test-user-id").Return(mockLogger)
    mockLogger.On("Info", "user created successfully").Return()

    // Execute
    user, err := service.CreateUser(context.Background(), name, email)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test-user-id", user.ID)
    assert.Equal(t, name, user.Name)
    assert.Equal(t, email, user.Email)

    // Verify all mocks were called as expected
    mockRepo.AssertExpectations(t)
    mockCache.AssertExpectations(t)
    mockRedisClient.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}

func TestUserService_CreateUser_EmailExists(t *testing.T) {
    // Setup mocks
    mockRepo := new(tests.MockUserRepository)
    mockCache := new(tests.MockCachePlugin)
    mockApp := new(tests.MockService)

    mockLogger := new(tests.MockLogger)
    mockApp.On("Log").Return(mockLogger)

    service := NewUserService(mockApp, mockRepo, mockCache)

    // Test data
    name := "John Doe"
    email := "john@example.com"
    existingUser := &repository.User{ID: "existing-id", Email: email}

    // Mock expectations - user already exists
    mockRepo.On("GetByEmail", mock.Anything, email).Return(existingUser, nil)

    // Execute
    user, err := service.CreateUser(context.Background(), name, email)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Contains(t, err.Error(), "already exists")

    mockRepo.AssertExpectations(t)
}
```

### Integration Test
```go
package integration

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "google.golang.org/grpc"
    "google.golang.org/grpc/test/bufconn"

    servicepb "github.com/example/user-service/gen"
)

func TestUserService_Integration(t *testing.T) {
    // Setup test server
    lis := bufconn.Listen(1024 * 1024)
    server := grpc.NewServer()

    // Register your service with test dependencies
    userServer := setupTestUserServer(t)
    servicepb.RegisterUserServiceServer(server, userServer)

    go func() {
        if err := server.Serve(lis); err != nil {
            t.Errorf("Server failed: %v", err)
        }
    }()
    defer server.Stop()

    // Create client connection
    conn, err := grpc.DialContext(context.Background(), "bufnet",
        grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
            return lis.Dial()
        }),
        grpc.WithInsecure(),
    )
    require.NoError(t, err)
    defer conn.Close()

    client := servicepb.NewUserServiceClient(conn)

    // Test create user
    createResp, err := client.CreateUser(context.Background(), &servicepb.CreateUserRequest{
        Name:  "John Doe",
        Email: "john@example.com",
    })
    require.NoError(t, err)
    assert.Equal(t, "created", createResp.Status)
    assert.NotEmpty(t, createResp.UserId)

    // Test get user
    getResp, err := client.GetUser(context.Background(), &servicepb.GetUserRequest{
        UserId: createResp.UserId,
    })
    require.NoError(t, err)
    assert.Equal(t, "John Doe", getResp.User.Name)
    assert.Equal(t, "john@example.com", getResp.User.Email)

    // Test list users
    listResp, err := client.ListUsers(context.Background(), &servicepb.ListUsersRequest{
        Limit:  10,
        Offset: 0,
    })
    require.NoError(t, err)
    assert.GreaterOrEqual(t, listResp.Total, int32(1))
    assert.Len(t, listResp.Users, 1)
}

func setupTestUserServer(t *testing.T) servicepb.UserServiceServer {
    // Setup test database and dependencies
    // Return configured server for testing
    // This would typically use test containers or in-memory databases
}
```

These examples provide comprehensive patterns for AI agents to understand and replicate when generating LastBackend Toolkit projects. Each example demonstrates proper separation of concerns, plugin usage, error handling, and testing strategies.