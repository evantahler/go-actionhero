# Plan: Recreating bun-actionhero Backend in Go

## Overview

This document outlines the plan to recreate the TypeScript/Bun ActionHero backend in Go. The goal is to maintain feature parity while leveraging Go's strengths and idiomatic patterns.

## Architecture Analysis

### Key Components Identified

1. **Core Framework**
   - Action system (transport-agnostic)
   - Connection management (HTTP, WebSocket, CLI, Resque)
   - Middleware system
   - Initializer system (plugin-like architecture)
   - Configuration system with environment overrides

2. **Servers**
   - HTTP/Web server (Bun.serve)
   - WebSocket server (integrated with HTTP)
   - CLI server (Commander.js)

3. **Data Layer**
   - Drizzle ORM → Go equivalent (GORM or Ent)
   - PostgreSQL database
   - Redis for sessions and pub/sub
   - Migration system

4. **Background Jobs**
   - node-resque → Go job queue (Asynq, Machinery, or custom)
   - Task scheduling (cron-like)
   - Worker pools

5. **Session Management**
   - Cookie-based sessions
   - Redis-backed storage
   - Session middleware

6. **Pub/Sub**
   - Redis pub/sub for real-time messaging
   - Channel subscriptions per connection

7. **Validation**
   - Zod schemas → Go struct tags + validation library (go-playground/validator or similar)

## Implementation Plan

### Phase 1: Foundation & Core Framework

#### 1.1 Project Structure
```
go-actionhero/
├── cmd/
│   └── actionhero/          # Main CLI entry point
├── internal/
│   ├── api/                 # Core API framework
│   │   ├── action.go        # Action interface and base
│   │   ├── connection.go    # Connection abstraction
│   │   ├── server.go        # Server interface
│   │   ├── middleware.go    # Middleware system
│   │   └── api.go           # Main API singleton
│   ├── config/              # Configuration system
│   │   ├── config.go        # Main config struct
│   │   ├── database.go
│   │   ├── redis.go
│   │   ├── logger.go
│   │   ├── server.go
│   │   └── session.go
│   ├── initializers/        # Plugin-like initializers
│   │   ├── initializer.go  # Base initializer interface
│   │   ├── actions.go      # Action loader
│   │   ├── connections.go
│   │   ├── db.go           # Database init
│   │   ├── redis.go
│   │   ├── resque.go       # Job queue init
│   │   ├── session.go
│   │   └── servers.go
│   ├── servers/
│   │   ├── web.go          # HTTP server
│   │   └── websocket.go    # WebSocket handling
│   ├── middleware/
│   │   └── session.go      # Session middleware
│   ├── schema/             # Database schemas
│   │   ├── users.go
│   │   └── messages.go
│   ├── ops/               # Business logic operations
│   │   ├── user_ops.go
│   │   └── message_ops.go
│   └── util/
│       ├── logger.go
│       ├── errors.go      # Typed errors
│       └── validation.go  # Input validation helpers
├── actions/               # User-defined actions
│   ├── user.go
│   ├── message.go
│   ├── session.go
│   └── status.go
├── migrations/            # Database migrations
├── go.mod
├── go.sum
└── README.md
```

#### 1.2 Core Types & Interfaces

**Action Interface:**
```go
type Action interface {
    Name() string
    Description() string
    Inputs() interface{} // Schema/struct for validation
    Middleware() []Middleware
    Web() *WebConfig
    Task() *TaskConfig
    Run(ctx context.Context, params interface{}, conn *Connection) (interface{}, error)
}
```

**Connection:**
```go
type Connection struct {
    Type         string
    Identifier   string
    ID           string
    Session      *Session
    Subscriptions map[string]bool
    RawConnection interface{} // For WebSocket, etc.
}
```

**Server Interface:**
```go
type Server interface {
    Name() string
    Initialize() error
    Start() error
    Stop() error
}
```

#### 1.3 Configuration System
- Use `viper` or similar for config management
- Support environment variable overrides
- YAML/JSON config files
- Per-environment configs (dev, test, production)

#### 1.4 Logger
- Structured logging with `logrus` or `zap`
- Support for colors (when appropriate)
- Log levels: debug, info, warn, error, fatal

### Phase 2: HTTP & WebSocket Server

#### 2.1 HTTP Server
- Use `net/http` or `gin`/`fiber` for routing
- Support for:
  - RESTful routes (GET, POST, PUT, DELETE, PATCH, OPTIONS)
  - Path parameters (`/user/:id`)
  - Query parameters
  - JSON body parsing
  - Form data parsing
  - Static file serving (optional)
  - CORS support

