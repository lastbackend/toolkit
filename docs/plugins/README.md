# Plugin System Overview

The LastBackend Toolkit plugin system provides seamless integration with external services through a unified interface. Plugins handle initialization, configuration, lifecycle management, and provide type-safe interfaces for your application code.

## How Plugins Work

### 1. Declaration in Proto Files
Plugins are declared using toolkit annotations:

```protobuf
// Database plugin
option (toolkit.plugins) = {
  prefix: "pgsql"           // Environment variable prefix
  plugin: "postgres_gorm"   // Plugin type from toolkit-plugins
};

// Cache plugin
option (toolkit.plugins) = {
  prefix: "cache"
  plugin: "redis"
};
```

### 2. Automatic Code Generation
The toolkit generates:
- Plugin interfaces for dependency injection
- Plugin initialization code
- Environment variable parsing
- Lifecycle management hooks

### 3. Environment Configuration
Plugins automatically parse environment variables:
```bash
# PostgreSQL plugin (prefix: "pgsql")
MYSERVICE_PGSQL_HOST=localhost
MYSERVICE_PGSQL_PORT=5432
MYSERVICE_PGSQL_USERNAME=user
MYSERVICE_PGSQL_PASSWORD=secret

# Redis plugin (prefix: "cache")
MYSERVICE_CACHE_HOST=localhost
MYSERVICE_CACHE_PORT=6379
```

### 4. Dependency Injection
Plugins are automatically injected into your components:

```go
func NewRepository(app toolkit.Service, db servicepb.PgsqlPlugin) *Repository {
    return &Repository{
        db:  db.DB(), // *gorm.DB ready to use
        log: app.Log(),
    }
}
```

## Available Plugins

