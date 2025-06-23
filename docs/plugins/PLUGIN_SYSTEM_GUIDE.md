# LastBackend Toolkit - Plugin System Complete Guide

## Table of Contents

1. [Overview](#overview)
2. [Configuration System](#configuration-system)
3. [Plugin Architecture](#plugin-architecture)
4. [Available Plugins](#available-plugins)
5. [Plugin Development](#plugin-development)
6. [Integration Patterns](#integration-patterns)
7. [Lifecycle Management](#lifecycle-management)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Overview

The LastBackend Toolkit features a sophisticated plugin system that enables modular integration of external services (databases, caches, message queues, etc.) into microservices. The system provides declarative configuration, automatic dependency injection, lifecycle management, and type-safe interfaces.

### Key Features

- **Declarative Configuration** - Define plugins in protobuf files
- **Automatic Code Generation** - Generated integration code and interfaces
- **Dependency Injection** - Built on Uber FX framework
- **Lifecycle Management** - Structured startup/shutdown hooks
- **Multi-Instance Support** - Multiple instances of same plugin type
- **Environment Configuration** - Environment variable-based configuration
- **Type Safety** - Compile-time type checking for plugin usage

## Configuration System

### Environment Variable System

The toolkit uses the `github.com/caarlos0/env` library for configuration management with hierarchical environment variable support.

#### Configuration Structure

```go
type Config struct {
    Server   ServerConfig   `envPrefix:"SERVER_"`
    Database DatabaseConfig `envPrefix:"DATABASE_"`
    Redis    RedisConfig    `envPrefix:"REDIS_"`
}

type DatabaseConfig struct {
    Host     string `env:"HOST" envDefault:"localhost" comment:"Database host address"`
    Port     int    `env:"PORT" envDefault:"5432" comment:"Database port"`
    Username string `env:"USERNAME" required:"true" comment:"Database username"`
    Password string `env:"PASSWORD" required:"true" comment:"Database password"`
    Database string `env:"DATABASE" required:"true" comment:"Database name"`
    SSLMode  string `env:"SSL_MODE" envDefault:"disable" comment:"SSL mode configuration"`
}
```

#### Environment Variable Tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `env` | Environment variable name | `env:"HOST"` |
| `envDefault` | Default value if not set | `envDefault:"localhost"` |
| `envPrefix` | Prefix for nested structs | `envPrefix:"DATABASE_"` |
| `required` | Mark field as required | `required:"true"` |
| `comment` | Documentation for help output | `comment:"Database host"` |

#### Configuration Lifecycle

```go
func (c *configController) Parse(v interface{}, prefix string, opts ...env.Options) error {
    c.parsed[prefix] = v
    opts = append(opts, env.Options{Prefix: c.buildPrefix(prefix)})
    return env.Parse(v, opts...)
}

func (c *configController) buildPrefix(prefix string) string {
    pt := make([]string, 0)
    if c.prefix != "" {
        pt = append(pt, c.prefix)  // Service prefix (e.g., "MYSERVICE")
    }
    if prefix != "" {
        pt = append(pt, prefix)     // Plugin prefix (e.g., "PGSQL")
    }
    if len(pt) > 0 {
        pt = append(pt, "")
        return strings.ToUpper(strings.Join(pt, ConfigPrefixSeparator))
    }
    return ""
}
```

#### Example Environment Variables

```bash
# Service with prefix "MYSERVICE"
MYSERVICE_PGSQL_HOST=localhost
MYSERVICE_PGSQL_PORT=5432
MYSERVICE_PGSQL_USERNAME=user
MYSERVICE_PGSQL_PASSWORD=pass
MYSERVICE_PGSQL_DATABASE=mydb

# Redis configuration
MYSERVICE_REDIS_HOST=localhost
MYSERVICE_REDIS_PORT=6379
MYSERVICE_REDIS_PASSWORD=redispass
```

### Configuration Display

The toolkit provides utilities to display configuration options:

```go
// Print configuration table
func (c *configController) PrintTable(all, nocomments bool) string {
    tw := table.NewWriter()
    tw.AppendHeader(table.Row{"ENVIRONMENT", "DEFAULT VALUE", "REQUIRED", "DESCRIPTION"})
    // ... populate table with configuration fields
    return tw.Render()
}

// Print YAML format
func (c *configController) PrintYaml(all, nocomments bool) string {
    yamlStr := "---"
    // ... generate YAML configuration
    return yamlStr
}
```

## Plugin Architecture

### Core Interfaces

#### Plugin Interface (Generic)

```go
type Plugin any  // All plugins implement this generic interface
```

#### Plugin Manager Interface

```go
type Plugin interface {
    Provide(constructor ...any)
    Constructors() []any
    Register(plugins []toolkit.Plugin)
    PreStart(ctx context.Context) error
    OnStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}
```

### Plugin Declaration in Proto Files

#### Global Plugins (File-level)

```protobuf
syntax = "proto3";
package myservice;

import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";

// Global plugins available to all services in this file
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};

option (toolkit.plugins) = {
  prefix: "redis"
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
  // Service-specific plugins
  option (toolkit.runtime).plugins = {
    prefix: "cache"
    plugin: "redis"
  };
  
  option (toolkit.runtime).plugins = {
    prefix: "metrics"
    plugin: "prometheus"
  };
  
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]
  };
}
```

### Generated Code Structure

When you define plugins in proto files, the toolkit generates integration code:

```go
// Generated plugin interface definitions
type PgsqlPlugin interface {
    postgres_gorm.Plugin
}

type RedisPlugin interface {
    redis.Plugin
}

type QueuePlugin interface {
    rabbitmq.Plugin
}

// Generated service constructor
func NewMyServiceService(name string, opts ...runtime.Option) (_ toolkit.Service, err error) {
    app := new(serviceMyService)
    
    app.runtime, err = controller.NewRuntime(context.Background(), name, opts...)
    if err != nil {
        return nil, err
    }

    // Plugin initialization
    plugin_pgsql := postgres_gorm.NewPlugin(app.runtime, &postgres_gorm.Options{Name: "pgsql"})
    plugin_redis := redis.NewPlugin(app.runtime, &redis.Options{Name: "redis"})
    plugin_queue := rabbitmq.NewPlugin(app.runtime, &rabbitmq.Options{Name: "queue"})

    // Plugin registration with dependency injection
    app.runtime.Plugin().Provide(func() PgsqlPlugin { return plugin_pgsql })
    app.runtime.Plugin().Provide(func() RedisPlugin { return plugin_redis })
    app.runtime.Plugin().Provide(func() QueuePlugin { return plugin_queue })

    return app.runtime.Service(), nil
}
```

### Plugin Lifecycle Management

#### Lifecycle Hooks

Plugins can implement optional lifecycle methods that are called during service startup/shutdown:

```go
const (
    PluginHookMethodPreStart     = "PreStart"     // Synchronous initialization
    PluginHookMethodOnStart      = "OnStart"      // Asynchronous startup
    PluginHookMethodOnStartSync  = "OnStartSync"  // Synchronous startup
    PluginHookMethodOnStop       = "OnStop"       // Asynchronous shutdown
    PluginHookMethodOnStopSync   = "OnStopSync"   // Synchronous shutdown
)
```

#### Hook Execution Flow

```go
func (c *pluginManager) PreStart(ctx context.Context) error {
    return c.hook(ctx, PluginHookMethodPreStart, true)
}

func (c *pluginManager) OnStart(ctx context.Context) error {
    // Execute synchronous hooks first
    err := c.hook(ctx, PluginHookMethodOnStartSync, true)
    if err != nil {
        return err
    }
    
    // Then execute asynchronous hooks
    return c.hook(ctx, PluginHookMethodOnStart, false)
}

func (c *pluginManager) OnStop(ctx context.Context) error {
    // Execute synchronous stop hooks first
    err := c.hook(ctx, PluginHookMethodOnStopSync, true)
    if err != nil {
        return err
    }
    
    // Then execute asynchronous stop hooks
    return c.hook(ctx, PluginHookMethodOnStop, false)
}
```

#### Reflection-Based Hook Invocation

```go
func (c *pluginManager) call(ctx context.Context, pkg toolkit.Plugin, kind string) error {
    args := []reflect.Value{reflect.ValueOf(ctx)}
    meth := reflect.ValueOf(pkg).MethodByName(kind)
    name := types.Type(pkg)

    if !reflect.ValueOf(meth).IsZero() {
        c.log.V(5).Infof("pluginManager.%s.call: %s", kind, name)
        res := meth.Call(args)

        if len(res) == 1 {
            if v := res[0].Interface(); v != nil {
                return v.(error)
            }
        }
    }
    return nil
}
```

## Available Plugins

### Database Plugins

#### PostgreSQL GORM Plugin

**Import:** `github.com/lastbackend/toolkit-plugins/postgres_gorm`

**Configuration:**
```go
type Options struct {
    DSN              string `env:"DSN" comment:"Database connection string"`
    Host             string `env:"HOST" envDefault:"localhost" comment:"Database host"`
    Port             int    `env:"PORT" envDefault:"5432" comment:"Database port"`
    Username         string `env:"USERNAME" comment:"Database username"`
    Password         string `env:"PASSWORD" comment:"Database password"`
    Database         string `env:"DATABASE" comment:"Database name"`
    SSLMode          string `env:"SSL_MODE" envDefault:"disable" comment:"SSL mode"`
    Timezone         string `env:"TIMEZONE" envDefault:"UTC" comment:"Database timezone"`
    MigrationsDir    string `env:"MIGRATIONS_DIR" comment:"Migrations directory path"`
    Debug            bool   `env:"DEBUG" envDefault:"false" comment:"Enable debug logging"`
}
```

**Interface:**
```go
type Plugin interface {
    DB() *gorm.DB
    Info()
    RunMigration() error
    PreStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}
```

**Usage:**
```go
func NewRepository(app toolkit.Service, db servicepb.PgsqlPlugin) *Repository {
    return &Repository{
        db:  db.DB(),  // *gorm.DB
        log: app.Log(),
    }
}

func (r *Repository) GetUser(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    return &user, err
}
```

#### PostgreSQL go-pg Plugin

**Import:** `github.com/lastbackend/toolkit-plugins/postgres_pg`

**Interface:**
```go
type Plugin interface {
    DB() *pg.DB
    RunMigration() error
    PreStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}
```

### Cache Plugins

#### Redis Plugin

**Import:** `github.com/lastbackend/toolkit-plugins/redis`

**Configuration:**
```go
type Options struct {
    Addr         string `env:"ADDR" comment:"Redis server address"`
    Username     string `env:"USERNAME" comment:"Redis username"`
    Password     string `env:"PASSWORD" comment:"Redis password"`
    Database     int    `env:"DATABASE" envDefault:"0" comment:"Redis database number"`
    DialTimeout  int    `env:"DIAL_TIMEOUT" envDefault:"5" comment:"Connection timeout in seconds"`
    ReadTimeout  int    `env:"READ_TIMEOUT" envDefault:"3" comment:"Read timeout in seconds"`
    WriteTimeout int    `env:"WRITE_TIMEOUT" envDefault:"3" comment:"Write timeout in seconds"`
    PoolSize     int    `env:"POOL_SIZE" envDefault:"10" comment:"Connection pool size"`
    ClusterMode  bool   `env:"CLUSTER_MODE" envDefault:"false" comment:"Enable cluster mode"`
}
```

**Interface:**
```go
type Plugin interface {
    Client() redis.Cmdable
    DB() redis.Cmdable
    ClusterDB() *redis.ClusterClient
    Print()
    PreStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}
```

**Usage:**
```go
func NewCache(app toolkit.Service, redis servicepb.RedisPlugin) *Cache {
    return &Cache{
        client: redis.Client(),
        log:    app.Log(),
    }
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}
```

### Message Queue Plugins

#### RabbitMQ Plugin

**Import:** `github.com/lastbackend/toolkit-plugins/rabbitmq`

**Configuration:**
```go
type Options struct {
    DSN      string `env:"DSN" comment:"RabbitMQ connection string"`
    Host     string `env:"HOST" envDefault:"localhost" comment:"RabbitMQ host"`
    Port     int    `env:"PORT" envDefault:"5672" comment:"RabbitMQ port"`
    Username string `env:"USERNAME" envDefault:"guest" comment:"RabbitMQ username"`
    Password string `env:"PASSWORD" envDefault:"guest" comment:"RabbitMQ password"`
    Vhost    string `env:"VHOST" envDefault:"/" comment:"RabbitMQ virtual host"`
    Prefetch int    `env:"PREFETCH" envDefault:"1" comment:"Prefetch count"`
}
```

**Interface:**
```go
type Plugin interface {
    Publish(event string, data []byte) error
    Subscribe(service, event string, handler func([]byte) error) error
    Channel() *amqp.Channel
    PreStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}
```

**Usage:**
```go
func NewEventBus(app toolkit.Service, queue servicepb.QueuePlugin) *EventBus {
    return &EventBus{
        queue: queue,
        log:   app.Log(),
    }
}

func (e *EventBus) PublishUserCreated(ctx context.Context, user *User) error {
    data, err := json.Marshal(user)
    if err != nil {
        return err
    }
    return e.queue.Publish("user.created", data)
}

func (e *EventBus) SubscribeToUserEvents(handler func([]byte) error) error {
    return e.queue.Subscribe("user-service", "user.created", handler)
}
```

### Real-time Plugins

#### Centrifuge Plugin

**Import:** `github.com/lastbackend/toolkit-plugins/centrifuge`

**Interface:**
```go
type Plugin interface {
    Node() *centrifuge.Node
    Publish(channel string, data []byte) error
    PreStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}
```

### Monitoring Plugins

#### Sentry Plugin

**Import:** `github.com/lastbackend/toolkit-plugins/sentry`

**Configuration:**
```go
type Options struct {
    DSN         string  `env:"DSN" comment:"Sentry DSN"`
    Environment string  `env:"ENVIRONMENT" comment:"Environment name"`
    SampleRate  float64 `env:"SAMPLE_RATE" envDefault:"1.0" comment:"Error sampling rate"`
}
```

## Plugin Development

### Creating a Custom Plugin

#### 1. Plugin Structure

```go
package myplugin

import (
    "context"
    "github.com/lastbackend/toolkit/pkg/runtime"
    "github.com/lastbackend/toolkit/pkg/runtime/logger"
)

// Plugin interface defines the capabilities
type Plugin interface {
    // Plugin-specific methods
    DoSomething() error
    GetClient() *MyClient
    
    // Optional lifecycle hooks
    PreStart(ctx context.Context) error
    OnStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}

// Plugin implementation
type plugin struct {
    runtime runtime.Runtime
    log     logger.Logger
    options *Options
    client  *MyClient
}

// Plugin options/configuration
type Options struct {
    Name     string `json:"name"`
    Host     string `env:"HOST" envDefault:"localhost" comment:"Service host"`
    Port     int    `env:"PORT" envDefault:"8080" comment:"Service port"`
    Username string `env:"USERNAME" comment:"Service username"`
    Password string `env:"PASSWORD" comment:"Service password"`
    Timeout  int    `env:"TIMEOUT" envDefault:"30" comment:"Connection timeout"`
}
```

#### 2. Plugin Constructor

```go
func NewPlugin(runtime runtime.Runtime, opts *Options) Plugin {
    p := &plugin{
        runtime: runtime,
        log:     runtime.Log(),
        options: opts,
    }
    
    // Parse configuration with plugin prefix
    runtime.Config().Parse(&p.options, opts.Name)
    
    return p
}
```

#### 3. Implement Plugin Methods

```go
func (p *plugin) DoSomething() error {
    p.log.Info("Doing something with plugin")
    return p.client.DoOperation()
}

func (p *plugin) GetClient() *MyClient {
    return p.client
}

// Lifecycle hook: Initialize connections
func (p *plugin) PreStart(ctx context.Context) error {
    p.log.Info("Initializing plugin connection")
    
    client, err := NewMyClient(p.options.Host, p.options.Port)
    if err != nil {
        return fmt.Errorf("failed to create client: %w", err)
    }
    
    p.client = client
    return nil
}

// Lifecycle hook: Start background processes
func (p *plugin) OnStart(ctx context.Context) error {
    p.log.Info("Starting plugin background processes")
    go p.backgroundWorker(ctx)
    return nil
}

// Lifecycle hook: Cleanup resources
func (p *plugin) OnStop(ctx context.Context) error {
    p.log.Info("Stopping plugin")
    if p.client != nil {
        return p.client.Close()
    }
    return nil
}

func (p *plugin) backgroundWorker(ctx context.Context) {
    // Background processing
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // Do work
        }
    }
}
```

#### 4. Add Test Support

```go
func NewTestPlugin(opts *Options) Plugin {
    // Create test runtime
    runtime := &testRuntime{log: &testLogger{}}
    
    if opts == nil {
        opts = &Options{
            Name: "test",
            Host: "localhost",
            Port: 8080,
        }
    }
    
    return NewPlugin(runtime, opts)
}

type testRuntime struct {
    log logger.Logger
}

func (t *testRuntime) Log() logger.Logger { return t.log }
func (t *testRuntime) Config() runtime.Config { return &testConfig{} }

type testLogger struct{}
func (t *testLogger) Info(args ...interface{}) {}
func (t *testLogger) Error(args ...interface{}) {}
// ... implement other logger methods
```

### Plugin Integration Example

#### 1. Proto Declaration

```protobuf
option (toolkit.plugins) = {
  prefix: "myplugin"
  plugin: "myplugin"
};

service MyService {
  option (toolkit.runtime) = {
    servers: [GRPC, HTTP]
  };
}
```

#### 2. Generated Integration

The toolkit will generate:

```go
// Plugin interface
type MypluginPlugin interface {
    myplugin.Plugin
}

// Service initialization
func NewMyServiceService(name string, opts ...runtime.Option) (_ toolkit.Service, err error) {
    // ... service setup
    
    plugin_myplugin := myplugin.NewPlugin(app.runtime, &myplugin.Options{Name: "myplugin"})
    app.runtime.Plugin().Provide(func() MypluginPlugin { return plugin_myplugin })
    
    return app.runtime.Service(), nil
}
```

#### 3. Usage in Application

```go
func NewController(
    app toolkit.Service,
    myPlugin servicepb.MypluginPlugin,
) *Controller {
    return &Controller{
        app:      app,
        myPlugin: myPlugin,
    }
}

func (c *Controller) ProcessRequest(ctx context.Context) error {
    return c.myPlugin.DoSomething()
}
```

## Integration Patterns

### Multi-Instance Plugin Pattern

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

Environment variables:
```bash
MYSERVICE_PRIMARY_DB_HOST=primary.db.com
MYSERVICE_PRIMARY_DB_DATABASE=app_db

MYSERVICE_ANALYTICS_DB_HOST=analytics.db.com
MYSERVICE_ANALYTICS_DB_DATABASE=analytics_db

MYSERVICE_SESSION_CACHE_HOST=session.redis.com
MYSERVICE_SESSION_CACHE_DATABASE=0

MYSERVICE_DATA_CACHE_HOST=data.redis.com
MYSERVICE_DATA_CACHE_DATABASE=1
```

### Plugin Composition Pattern

Combine multiple plugins in a single component:

```go
type DataService struct {
    app       toolkit.Service
    primary   servicepb.PrimaryDbPlugin
    analytics servicepb.AnalyticsDbPlugin
    cache     servicepb.DataCachePlugin
    queue     servicepb.QueuePlugin
}

func NewDataService(
    app toolkit.Service,
    primary servicepb.PrimaryDbPlugin,
    analytics servicepb.AnalyticsDbPlugin,
    cache servicepb.DataCachePlugin,
    queue servicepb.QueuePlugin,
) *DataService {
    return &DataService{
        app:       app,
        primary:   primary,
        analytics: analytics,
        cache:     cache,
        queue:     queue,
    }
}

func (d *DataService) CreateUser(ctx context.Context, user *User) error {
    // Save to primary database
    if err := d.primary.DB().WithContext(ctx).Create(user).Error; err != nil {
        return err
    }
    
    // Cache user data
    userData, _ := json.Marshal(user)
    d.cache.Client().Set(ctx, fmt.Sprintf("user:%s", user.ID), userData, time.Hour)
    
    // Send analytics event
    event := map[string]interface{}{
        "event": "user_created",
        "user_id": user.ID,
        "timestamp": time.Now(),
    }
    eventData, _ := json.Marshal(event)
    d.analytics.DB().WithContext(ctx).Create(&AnalyticsEvent{
        Type: "user_created",
        Data: eventData,
    })
    
    // Publish event
    d.queue.Publish("user.created", userData)
    
    return nil
}
```

### Plugin Factory Pattern

Create plugin factories for dynamic plugin creation:

```go
type PluginFactory struct {
    runtime runtime.Runtime
}

func NewPluginFactory(runtime runtime.Runtime) *PluginFactory {
    return &PluginFactory{runtime: runtime}
}

func (f *PluginFactory) CreateDatabasePlugin(name, pluginType string) (interface{}, error) {
    switch pluginType {
    case "postgres_gorm":
        return postgres_gorm.NewPlugin(f.runtime, &postgres_gorm.Options{Name: name}), nil
    case "postgres_pg":
        return postgres_pg.NewPlugin(f.runtime, &postgres_pg.Options{Name: name}), nil
    default:
        return nil, fmt.Errorf("unknown database plugin type: %s", pluginType)
    }
}
```

## Best Practices

### 1. Plugin Configuration

#### Use Descriptive Prefixes
```protobuf
// Good: Descriptive prefixes
option (toolkit.plugins) = { prefix: "user_db", plugin: "postgres_gorm" };
option (toolkit.plugins) = { prefix: "session_cache", plugin: "redis" };

// Avoid: Generic prefixes
option (toolkit.plugins) = { prefix: "db1", plugin: "postgres_gorm" };
option (toolkit.plugins) = { prefix: "cache1", plugin: "redis" };
```

#### Environment Variable Naming
```bash
# Good: Clear hierarchy
MYSERVICE_USER_DB_HOST=localhost
MYSERVICE_USER_DB_PORT=5432
MYSERVICE_SESSION_CACHE_HOST=localhost
MYSERVICE_SESSION_CACHE_PORT=6379

# Bad: Unclear structure
MYSERVICE_DB1_HOST=localhost
MYSERVICE_CACHE1_HOST=localhost
```

### 2. Error Handling

#### Plugin Initialization Errors
```go
func (p *plugin) PreStart(ctx context.Context) error {
    client, err := NewMyClient(p.options.Host, p.options.Port)
    if err != nil {
        return fmt.Errorf("failed to initialize %s plugin: %w", p.options.Name, err)
    }
    
    // Test connection
    if err := client.Ping(ctx); err != nil {
        return fmt.Errorf("failed to connect to %s: %w", p.options.Host, err)
    }
    
    p.client = client
    return nil
}
```

#### Graceful Degradation
```go
func (c *Controller) GetUser(ctx context.Context, id string) (*User, error) {
    // Try cache first
    if cached, err := c.cache.Get(ctx, fmt.Sprintf("user:%s", id)); err == nil {
        var user User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            return &user, nil
        }
    }
    
    // Fallback to database
    var user User
    if err := c.db.DB().WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
        return nil, err
    }
    
    // Update cache (best effort)
    if data, err := json.Marshal(user); err == nil {
        c.cache.Set(ctx, fmt.Sprintf("user:%s", id), data, time.Hour)
    }
    
    return &user, nil
}
```

### 3. Resource Management

#### Connection Pooling
```go
type Options struct {
    MaxConnections     int           `env:"MAX_CONNECTIONS" envDefault:"10"`
    MaxIdleConnections int           `env:"MAX_IDLE_CONNECTIONS" envDefault:"5"`
    ConnMaxLifetime    time.Duration `env:"CONN_MAX_LIFETIME" envDefault:"1h"`
    ConnMaxIdleTime    time.Duration `env:"CONN_MAX_IDLE_TIME" envDefault:"30m"`
}

func (p *plugin) PreStart(ctx context.Context) error {
    db, err := gorm.Open(postgres.Open(p.buildDSN()), &gorm.Config{})
    if err != nil {
        return err
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    
    sqlDB.SetMaxOpenConns(p.options.MaxConnections)
    sqlDB.SetMaxIdleConns(p.options.MaxIdleConnections)
    sqlDB.SetConnMaxLifetime(p.options.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(p.options.ConnMaxIdleTime)
    
    p.db = db
    return nil
}
```

#### Cleanup in OnStop
```go
func (p *plugin) OnStop(ctx context.Context) error {
    if p.db != nil {
        sqlDB, err := p.db.DB()
        if err != nil {
            p.log.Error("Failed to get SQL DB for cleanup: %v", err)
            return nil
        }
        
        if err := sqlDB.Close(); err != nil {
            p.log.Error("Failed to close database connection: %v", err)
            return err
        }
    }
    return nil
}
```

### 4. Testing

#### Plugin Mocking
```go
type MockPlugin struct {
    mock.Mock
}

func (m *MockPlugin) DoSomething() error {
    args := m.Called()
    return args.Error(0)
}

func (m *MockPlugin) GetClient() *MyClient {
    args := m.Called()
    return args.Get(0).(*MyClient)
}

// Test with mock
func TestController(t *testing.T) {
    mockPlugin := new(MockPlugin)
    controller := NewController(nil, mockPlugin)
    
    mockPlugin.On("DoSomething").Return(nil)
    
    err := controller.ProcessRequest(context.Background())
    assert.NoError(t, err)
    
    mockPlugin.AssertExpectations(t)
}
```

#### Integration Testing
```go
func TestPluginIntegration(t *testing.T) {
    // Use test plugin
    plugin := myplugin.NewTestPlugin(&myplugin.Options{
        Host: "localhost",
        Port: 8080,
    })
    
    ctx := context.Background()
    
    // Test lifecycle
    err := plugin.PreStart(ctx)
    require.NoError(t, err)
    
    // Test functionality
    err = plugin.DoSomething()
    assert.NoError(t, err)
    
    // Cleanup
    err = plugin.OnStop(ctx)
    assert.NoError(t, err)
}
```

### 5. Documentation

#### Plugin Interface Documentation
```go
// Plugin provides database access using GORM with PostgreSQL.
// It supports automatic migrations, connection pooling, and health checks.
type Plugin interface {
    // DB returns the GORM database instance for executing queries.
    // The returned *gorm.DB is safe for concurrent use.
    DB() *gorm.DB
    
    // RunMigration executes database migrations from the configured directory.
    // It should be called during application startup after PreStart.
    RunMigration() error
    
    // Info prints plugin configuration information to stdout.
    // Useful for debugging and configuration verification.
    Info()
    
    // PreStart initializes the database connection and verifies connectivity.
    // It must be called before any other plugin methods.
    PreStart(ctx context.Context) error
    
    // OnStop gracefully closes the database connection.
    // It should be called during application shutdown.
    OnStop(ctx context.Context) error
}
```

#### Configuration Documentation
```go
// Options configures the PostgreSQL GORM plugin.
type Options struct {
    // Name is the plugin instance identifier for configuration prefix.
    Name string `json:"name"`
    
    // DSN is the complete database connection string.
    // If provided, it overrides individual connection parameters.
    DSN string `env:"DSN" comment:"Complete database connection string"`
    
    // Host is the PostgreSQL server hostname or IP address.
    Host string `env:"HOST" envDefault:"localhost" comment:"PostgreSQL server host"`
    
    // Port is the PostgreSQL server port number.
    Port int `env:"PORT" envDefault:"5432" comment:"PostgreSQL server port"`
    
    // Username for database authentication.
    Username string `env:"USERNAME" comment:"Database username"`
    
    // Password for database authentication.
    Password string `env:"PASSWORD" comment:"Database password"`
    
    // Database name to connect to.
    Database string `env:"DATABASE" comment:"Database name"`
    
    // SSLMode controls SSL connection behavior.
    // Valid values: disable, require, verify-ca, verify-full
    SSLMode string `env:"SSL_MODE" envDefault:"disable" comment:"SSL connection mode"`
    
    // Timezone for timestamp parsing and formatting.
    Timezone string `env:"TIMEZONE" envDefault:"UTC" comment:"Database timezone"`
    
    // MigrationsDir is the directory containing SQL migration files.
    MigrationsDir string `env:"MIGRATIONS_DIR" comment:"Directory path for migrations"`
    
    // Debug enables GORM debug logging for SQL queries.
    Debug bool `env:"DEBUG" envDefault:"false" comment:"Enable SQL query logging"`
}
```

## Troubleshooting

### Common Issues

#### 1. Plugin Not Found
```
Error: plugin "postgres_gorm" not found
```

**Solution:** Ensure the plugin is imported in your go.mod:
```go
require (
    github.com/lastbackend/toolkit-plugins/postgres_gorm v0.0.0-20240114174800-797efec18f22
)
```

#### 2. Configuration Not Loaded
```
Error: required environment variable not set: MYSERVICE_PGSQL_USERNAME
```

**Solutions:**
- Check environment variable naming: `{SERVICE_PREFIX}_{PLUGIN_PREFIX}_{FIELD_NAME}`
- Verify the service prefix matches the runtime configuration
- Ensure the plugin prefix matches the proto declaration

#### 3. Dependency Injection Failure
```
Error: failed to provide PgsqlPlugin: no constructor registered
```

**Solution:** Verify the generated code includes plugin registration:
```go
app.runtime.Plugin().Provide(func() PgsqlPlugin { return plugin_pgsql })
```

#### 4. Plugin Lifecycle Errors
```
Error: plugin PreStart failed: connection refused
```

**Solutions:**
- Check plugin configuration (host, port, credentials)
- Verify external service is running and accessible
- Check network connectivity and firewall rules
- Review plugin logs for detailed error information

### Debugging Tools

#### Configuration Display
```go
// Print all configuration options
app.Config().PrintTable(true, false)

// Print plugin-specific configuration
plugin.Info()
```

#### Plugin Status
```go
// Check plugin registration
constructors := app.Plugin().Constructors()
fmt.Printf("Registered plugins: %d\n", len(constructors))

// Monitor plugin lifecycle
app.RegisterOnStartHook(func(ctx context.Context) error {
    app.Log().Info("All plugins started successfully")
    return nil
})
```

#### Health Checks
Implement health checks for your plugins:
```go
func (p *plugin) HealthCheck(ctx context.Context) error {
    if p.db == nil {
        return fmt.Errorf("database not initialized")
    }
    
    sqlDB, err := p.db.DB()
    if err != nil {
        return fmt.Errorf("failed to get SQL DB: %w", err)
    }
    
    if err := sqlDB.PingContext(ctx); err != nil {
        return fmt.Errorf("database ping failed: %w", err)
    }
    
    return nil
}
```

This comprehensive guide covers all aspects of the LastBackend Toolkit plugin system, from basic usage to advanced development patterns. The system provides a powerful foundation for building modular, maintainable microservices with standardized external service integrations.