#### 2.2 WebSocket Server
- Use `gorilla/websocket` or `nhooyr.io/websocket`
- Handle WebSocket upgrades from HTTP
- Support action execution via WebSocket
- Support subscribe/unsubscribe messages
- Broadcast messages to subscribed connections

#### 2.3 Route Matching
- Implement regex-based route matching
- Support path parameters extraction
- Match HTTP method + route pattern

### Phase 3: Action System

#### 3.1 Action Registration
- Auto-discovery of actions (similar to glob loader)
- Action registry/map
- Support for action metadata (name, description, etc.)

#### 3.2 Input Validation
- Replace Zod with struct tags + `go-playground/validator`
- Custom validation rules
- Secret field marking (for logging redaction)
- Type coercion and transformation

#### 3.3 Middleware System
- Before/after middleware hooks
- Middleware chain execution
- Ability to modify params and responses
- Error handling in middleware

#### 3.4 Action Execution Flow
```
1. Find action by name
2. Load session (if not loaded)
3. Parse and validate inputs
4. Run before middleware
5. Execute action.run()
6. Run after middleware
7. Return response/error
```

### Phase 4: Database & ORM

#### 4.1 ORM Choice
**Option A: GORM** (Most popular, feature-rich)
- Pros: Mature, migrations, associations, hooks
- Cons: Can be slower, reflection-heavy

