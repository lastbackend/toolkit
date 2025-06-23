# LastBackend Toolkit - Troubleshooting Guide for AI Agents

This guide helps AI agents diagnose and resolve common issues when working with the LastBackend Toolkit.

## Common Issues and Solutions

### 1. Plugin Configuration Issues

#### Plugin Not Found Error
```
Error: plugin "postgres_gorm" not found
```

**Causes:**
- Plugin not imported in go.mod
- Incorrect plugin name in proto file
- Missing plugin dependencies

**Solutions:**
```bash
# Add plugin dependency to go.mod
go get github.com/lastbackend/toolkit-plugins/postgres_gorm@latest

# Verify plugin name matches exactly
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"  // Must match package name exactly
};
```

#### Environment Variable Not Parsed
```
Error: required environment variable not set: MYSERVICE_PGSQL_USERNAME
```

**Causes:**
- Incorrect environment variable naming
- Missing service prefix configuration
- Wrong plugin prefix in proto file

**Solutions:**
```go
// Ensure service prefix is set in main.go
app, err := servicepb.NewMyServiceService("my-service",
    runtime.WithEnvPrefix("MYSERVICE"),  // This sets the prefix
)

// Environment variable pattern: {SERVICE_PREFIX}_{PLUGIN_PREFIX}_{SETTING}
MYSERVICE_PGSQL_USERNAME=user
MYSERVICE_PGSQL_PASSWORD=secret
```

#### Plugin Dependency Injection Failure
```
Error: failed to provide PgsqlPlugin: no constructor registered
```

**Causes:**
- Generated code not regenerated after proto changes
- Plugin not properly declared in proto file
- Missing plugin registration in generated code

**Solutions:**
```bash
# Regenerate code after proto changes
go generate ./...
# or
./scripts/generate.sh

# Verify plugin declaration in proto
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

# Check generated code includes registration
app.runtime.Plugin().Provide(func() PgsqlPlugin { return plugin_pgsql })
```

### 2. Code Generation Issues

#### Generated Code Out of Sync
```
Error: undefined: servicepb.PgsqlPlugin
```

**Causes:**
- Proto file changed but code not regenerated
- Missing imports in proto file
- Incorrect go_package option

**Solutions:**
```bash
# Always regenerate after proto changes
go generate ./...

# Verify required imports in proto file
import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";

# Check go_package option
option go_package = "github.com/yourorg/yourservice/gen;servicepb";
```

#### Protoc Plugin Not Found
```
Error: protoc-gen-toolkit: program not found or is not executable
```

**Solutions:**
```bash
# Install required protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/envoyproxy/protoc-gen-validate@latest
go install github.com/lastbackend/toolkit/protoc-gen-toolkit@latest

# Verify PATH includes GOPATH/bin
export PATH=$PATH:$(go env GOPATH)/bin
```

### 3. Service Startup Issues

#### Plugin Initialization Failed
```
Error: plugin PreStart failed: connection refused
```

**Causes:**
- External service (database, redis) not running
- Incorrect connection configuration
- Network connectivity issues

**Solutions:**
```bash
# Check external services are running
docker ps  # or systemctl status postgresql redis

# Verify connection configuration
MYSERVICE_PGSQL_HOST=localhost  # Correct host
MYSERVICE_PGSQL_PORT=5432       # Correct port

# Test connection manually
psql -h localhost -p 5432 -U user -d database
redis-cli -h localhost -p 6379 ping
```

#### Dependency Injection Cycle
```
Error: dependency cycle detected
```

**Causes:**
- Circular dependencies between components
- Incorrect constructor parameters

**Solutions:**
```go
// Avoid circular dependencies
// ❌ Bad: Repository depends on Service, Service depends on Repository
func NewRepository(service *Service) *Repository { ... }
func NewService(repo *Repository) *Service { ... }

// ✅ Good: Clean layer separation
func NewRepository(app toolkit.Service, db servicepb.PgsqlPlugin) *Repository { ... }
func NewService(app toolkit.Service, repo *Repository) *Service { ... }
```

### 4. Runtime Errors

#### Plugin Interface Panic
```
panic: interface conversion: *plugin.plugin does not implement servicepb.PgsqlPlugin
```

**Causes:**
- Plugin interface mismatch
- Generated code out of sync
- Plugin version incompatibility

**Solutions:**
```bash
# Regenerate all code
go generate ./...

# Update plugin dependencies
go get -u github.com/lastbackend/toolkit-plugins/postgres_gorm

# Verify plugin interface matches generated code
type PgsqlPlugin interface {
    postgres_gorm.Plugin  // Must match plugin package
}
```

#### Configuration Parsing Error
```
Error: env: required environment variable "PGSQL_USERNAME" is empty
```

**Causes:**
- Missing environment variables
- Incorrect variable naming
- Configuration not registered

**Solutions:**
```bash
# Use correct environment variable pattern
MYSERVICE_PGSQL_USERNAME=user  # Not just PGSQL_USERNAME

# Verify configuration registration in main.go
cfg := config.New()
if err := app.RegisterConfig(cfg); err != nil {
    app.Log().Error(err)
    return
}
```

### 5. HTTP/gRPC Issues

#### HTTP Routes Not Working
```
Error: 404 Not Found for /users/123
```

**Causes:**
- Missing HTTP annotations in proto
- Incorrect HTTP path mapping
- Server not configured for HTTP

**Solutions:**
```protobuf
// Add HTTP annotations
rpc GetUser(GetUserRequest) returns (GetUserResponse) {
  option (google.api.http) = {
    get: "/users/{user_id}"  // Correct path mapping
  };
};

// Enable HTTP server
service UserService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]  // Include HTTP
  };
}
```

