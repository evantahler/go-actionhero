# Task List: Go ActionHero Backend Implementation

## Phase 1: Foundation & Core Framework

### Project Setup
- [x] Create project directory structure (cmd/, internal/, actions/, migrations/)
- [x] Initialize go.mod with Go 1.21+ and add initial dependencies

### Core Types & Interfaces
- [x] Define Action interface (Name, Description, Inputs, Middleware, Web, Task, Run)
- [x] Define Connection struct with Type, Identifier, ID, Session, Subscriptions
- [x] Define Server interface (Name, Initialize, Start, Stop)
- [x] Define Middleware interface with RunBefore and RunAfter hooks

### Configuration & Logging
- [x] Implement configuration system using viper with env var overrides and YAML/JSON support
- [x] Create config structs (database, redis, logger, server, session, tasks)
- [x] Implement structured logger using logrus/zap with colors and log levels (debug, info, warn, error, fatal)

### Core Framework
- [x] Create main API singleton struct with initialization lifecycle
- [x] Implement Initializer interface and system for plugin-like architecture
- [x] Create TypedError system with error codes/types and HTTP status mapping

## Phase 2: HTTP & WebSocket Server

### HTTP Server
- [ ] Implement basic HTTP server using net/http or gin/fiber
- [ ] Implement regex-based route matching with path parameter extraction
- [ ] Support RESTful HTTP methods (GET, POST, PUT, DELETE, PATCH, OPTIONS)
- [ ] Implement request parsing (JSON body, form data, query params, path params)
- [ ] Add CORS support with configurable origins, methods, and headers
- [ ] Add optional static file serving support

### WebSocket Server
- [ ] Implement WebSocket upgrade handling from HTTP server
- [ ] Handle WebSocket messages (action execution, subscribe/unsubscribe)
- [ ] Implement WebSocket message broadcasting to subscribed connections

## Phase 3: Action System

### Action Registration & Discovery
- [ ] Implement action auto-discovery and registry system
- [ ] Support action metadata storage (name, description, web config, task config)

### Input Validation
- [ ] Implement input validation using go-playground/validator with struct tags
- [ ] Add secret field marking and redaction in logs
- [ ] Implement type coercion and transformation for input params

### Middleware System
- [ ] Implement middleware chain execution (before/after hooks)
- [ ] Support middleware modifying params and responses

### Action Execution
- [ ] Implement action execution flow (find action, load session, validate, middleware, run, return)

## Phase 4: Database & ORM

### ORM Setup
- [ ] Choose and integrate ORM (GORM recommended for start)

### Schema Definition
- [ ] Define User schema/model matching TypeScript version
- [ ] Define Message schema/model matching TypeScript version
- [ ] Set up migration system using golang-migrate or GORM migrations
- [ ] Add auto-migration option for development

### Database Operations
- [ ] Implement database operations (CRUD, query builders, transactions)

## Phase 5: Redis & Session Management

### Redis Client
- [ ] Set up Redis client using go-redis/redis with connection pooling

### Session Management
- [ ] Implement Redis-backed session storage with TTL support
- [ ] Implement cookie-based session ID management
- [ ] Create session middleware for authentication

### Pub/Sub
- [ ] Set up Redis pub/sub for channel subscriptions
- [ ] Implement broadcast messages to channels
- [ ] Integrate pub/sub with WebSocket for real-time updates

## Phase 6: Background Jobs

### Job Queue Setup
- [ ] Set up job queue using Asynq (Redis-based)

### Task System
- [ ] Implement wrapping actions as jobs
- [ ] Implement worker pools for job processing
- [ ] Add job retries and error handling
- [ ] Support task enqueueing from actions

### Recurrent Tasks
- [ ] Implement recurrent task scheduling with frequency property
- [ ] Add scheduler for periodic tasks (cron-like)
- [ ] Prevent duplicate task executions

## Phase 7: CLI Support

### CLI Framework
- [ ] Set up CLI framework using cobra
- [ ] Auto-generate CLI commands from actions

### CLI Actions
- [ ] Support action execution via CLI with flags
- [ ] Implement list all actions command
- [ ] Support different output formats (JSON, table)

## Phase 8: Testing & Validation

### Testing Strategy
- [ ] Write unit tests for core components
- [ ] Write integration tests for actions
- [ ] Write HTTP server tests using httptest
- [ ] Write WebSocket tests

### Test Utilities
- [ ] Create test helpers for action execution and mock connections

## Phase 9: Error Handling

### Typed Errors
- [ ] Implement typed error system matching TypeScript implementation
- [ ] Define error codes/types and HTTP status code mapping

### Error Responses
- [ ] Implement consistent error response format (JSON)
- [ ] Add error logging with stack traces

## Phase 10: Additional Features

### Swagger/OpenAPI
- [ ] Generate OpenAPI spec from actions using swaggo/swag
- [ ] Add Swagger UI endpoint

### Health Checks
- [ ] Implement health check endpoints (status, database, redis)

### Graceful Shutdown
- [ ] Implement graceful shutdown with signal handling (SIGTERM, SIGINT)
- [ ] Add connection cleanup and job queue cleanup on shutdown

## Porting Existing Actions

- [ ] Port existing user actions from TypeScript (user:create, user:edit, user:view)
- [ ] Port existing message actions from TypeScript (message:create, messages:list, messages:cleanup, messages:hello)
- [ ] Port existing session actions from TypeScript
- [ ] Port existing status actions from TypeScript
- [ ] Port existing swagger actions from TypeScript
- [ ] Port existing files actions from TypeScript

## Documentation & Polish

- [ ] Create comprehensive README.md with setup instructions and usage examples
- [ ] Conduct performance testing and optimization

---

**Total Tasks: 70**
**Completed: 13** (Phase 1 complete!)

**Estimated Timeline:** 10 weeks (as outlined in PLAN.md)

