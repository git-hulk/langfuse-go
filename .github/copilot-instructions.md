
When generating and reviewing Go code, you MUST follow these guidelines:

### File Structure
- Always include proper package declarations at the top
- Add necessary imports in the correct order (standard library, third-party, local)
- Use `goimports` formatting automatically

### Function and Method Generation
- Generate complete functions with proper error handling
- Include context parameters for functions that might perform I/O operations
- Add proper documentation comments for exported functions

```go
// ProcessUserData processes user data and returns the result.
// It returns an error if the data is invalid or processing fails.
func ProcessUserData(ctx context.Context, userData *UserData) (*Result, error) {
    if userData == nil {
        return nil, fmt.Errorf("user data cannot be nil")
    }

    // Implementation here

    return result, nil
}
```

### Struct Generation
- Generate structs with proper field tags for JSON, YAML, etc.
- Include validation tags when appropriate
- Use pointer types for optional fields

```go
type User struct {
    ID       int64     `json:"id" db:"id"`
    Name     string    `json:"name" db:"name" validate:"required"`
    Email    string    `json:"email" db:"email" validate:"required,email"`
    Age      *int      `json:"age,omitempty" db:"age"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

- Use `any` instead of `interface{}` for generic types in Go 1.18+


### Interface Generation
- Keep interfaces small and focused (Interface Segregation Principle)
- Use descriptive names ending with "-er" when appropriate
- Place interfaces close to their usage, not their implementation

```go
type UserProcessor interface {
    ProcessUser(ctx context.Context, user *User) error
}

type UserRepository interface {
    GetUser(ctx context.Context, id int64) (*User, error)
    SaveUser(ctx context.Context, user *User) error
}
```

## Error Handling Patterns

### Always Generate Error Handling
- Never ignore errors; always handle them explicitly
- Use error wrapping to provide context
- Start error messages with lowercase letters

```go
result, err := someOperation()
if err != nil {
    return nil, fmt.Errorf("failed to perform operation: %w", err)
}
```

### Custom Error Types
- Create custom error types for domain-specific errors
- Implement the `error` interface properly

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Message)
}
```

## Testing Code Generation

### Test Function Structure
- Use `github.com/stretchr/testify` for assertions
- Follow the AAA pattern (Arrange, Act, Assert)
- Use table-driven tests for multiple test cases

```go
func TestProcessUser(t *testing.T) {
    tests := []struct {
        name     string
        input    *User
        expected *Result
        wantErr  bool
    }{
        {
            name:     "valid user",
            input:    &User{Name: "John", Email: "john@example.com"},
            expected: &Result{Success: true},
            wantErr:  false,
        },
        {
            name:    "nil user",
            input:   nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ProcessUser(context.Background(), tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

- Use httptest for HTTP handler tests

## Naming Conventions for Code Generation

### Boolean Options
- Use `Enabled` suffix for feature toggles: `CacheEnabled`, `LoggingEnabled`
- Use `Is` prefix for state checks: `IsActive`, `IsValid`

### Constants
- Use ALL_CAPS with underscores for package-level constants
- Group related constants in blocks

```go
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
    
    StatusActive   = "active"
    StatusInactive = "inactive"
    StatusPending  = "pending"
)
```

### Package Names
- Use short, descriptive package names
- Avoid stuttering (don't repeat package name in type names)
- Use singular nouns

