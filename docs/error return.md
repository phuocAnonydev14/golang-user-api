# Go Error Handling Best Practices

## 1. Basic Error Return Pattern

```go
func doSomething() error {
    if somethingWrong {
        return fmt.Errorf("something went wrong: %v", details)
    }
    return nil
}

// Usage
if err := doSomething(); err != nil {
    // handle error
    return err
}
```

## 2. Error Wrapping (Go 1.13+)

```go
import "fmt"

func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("failed to open file %s: %w", filename, err)
    }
    defer file.Close()
    
    // ... process file
    return nil
}
```

## 3. Custom Error Types

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

func validateUser(user User) error {
    if user.Email == "" {
        return ValidationError{
            Field:   "email",
            Message: "email is required",
        }
    }
    return nil
}
```

## 4. Returning Multiple Values

```go
func parseConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    return &config, nil
}
```

## 5. Error Checking Patterns

### Early Return Pattern
```go
func createUser(req CreateUserRequest) (*User, error) {
    // Validate input
    if err := validateRequest(req); err != nil {
        return nil, err
    }
    
    // Check if user exists
    exists, err := userExists(req.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to check user existence: %w", err)
    }
    if exists {
        return nil, fmt.Errorf("user already exists")
    }
    
    // Create user
    user := &User{
        ID:       generateID(),
        Email:    req.Email,
        Username: req.Username,
    }
    
    return user, nil
}
```

## 6. HTTP Handler Error Patterns

```go
func CreateUserHandler(c echo.Context) error {
    var req CreateUserRequest
    
    // Decode request
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request format",
        })
    }
    
    // Validate request
    if err := validate.Struct(req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": err.Error(),
        })
    }
    
    // Create user
    user, err := createUser(req)
    if err != nil {
        // Log the actual error
        log.Printf("Failed to create user: %v", err)
        
        // Return generic error to client
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to create user",
        })
    }
    
    return c.JSON(http.StatusCreated, user)
}
```

## 7. Error Sentinel Values

```go
var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidEmail = errors.New("invalid email format")
)

func findUser(email string) (*User, error) {
    if email == "" {
        return nil, ErrInvalidEmail
    }
    
    // ... search logic
    
    return nil, ErrUserNotFound
}

// Usage with errors.Is()
user, err := findUser(email)
if err != nil {
    if errors.Is(err, ErrUserNotFound) {
        return c.JSON(http.StatusNotFound, map[string]string{
            "error": "User not found",
        })
    }
    return c.JSON(http.StatusBadRequest, map[string]string{
        "error": err.Error(),
    })
}
```

## 8. Don'ts

❌ **Don't ignore errors**
```go
// BAD
data, _ := os.ReadFile("config.json")
```

❌ **Don't panic for normal errors**
```go
// BAD
if err != nil {
    panic(err)
}
```

❌ **Don't return both value and error**
```go
// BAD
func getUser() (*User, error) {
    user := &User{ID: "123"}
    return user, fmt.Errorf("some error")
}
```

## 9. Do's

✅ **Always check errors**
```go
data, err := os.ReadFile("config.json")
if err != nil {
    return fmt.Errorf("failed to read config: %w", err)
}
```

✅ **Provide context in error messages**
```go
return fmt.Errorf("failed to create user %s: %w", username, err)
```

✅ **Use error wrapping to preserve original error**
```go
return fmt.Errorf("database operation failed: %w", err)
```

✅ **Return nil for success cases**
```go
func doSomething() error {
    // ... successful operation
    return nil
}
```

## 10. Error Handling in Different Layers

### Service Layer
```go
func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    if err := s.validateUser(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    user, err := s.repo.Create(req)
    if err != nil {
        return nil, fmt.Errorf("failed to create user in database: %w", err)
    }
    
    return user, nil
}
```

### Repository Layer
```go
func (r *UserRepository) Create(user User) error {
    query := "INSERT INTO users (id, email, username) VALUES (?, ?, ?)"
    _, err := r.db.Exec(query, user.ID, user.Email, user.Username)
    if err != nil {
        return fmt.Errorf("failed to execute insert query: %w", err)
    }
    return nil
}
```

Remember: In Go, errors are values. Handle them explicitly and provide meaningful context!