**Option B: Ent** (Facebook's codegen ORM)
- Pros: Type-safe, performant, great for complex schemas
- Cons: Code generation step, newer

**Option C: sqlc** (SQL-first)
- Pros: Type-safe, fast, no runtime overhead
- Cons: More manual SQL writing

**Recommendation: Start with GORM for rapid development, consider Ent for production**

#### 4.2 Schema Definition
- Define models matching TypeScript schemas
- Support for migrations (use `golang-migrate` or GORM migrations)
- Auto-migration option for development

#### 4.3 Database Operations
- CRUD operations
- Query builders
- Transactions support

### Phase 5: Redis & Session Management

#### 5.1 Redis Client
- Use `go-redis/redis` or `gomodule/redigo`
- Connection pooling
- Pub/sub support

#### 5.2 Session Management
- Cookie-based session IDs
- Redis storage for session data
- TTL support
- Session middleware for authentication

#### 5.3 Pub/Sub
- Channel subscriptions per connection
- Broadcast messages to channels
- WebSocket integration for real-time updates

### Phase 6: Background Jobs

#### 6.1 Job Queue Choice
**Option A: Asynq** (Redis-based, similar to Resque)
- Pros: Simple API, Redis-backed, good docs
- Cons: Less features than Machinery

**Option B: Machinery** (More feature-rich)
- Pros: Multiple brokers, retries, workflows
- Cons: More complex

**Option C: Custom implementation** (Using Redis directly)
- Pros: Full control, matches Resque pattern
- Cons: More work to implement

**Recommendation: Start with Asynq for simplicity**

#### 6.2 Task System
- Wrap actions as jobs
- Support for scheduled tasks (cron-like)
- Worker pools
- Job retries and error handling
- Task enqueueing from actions

#### 6.3 Recurrent Tasks
- Support for `frequency` property on actions
- Scheduler for periodic tasks
- Prevent duplicate executions

### Phase 7: CLI Support

#### 7.1 CLI Framework
- Use `cobra` or `urfave/cli`
- Auto-generate CLI commands from actions
- Support for action execution via CLI
- Help text generation

#### 7.2 CLI Actions
- List all actions
- Execute actions with flags
- Support for different output formats (JSON, table)

### Phase 8: Testing & Validation

#### 8.1 Testing Strategy
- Unit tests for core components
- Integration tests for actions
- HTTP server tests (use `httptest`)
- WebSocket tests
- Database tests with test containers or in-memory DB

#### 8.2 Test Utilities
- Test helpers for action execution
- Mock connections
- Test fixtures

### Phase 9: Error Handling

#### 9.1 Typed Errors
- Custom error types matching TypeScript implementation
- Error codes/types
- Stack traces
- HTTP status code mapping

#### 9.2 Error Responses
- Consistent error format
- JSON error responses
- Error logging

### Phase 10: Additional Features

#### 10.1 Swagger/OpenAPI
- Generate OpenAPI spec from actions
- Swagger UI endpoint
- Use `swaggo/swag` or similar

#### 10.2 Health Checks
- Status endpoint
- Database health check
- Redis health check

#### 10.3 Graceful Shutdown
- Signal handling (SIGTERM, SIGINT)
- Graceful server shutdown
- Connection cleanup
- Job queue cleanup

## Technology Stack Recommendations

### Core
- **Go**: 1.21+ (for generics support)
- **HTTP Server**: `net/http` (standard library) or `gin`/`fiber`
- **WebSocket**: `gorilla/websocket` or `nhooyr.io/websocket`
- **CLI**: `cobra` or `urfave/cli`

### Data & Storage
- **ORM**: GORM or Ent
- **Migrations**: `golang-migrate/migrate` or GORM migrations
- **Redis**: `go-redis/redis`
- **PostgreSQL**: `lib/pq` or `pgx`

### Background Jobs
- **Job Queue**: Asynq or Machinery
- **Scheduler**: Built into Asynq or `robfig/cron`

### Validation & Config
- **Validation**: `go-playground/validator`
- **Config**: `spf13/viper`
- **Environment**: `joho/godotenv`

### Logging & Monitoring
- **Logging**: `sirupsen/logrus` or `uber-go/zap`
- **Structured Logging**: JSON output support

### Testing
- **Testing**: Standard `testing` package
- **HTTP Testing**: `net/http/httptest`
- **Assertions**: `testify/assert`
- **Test Containers**: `testcontainers/testcontainers-go` (optional)

## Migration Strategy

### Step-by-Step Approach

1. **Week 1-2: Foundation**
   - Set up project structure
   - Implement core API framework
   - Configuration system
   - Logger

2. **Week 3-4: HTTP Server**
   - Basic HTTP server
   - Route matching
   - Action execution
   - Middleware system

3. **Week 5: Database**
   - Set up ORM
   - Define schemas
   - Migration system
   - Basic CRUD operations

4. **Week 6: Sessions & Redis**
   - Redis client setup
   - Session management
   - Session middleware

5. **Week 7: WebSocket**
   - WebSocket server
   - Message handling
   - Pub/sub integration

6. **Week 8: Background Jobs**
   - Job queue setup
   - Worker implementation
   - Task scheduling

7. **Week 9: CLI**
   - CLI framework
   - Action commands
   - Help system

8. **Week 10: Testing & Polish**
   - Comprehensive tests
   - Error handling improvements
   - Documentation
   - Performance optimization

## Key Differences from TypeScript Version

### Advantages in Go
- **Performance**: Better concurrency, lower memory footprint
- **Type Safety**: Compile-time type checking
- **Deployment**: Single binary, easier deployment
- **Standard Library**: Rich standard library for HTTP, JSON, etc.

### Challenges
- **Generics**: Need Go 1.21+ for better generic support
- **Reflection**: Some dynamic features require reflection
- **Validation**: Struct tags vs Zod schemas (less flexible but more performant)
- **Error Handling**: Explicit error handling vs exceptions

## Action Examples (Go vs TypeScript)

### TypeScript Action:
```typescript
export class UserCreate implements Action {
  name = "user:create";
  inputs = z.object({
    name: z.string().min(3),
    email: z.string().email(),
    password: z.string().min(8).secret(),
  });
  async run(params: ActionParams<UserCreate>) {
    // implementation
  }
}
```

### Go Action:
```go
type UserCreate struct{}

func (a *UserCreate) Name() string { return "user:create" }

func (a *UserCreate) Inputs() interface{} {
    return &UserCreateInputs{}
}

type UserCreateInputs struct {
    Name     string `validate:"required,min=3"`
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8" secret:"true"`
}

func (a *UserCreate) Run(ctx context.Context, params interface{}, conn *Connection) (interface{}, error) {
    inputs := params.(*UserCreateInputs)
    // implementation
}
```

## Open Questions

1. **ORM Choice**: GORM vs Ent vs sqlc?
2. **HTTP Framework**: Standard library vs Gin vs Fiber?
3. **Job Queue**: Asynq vs Machinery vs custom?
4. **WebSocket Library**: gorilla/websocket vs nhooyr.io/websocket?
5. **Validation Library**: go-playground/validator vs custom?
6. **CLI Framework**: Cobra vs urfave/cli?

## Next Steps

1. Review and approve this plan
2. Set up initial project structure
3. Implement Phase 1 (Foundation)
4. Iterate through phases with testing
5. Port existing actions from TypeScript
6. Ensure feature parity
7. Performance testing and optimization

## References

- Original Repository: https://github.com/evantahler/bun-actionhero
- Go Best Practices: https://go.dev/doc/effective_go
- Project Layout: https://github.com/golang-standards/project-layout