#### gRPC Connection Refused
```
Error: connection refused: dial tcp 127.0.0.1:9090: connect: connection refused
```

**Causes:**
- gRPC server not started
- Wrong port configuration
- Service startup failure

**Solutions:**
```bash
# Check service is running
netstat -tulpn | grep :9090

# Verify server configuration
MYSERVICE_SERVER_GRPC_PORT=9090

# Check service logs for startup errors
```

### 6. Testing Issues

#### Mock Generation Failed
```
Error: mockery not found
```

**Solutions:**
```bash
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Configure mock generation in proto
option (toolkit.tests_spec) = {
  mockery: {
    package: "github.com/yourorg/yourservice/gen/tests"
  }
};

# Regenerate mocks
go generate ./...
```

#### Test Database Connection Issues
```
Error: test database connection failed
```

**Solutions:**
```go
// Use test-specific configuration
func TestRepository(t *testing.T) {
    // Use in-memory database or test container
    db := setupTestDB(t)
    repo := NewRepository(nil, &testPgsqlPlugin{db: db})
    
    // ... test code
}

// Or use mocks for unit tests
func TestService(t *testing.T) {
    mockRepo := new(tests.MockRepository)
    service := NewService(nil, mockRepo, nil, nil)
    
    // ... test with mocks
}
```

## Debugging Strategies

### 1. Enable Debug Logging
```bash
# Enable debug logging for plugins
MYSERVICE_PGSQL_DEBUG=true
MYSERVICE_LOG_LEVEL=debug

# Check plugin initialization
tail -f /var/log/myservice.log | grep -i plugin
```

### 2. Verify Environment Variables
```bash
# Print all environment variables with service prefix
env | grep MYSERVICE_ | sort

# Check specific plugin configuration
env | grep MYSERVICE_PGSQL_
```

### 3. Test Plugin Connections
```go
// Add health check endpoints
func (s *Server) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
    status := "healthy"
    
    // Test database connection
    if err := s.pgsql.DB().Raw("SELECT 1").Error; err != nil {
        s.app.Log().WithError(err).Error("database health check failed")
        status = "unhealthy"
    }
    
    // Test cache connection
    if err := s.cache.Client().Ping(ctx).Err(); err != nil {
        s.app.Log().WithError(err).Warn("cache health check failed")
        // Don't fail - cache is not critical
    }
    
    return &HealthCheckResponse{Status: status}, nil
}
```

### 4. Validate Generated Code
```bash
# Check generated files exist
ls -la gen/
ls -la gen/*.pb.go
ls -la gen/*service.pb.toolkit.go

# Verify plugin interfaces are generated
grep -n "Plugin interface" gen/*service.pb.toolkit.go

# Check plugin registration
grep -n "Plugin().Provide" gen/*service.pb.toolkit.go
```

## Common Anti-Patterns to Avoid

### 1. Manual Plugin Configuration
```go
// ❌ Don't do this - plugins handle their own config
type Config struct {
    Database DatabaseConfig `envPrefix:"DATABASE_"`
    Redis    RedisConfig    `envPrefix:"REDIS_"`
}

// ✅ Do this - minimal app config only
type Config struct {
    AppName     string `env:"APP_NAME" envDefault:"my-service"`
    Environment string `env:"ENVIRONMENT" envDefault:"development"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
}
```

### 2. Hardcoded Connection Strings
```go
// ❌ Don't do this
db, err := gorm.Open(postgres.Open("host=localhost user=postgres..."))

// ✅ Do this - use plugin
func NewRepository(app toolkit.Service, db servicepb.PgsqlPlugin) *Repository {
    return &Repository{db: db.DB()}  // Plugin handles connection
}
```

### 3. Ignoring Plugin Lifecycle
```go
// ❌ Don't do this - bypass plugin system
func main() {
    db := setupDatabaseManually()
    
    app, _ := servicepb.NewMyServiceService("my-service")
    // ... direct DB usage
}

// ✅ Do this - use plugin system
func main() {
    app, _ := servicepb.NewMyServiceService("my-service")
    // Plugins are automatically initialized via lifecycle hooks
    app.Start(context.Background())
}
```

### 4. Mixed Business Logic with Infrastructure
```go
// ❌ Don't do this
func (s *Server) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    // Direct database access in handler
    user := &User{Name: req.Name}
    s.db.Create(user)
    // ...
}

// ✅ Do this - clean layer separation
func (s *Server) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    user, err := s.service.CreateUser(ctx, req)  // Delegate to service layer
    if err != nil {
        return nil, s.handleError(err)
    }
    return &CreateUserResponse{UserId: user.ID}, nil
}
```

## Environment Setup Checklist

### Prerequisites
- [ ] Go 1.21+ installed
- [ ] protoc compiler installed
- [ ] Required protoc plugins installed
- [ ] External services (PostgreSQL, Redis) running
- [ ] Environment variables configured

### Project Setup
- [ ] Proto file with correct imports
- [ ] Plugin declarations with descriptive prefixes
- [ ] Service runtime configuration
- [ ] go.mod with required dependencies
- [ ] Generation script created
- [ ] Environment variables set with correct prefix

### Code Generation
- [ ] Run `go generate ./...` after proto changes
- [ ] Verify generated files exist
- [ ] Check plugin interfaces are generated
- [ ] Confirm plugin registration in generated code

### Runtime
- [ ] Service starts without errors
- [ ] Plugin connections established
- [ ] Health checks pass
- [ ] HTTP/gRPC endpoints accessible

Use this troubleshooting guide to quickly diagnose and resolve common issues when working with the LastBackend Toolkit.