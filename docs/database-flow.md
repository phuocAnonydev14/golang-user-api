# Database Connection and Interaction Flow in Go

## Overview
This document explains how to connect and interact with PostgreSQL database in Go using the pgx driver.

## 1. Database Connection Flow

### Step 1: Initialize Connection Pool
```go
// pkg/db/postgres.go
func InitPostgres() error {
    // Get database URL from environment
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        return fmt.Errorf("DATABASE_URL environment variable is not set")
    }

    // Create connection pool with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Create pool - this is the main connection manager
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
    }
    
    // Store globally accessible pool
    DB = pool
    return nil
}
```

**Key Points:**
- Uses connection pooling for efficiency
- Sets connection timeout
- Stores pool globally for app-wide access

### Step 2: Database Schema Setup (Migrations)
```go
// pkg/db/migrate.go
func RunMigrations() error {
    // 1. Create migrations tracking table
    createMigrationsTable := `
        CREATE TABLE IF NOT EXISTS migrations (
            id SERIAL PRIMARY KEY,
            filename VARCHAR(255) NOT NULL UNIQUE,
            executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
    `
    
    // 2. Execute table creation
    if _, err := DB.Exec(context.Background(), createMigrationsTable); err != nil {
        return fmt.Errorf("failed to create migrations table: %w", err)
    }

    // 3. Read migration files from directory
    files, err := os.ReadDir("migrations")
    
    // 4. Check which migrations already executed
    executedMigrations := make(map[string]bool)
    rows, err := DB.Query(context.Background(), "SELECT filename FROM migrations")
    
    // 5. Execute only new migrations
    for _, filename := range sqlFiles {
        if !executedMigrations[filename] {
            // Read and execute migration
            content, _ := os.ReadFile(filepath.Join("migrations", filename))
            DB.Exec(context.Background(), string(content))
            
            // Record as executed
            DB.Exec(context.Background(), 
                "INSERT INTO migrations (filename) VALUES ($1)", filename)
        }
    }
}
```

## 2. Repository Pattern for Database Interaction

### Repository Structure
```go
// internal/user/repository.go
type Repository struct {
    db *pgxpool.Pool  // Database connection pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
    return &Repository{db: db}
}
```

### CRUD Operations Flow

#### Create Operation
```go
func (r *Repository) Create(user CreateUserRequest) (*UserResponse, error) {
    // 1. Generate unique ID
    id := uuid.NewString()
    
    // 2. Prepare SQL query with placeholders ($1, $2, etc.)
    query := `
        INSERT INTO users (id, username, email, age) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id, username, email, age, created_at
    `
    
    // 3. Execute query and scan result into struct
    var response UserResponse
    var createdAt interface{}
    
    err := r.db.QueryRow(context.Background(), query, 
        id, user.Username, user.Email, user.Age).
        Scan(&response.ID, &response.Username, &response.Email, &response.Age, &createdAt)
    
    // 4. Handle errors and return
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return &response, nil
}
```

#### Read Operation (Single)
```go
func (r *Repository) GetByID(id string) (*UserResponse, error) {
    // 1. Prepare SELECT query
    query := `SELECT id, username, email, age FROM users WHERE id = $1`
    
    // 2. Execute query and scan into struct
    var user UserResponse
    err := r.db.QueryRow(context.Background(), query, id).
        Scan(&user.ID, &user.Username, &user.Email, &user.Age)
    
    // 3. Handle not found or other errors
    if err != nil {
        return nil, fmt.Errorf("failed to get user by ID: %w", err)
    }
    
    return &user, nil
}
```

#### Read Operation (Multiple)
```go
func (r *Repository) GetAll() ([]UserResponse, error) {
    // 1. Prepare query for multiple rows
    query := `SELECT id, username, email, age FROM users ORDER BY created_at DESC`
    
    // 2. Execute query - returns rows iterator
    rows, err := r.db.Query(context.Background(), query)
    if err != nil {
        return nil, fmt.Errorf("failed to get users: %w", err)
    }
    defer rows.Close()  // Always close rows!
    
    // 3. Iterate through rows and collect results
    var users []UserResponse
    for rows.Next() {
        var user UserResponse
        if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Age); err != nil {
            return nil, fmt.Errorf("failed to scan user: %w", err)
        }
        users = append(users, user)
    }
    
    return users, nil
}
```

#### Update Operation
```go
func (r *Repository) Update(id string, user CreateUserRequest) (*UserResponse, error) {
    // 1. Prepare UPDATE query with RETURNING clause
    query := `
        UPDATE users 
        SET username = $2, email = $3, age = $4, updated_at = NOW()
        WHERE id = $1 
        RETURNING id, username, email, age
    `
    
    // 2. Execute and scan updated row
    var response UserResponse
    err := r.db.QueryRow(context.Background(), query, 
        id, user.Username, user.Email, user.Age).
        Scan(&response.ID, &response.Username, &response.Email, &response.Age)
    
    if err != nil {
        return nil, fmt.Errorf("failed to update user: %w", err)
    }
    
    return &response, nil
}
```

#### Delete Operation
```go
func (r *Repository) Delete(id string) error {
    // 1. Prepare DELETE query
    query := `DELETE FROM users WHERE id = $1`
    
    // 2. Execute and check affected rows
    result, err := r.db.Exec(context.Background(), query, id)
    if err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }
    
    // 3. Verify deletion happened
    rowsAffected := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("user not found")
    }
    
    return nil
}
```

## 3. Handler Layer Integration

### Handler Structure
```go
// internal/user/handler.go
type Handler struct {
    repo *Repository  // Dependency injection
}

func NewHandler(repo *Repository) *Handler {
    return &Handler{repo: repo}
}
```

### HTTP Handler Flow
```go
func (h *Handler) CreateUser(c echo.Context) error {
    // 1. Parse and validate request
    var req CreateUserRequest
    if err := decodeStrictJSON(c, &req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body"})
    }

    if err := validate.Struct(req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": err.Error()})
    }

    // 2. Call repository layer
    user, err := h.repo.Create(req)
    if err != nil {
        // Log actual error, return generic message
        log.Printf("Failed to create user: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to create user"})
    }

    // 3. Return success response
    return c.JSON(http.StatusCreated, user)
}
```

## 4. Application Startup Flow

```go
// cmd/main.go
func main() {
    // 1. Initialize database connection
    if err := db.InitPostgres(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // 2. Run database migrations
    if err := db.RunMigrations(); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }

    // 3. Seed initial data (optional)
    if err := db.SeedDatabase(); err != nil {
        log.Printf("Warning: Failed to seed database: %v", err)
    }

    // 4. Initialize application layers
    userRepo := user.NewRepository(db.DB)      // Repository layer
    userHandler := user.NewHandler(userRepo)   // Handler layer

    // 5. Setup HTTP server and routes
    e := echo.New()
    httpxecho.RegisterRoutes(e, userHandler)

    // 6. Start server
    if err := e.Start(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

## 5. Key Database Concepts

### Connection Pooling
- **Purpose**: Reuse database connections for efficiency
- **pgxpool.Pool**: Manages multiple connections automatically
- **Benefits**: Better performance, resource management

### Context Usage
- **Purpose**: Control timeouts, cancellation
- **Pattern**: Always pass `context.Background()` or timeout context
- **Example**: `ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)`

### Parameter Placeholders
- **Pattern**: Use `$1, $2, $3...` instead of string concatenation
- **Security**: Prevents SQL injection attacks
- **Example**: `SELECT * FROM users WHERE id = $1` with `query(ctx, sql, userID)`

### Error Handling
- **Wrap errors**: Use `fmt.Errorf("context: %w", err)` for error wrapping
- **Layer separation**: Repository returns detailed errors, handlers return generic ones
- **Logging**: Log detailed errors server-side, return safe messages to clients

### Transaction Pattern (Advanced)
```go
func (r *Repository) CreateUserWithProfile(user User, profile Profile) error {
    // Begin transaction
    tx, err := r.db.Begin(context.Background())
    if err != nil {
        return err
    }
    defer tx.Rollback() // Rollback if not committed

    // Execute multiple operations
    _, err = tx.Exec(ctx, "INSERT INTO users...", user.ID, user.Name)
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, "INSERT INTO profiles...", profile.UserID, profile.Bio)
    if err != nil {
        return err
    }

    // Commit transaction
    return tx.Commit(context.Background())
}
```

## 6. Best Practices

### ✅ Do's
- Use connection pooling
- Always use parameterized queries
- Handle errors at appropriate layers
- Use context for timeouts
- Close rows when iterating
- Use transactions for multi-operation consistency

### ❌ Don'ts
- Don't use string concatenation for SQL
- Don't ignore database errors
- Don't forget to close rows/connections
- Don't expose internal errors to clients
- Don't use global database connections without pooling

## 7. Environment Setup

```env
# .env file
DATABASE_URL=postgres://username:password@localhost:5432/dbname?sslmode=disable
```

```bash
# Start PostgreSQL locally
docker run --name postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=userapi -p 5432:5432 -d postgres:15

# Set environment variable
export DATABASE_URL="postgres://postgres:password@localhost:5432/userapi?sslmode=disable"

# Run application
go run cmd/main.go
```

This flow ensures robust, secure, and maintainable database interactions in Go applications.