| Plugin | Package | Purpose | Interface |
|--------|---------|---------|-----------|
| [postgres_gorm](available-plugins.md#postgresql-gorm) | `postgres_gorm` | PostgreSQL with GORM ORM | `Plugin.DB() *gorm.DB` |
| [postgres_pg](available-plugins.md#postgresql-go-pg) | `postgres_pg` | PostgreSQL with go-pg | `Plugin.DB() *pg.DB` |
| [postgres_pgx](available-plugins.md#postgresql-pgx) | `postgres_pgx` | PostgreSQL with pgx driver | `Plugin.DB() *pgxpool.Pool` |
| [redis](available-plugins.md#redis) | `redis` | Redis cache/pub-sub | `Plugin.Client() redis.Cmdable` |
| [rabbitmq](available-plugins.md#rabbitmq) | `rabbitmq` | Message queue | `Plugin.Publish/Subscribe` |
| [centrifuge](available-plugins.md#centrifuge) | `centrifuge` | Real-time messaging | `Plugin.Node() *centrifuge.Node` |
| [sentry](available-plugins.md#sentry) | `sentry` | Error monitoring | Error tracking integration |
| [resolver_consul](available-plugins.md#consul) | `resolver_consul` | Service discovery | Consul integration |

## Plugin Architecture

### Lifecycle Management
Plugins follow a structured lifecycle:

1. **Initialization** - Plugin instances created with configuration
2. **PreStart** - Connections established, resources initialized
3. **OnStart** - Background processes started (async)
4. **Running** - Plugin available for use
5. **OnStop** - Graceful shutdown, resources cleaned up

### Configuration Hierarchy
Environment variables follow this pattern:
```
{SERVICE_PREFIX}_{PLUGIN_PREFIX}_{SETTING_NAME}
```

Example:
- Service prefix: `USER_SERVICE` (from `runtime.WithEnvPrefix`)
- Plugin prefix: `PGSQL` (from proto declaration)
- Setting: `HOST`
- Result: `USER_SERVICE_PGSQL_HOST=localhost`

### Multi-Instance Support
You can use multiple instances of the same plugin type:

```protobuf
// Primary database
option (toolkit.plugins) = {
  prefix: "primary_db"
  plugin: "postgres_gorm"
};

// Analytics database
option (toolkit.plugins) = {
  prefix: "analytics_db"
  plugin: "postgres_gorm"
};

// Session cache
option (toolkit.plugins) = {
  prefix: "session_cache"
  plugin: "redis"
};

// Data cache
option (toolkit.plugins) = {
  prefix: "data_cache"
  plugin: "redis"
};
```

## Quick Start

### 1. Choose Plugins
Determine which plugins your service needs:
- **Database**: Choose between GORM, go-pg, or pgx
- **Cache**: Redis for most use cases
- **Message Queue**: RabbitMQ for reliable messaging
- **Real-time**: Centrifuge for WebSocket connections
- **Monitoring**: Sentry for error tracking

### 2. Declare in Proto
```protobuf
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

option (toolkit.plugins) = {
  prefix: "cache"
  plugin: "redis"
};
```

### 3. Set Environment Variables
```bash
MYSERVICE_PGSQL_HOST=localhost
MYSERVICE_PGSQL_USERNAME=user
MYSERVICE_PGSQL_PASSWORD=secret

MYSERVICE_CACHE_HOST=localhost
MYSERVICE_CACHE_PORT=6379
```

### 4. Use in Code
```go
func NewRepository(app toolkit.Service, db servicepb.PgsqlPlugin) *Repository {
    return &Repository{db: db.DB()}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Create(user).Error
}
```

## Best Practices

### 1. Use Descriptive Prefixes
```protobuf
// ✅ Good - clear purpose
option (toolkit.plugins) = { prefix: "user_db", plugin: "postgres_gorm" };
option (toolkit.plugins) = { prefix: "session_cache", plugin: "redis" };

// ❌ Bad - unclear purpose
option (toolkit.plugins) = { prefix: "db1", plugin: "postgres_gorm" };
option (toolkit.plugins) = { prefix: "cache1", plugin: "redis" };
```

### 2. Handle Plugin Errors Gracefully
```go
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    // Try cache first (non-critical)
    if cached := s.getCachedUser(ctx, id); cached != nil {
        return cached, nil
    }
    
    // Fallback to database (critical)
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Update cache (best effort)
    s.cacheUser(ctx, user)
    
    return user, nil
}
```

### 3. Separate Critical from Non-Critical
- **Critical plugins** (database): Fail service startup if unavailable
- **Non-critical plugins** (cache): Degrade gracefully if unavailable

### 4. Use Plugin Health Checks
```go
func (s *Server) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
    status := "healthy"
    
    // Check critical plugins
    if err := s.db.DB().Raw("SELECT 1").Error; err != nil {
        return nil, status.Error(codes.Unavailable, "database unhealthy")
    }
    
    // Check non-critical plugins (don't fail)
    if err := s.cache.Client().Ping(ctx).Err(); err != nil {
        s.app.Log().Warn("cache unavailable")
    }
    
    return &HealthCheckResponse{Status: status}, nil
}
```

## Documentation

- **[Available Plugins](available-plugins.md)** - Detailed documentation for each plugin
- **[Plugin Development](development.md)** - Guide for creating custom plugins
- **[Complete Plugin Guide](PLUGIN_SYSTEM_GUIDE.md)** - Comprehensive plugin system documentation

## Examples

See the [examples directory](../../examples/) for real-world plugin usage:
- **[service/](../../examples/service/)** - PostgreSQL + Redis + RabbitMQ
- **[gateway/](../../examples/gateway/)** - Plugin-free API gateway
- **[wss/](../../examples/wss/)** - Redis + Centrifuge for real-time features

## Getting Help

- **Discord**: [Join our community](https://discord.gg/WhK9ujvem9)
- **Issues**: [Report plugin issues](https://github.com/lastbackend/toolkit/issues)
- **Plugin Repository**: [toolkit-plugins](https://github.com/lastbackend/toolkit-plugins)

The plugin system eliminates infrastructure boilerplate and lets you focus on business logic while maintaining type safety and clean architecture.