# Agent Best Practices & Guidelines

This document outlines the best practices and guidelines for developing the Go ActionHero backend. Follow these principles consistently throughout the project.

**Core Principles**: TDD, DRY, and Simplicity (KISS) are the foundation of our development approach.

## Core Development Principles

### 1. Test-Driven Development (TDD)
**Always write tests first, then implement the feature.**

- Write failing tests before writing implementation code
- Write the minimum code necessary to make tests pass
- Refactor while keeping tests green
- Aim for high test coverage (>80% for core components)
- Use table-driven tests for multiple scenarios
- Test both success and error cases

**Example:**
```go
// 1. Write test first
func TestUserCreate_ValidatesInput(t *testing.T) {
    // Test implementation
}

// 2. Write implementation to make test pass
func (a *UserCreate) Run(...) {
    // Implementation
}
```

### 2. DRY (Don't Repeat Yourself)
**Avoid code duplication. Extract common patterns into reusable functions/types.**

- Identify repeated patterns early
- Create helper functions, utilities, or shared types
- Use composition over duplication
- Refactor when you notice duplication (rule of three: refactor on third occurrence)

**Example:**
```go
// Bad: Repeated validation logic
func validateUserEmail(email string) error { ... }
func validateMessageEmail(email string) error { ... }

// Good: Shared validation
func validateEmail(email string) error { ... }
```

### 3. Simplicity (KISS - Keep It Simple, Stupid)
**Always choose the simplest solution that works. Complexity should be justified.**

- Prefer simple, straightforward code over clever solutions
- Avoid over-engineering and premature abstraction
- Use standard library when possible instead of external dependencies
- Write code that is easy to read and understand
- When in doubt, choose the simpler approach
- Complexity should solve a real problem, not anticipate future needs

**Example:**
```go
// Good: Simple and clear
func GetUser(id int) (*User, error) {
    var user User
    err := db.First(&user, id).Error
    return &user, err
}

// Bad: Over-engineered with unnecessary abstraction
func GetUser(id int) (*User, error) {
    return userRepositoryFactory().getUserService().getUserById(id)
}
```

**Questions to ask:**
- Is this the simplest way to solve this problem?
- Will future developers understand this easily?
- Do we really need this abstraction right now?
- Can we use standard library instead?

### 4. SOLID Principles
- **Single Responsibility**: Each function/type should have one reason to change
- **Open/Closed**: Open for extension, closed for modification
- **Liskov Substitution**: Subtypes must be substitutable for their base types
- **Interface Segregation**: Many specific interfaces are better than one general interface
- **Dependency Inversion**: Depend on abstractions, not concretions

### 4. Go Idioms & Best Practices

#### Error Handling
- Always check errors explicitly
- Use `fmt.Errorf` with `%w` for error wrapping (Go 1.13+)
- Return errors, don't panic (except in truly unrecoverable situations)
- Provide context in error messages

```go
// Good
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Bad
if err != nil {
    panic(err)
}
```

#### Naming Conventions
- Use camelCase for unexported names
- Use PascalCase for exported names
- Keep names short but descriptive
- Use verbs for functions, nouns for types
- Avoid abbreviations unless widely understood

```go
// Good
type User struct { ... }
func CreateUser(...) { ... }
func (u *User) Validate() error { ... }

// Bad
type usr struct { ... }
func create_usr(...) { ... }
```

#### Interfaces
- Keep interfaces small (prefer 1-3 methods)
- Define interfaces where they're used, not where they're implemented
- Use interfaces for abstraction and testability

```go
// Good: Small, focused interface
type Validator interface {
    Validate() error
}

// Bad: Large interface with many responsibilities
type UserManager interface {
    Create() error
    Update() error
    Delete() error
    Validate() error
    SendEmail() error
    // ... many more
}
```

#### Context Usage
- Always accept `context.Context` as the first parameter in functions that do I/O
- Use context for cancellation, timeouts, and request-scoped values
- Don't store contexts in structs; pass them explicitly

```go
// Good
func (a *UserCreate) Run(ctx context.Context, params interface{}, conn *Connection) (interface{}, error) {
    // Use ctx for cancellation/timeouts
}

// Bad
type UserCreate struct {
    ctx context.Context // Don't store context
}
```

#### Concurrency
- Use channels for communication between goroutines
- Prefer channels over shared memory
- Use `sync` package primitives (Mutex, WaitGroup) when needed
- Always clean up goroutines (avoid leaks)

```go
// Good: Use channels
ch := make(chan Result)
go func() {
    ch <- doWork()
}()

// Bad: Shared memory without synchronization
var result Result
go func() {
    result = doWork() // Race condition!
}()
```

## Code Quality Standards

