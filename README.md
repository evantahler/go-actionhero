# Go ActionHero

A Go port of the ActionHero framework - a transport-agnostic API framework for building web applications.

## Status

🚧 **Work in Progress** - This is an active port from the TypeScript/Bun version.

## Features

- Transport-agnostic Actions (HTTP, WebSocket, CLI, Background Jobs)
- Built-in session management (Redis-backed)
- Background job queue (Resque-like)
- Middleware system
- Database ORM with migrations
- Pub/Sub for real-time messaging

## Project Structure

```
go-actionhero/
├── cmd/actionhero/     # Main CLI entry point
├── internal/           # Internal packages
│   ├── api/           # Core API framework
│   ├── config/         # Configuration system
│   ├── initializers/   # Plugin-like initializers
│   ├── servers/        # HTTP/WebSocket servers
│   ├── middleware/     # Middleware implementations
│   ├── schema/         # Database schemas
│   ├── ops/            # Business logic operations
│   └── util/           # Utilities
├── actions/            # User-defined actions
└── migrations/         # Database migrations
```

## Development

### Prerequisites

- Go 1.21 or later
- PostgreSQL
- Redis

### Setup

```bash
# Clone the repository
git clone https://github.com/evantahler/go-actionhero.git
cd go-actionhero

# Install dependencies
make install

# Configure environment (optional)
cp .env.example .env
# Edit .env with your settings

# Build the binary
make build

# Run tests
make test

# Run the server
./actionhero start
```

### Available Make Targets

```bash
make help           # Display all available targets
make build          # Build the binary
make clean          # Remove build artifacts
make test           # Run tests
make test-coverage  # Run tests with coverage report
make lint           # Run golangci-lint
make fmt            # Format Go code
make vet            # Run go vet
make install        # Install dependencies
make dev            # Build and run the server
make check          # Run all checks (format, vet, lint, test)
```

### Configuration

Configuration can be provided via:
1. **Default values** - Sensible defaults for all settings
2. **YAML/JSON config files** - `config.yaml` or `config.json` (optional)
3. **Environment-specific config** - `config.dev.yaml`, `config.test.yaml`, etc. (optional)
4. **`.env` files** - `.env`, `.env.local`, `.env.{NODE_ENV}` (optional)
5. **Environment variables** - `ACTIONHERO_*` prefixed variables (highest priority)

All settings use the `ACTIONHERO_` prefix when set as environment variables. For example:
- `ACTIONHERO_SERVER_WEB_PORT=9000`
- `ACTIONHERO_DATABASE_HOST=db.example.com`
- `ACTIONHERO_LOGGER_LEVEL=debug`

See `.env.example` for all available configuration options.

## Progress

### ✅ Completed (Phase 1)
- Project structure
- Core interfaces (Action, Connection, Server, Middleware)
- API singleton with initialization lifecycle
- Initializer interface for plugin-like architecture
- Typed error system
- Configuration system (with .env file support)
- Structured logger (logrus)
- CLI entry point with welcome message
- Makefile for build, test, and lint
- GitHub Actions CI/CD
- Comprehensive tests

### 📋 Planned
- HTTP server
- WebSocket server
- Database integration
- Session management
- Background jobs
- Full CLI command support (action execution, etc.)

## License

MIT

## References

- Original TypeScript/Bun version: https://github.com/evantahler/bun-actionhero