### 5. Code Organization
- Follow the project structure defined in PLAN.md
- Keep related code together
- Separate concerns (business logic, data access, HTTP handling)
- Use internal packages to prevent external access

### 6. Documentation
- Write godoc comments for all exported functions, types, and packages
- Include examples in documentation where helpful
- Keep comments up-to-date with code changes
- Use comments to explain "why", not "what"

```go
// UserCreate handles user creation requests.
// It validates input, checks for duplicate emails, and creates the user.
type UserCreate struct{}

// Run executes the user creation action.
// Returns the created user or an error if validation fails.
func (a *UserCreate) Run(ctx context.Context, params interface{}, conn *Connection) (interface{}, error) {
    // Implementation
}
```

### 7. Performance Considerations
- Profile before optimizing
- Use `go test -bench` for benchmarking
- Avoid premature optimization
- Consider memory allocations in hot paths
- Use `sync.Pool` for frequently allocated objects

### 8. Security Best Practices
- Never log sensitive data (passwords, tokens, secrets)
- Use parameterized queries (prevent SQL injection)
- Validate and sanitize all inputs
- Use secure session management
- Implement proper CORS policies
- Use HTTPS in production

```go
// Good: Secret fields are redacted in logs
type UserCreateInputs struct {
    Password string `validate:"required,min=8" secret:"true"`
}

// Bad: Secrets logged in plain text
logger.Info("Creating user with password:", password)
```

## Testing Standards

### 9. Test Structure
- Use `testify/assert` for assertions
- Use table-driven tests for multiple test cases
- Use subtests for organizing related tests
- Keep tests independent and isolated
- Use test fixtures for complex setup

```go
func TestUserCreate_Run(t *testing.T) {
    tests := []struct {
        name    string
        input   *UserCreateInputs
        wantErr bool
    }{
        {
            name: "valid input",
            input: &UserCreateInputs{
                Name:     "John Doe",
                Email:    "john@example.com",
                Password: "password123",
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            input: &UserCreateInputs{
                Name:     "John Doe",
                Email:    "invalid-email",
                Password: "password123",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 10. Test Coverage
- Aim for >80% coverage on core components
- Test error paths, not just happy paths
- Test edge cases and boundary conditions
- Use integration tests for full workflows
- Mock external dependencies (databases, Redis, etc.)

## Git & Version Control

### 11. Commit Messages
- Use clear, descriptive commit messages
- Follow conventional commits format when possible
- Keep commits focused (one logical change per commit)
- Reference issue numbers when applicable

```
feat: add user creation action
fix: resolve session middleware authentication issue
test: add tests for message actions
refactor: extract common validation logic
```

### 12. Code Review
- Review your own code before requesting review
- Keep PRs focused and reasonably sized
- Address review feedback promptly
- Use PRs for discussion and learning

## Project-Specific Guidelines

### 13. Action Development
- Follow the Action interface pattern
- Use struct tags for validation
- Implement proper error handling
- Support all transport types (HTTP, WebSocket, CLI, Tasks)
- Document action inputs and outputs

### 14. Middleware Development
- Keep middleware focused and reusable
- Support both RunBefore and RunAfter hooks
- Return errors to halt execution
- Support modifying params and responses

### 15. Database Operations
- Use transactions for multi-step operations
- Handle database errors appropriately
- Use prepared statements for repeated queries
- Implement proper connection pooling

### 16. Configuration Management
- Use environment variables for sensitive data
- Provide sensible defaults
- Support multiple environments (dev, test, prod)
- Validate configuration on startup

## Continuous Improvement

### 17. Refactoring
- Refactor when you see duplication (rule of three)
- Refactor when tests are hard to write
- Refactor when code is hard to understand
- Refactor to simplify complex code
- Keep refactoring incremental and safe (tests protect you)
- Simplify before adding new features

### 18. Learning & Growth
- Read Go best practices and idioms
- Review standard library code
- Learn from code reviews
- Stay updated with Go releases and features

## Checklist for Every Feature

Before considering a feature complete:

- [ ] Tests written first (TDD)
- [ ] All tests passing
- [ ] Code follows DRY principles
- [ ] Code is as simple as possible (KISS)
- [ ] Error handling implemented
- [ ] Documentation added (godoc)
- [ ] No code duplication
- [ ] Follows Go idioms
- [ ] Security considerations addressed
- [ ] Performance acceptable (or profiled)
- [ ] Code reviewed (self-review minimum)

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Testing Best Practices](https://github.com/golang/go/wiki/TestComments)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

---

**Remember**: These practices are guidelines, not strict rules. Use judgment and adapt as needed for the specific situation, but default to these practices unless there's a good reason to deviate.

