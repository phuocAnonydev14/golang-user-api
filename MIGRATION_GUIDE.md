# Database Migration Guide

This project supports two ways to run database migrations:

## Method 1: Manual Migration Command (Recommended for Production)

Use the dedicated migration command for better control:

```bash
# Run all pending migrations
go run cmd/migrate/main.go -up

# View help
go run cmd/migrate/main.go -help

# Rollback (not implemented yet)
go run cmd/migrate/main.go -down
```

## Method 2: Auto-Migration (Convenient for Development)

Set environment variable to enable auto-migration when the app starts:

```bash
# Enable auto-migration
export AUTO_MIGRATE=true
go run cmd/main.go

# Or inline
AUTO_MIGRATE=true go run cmd/main.go
```

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string (required)
- `AUTO_MIGRATE` - Enable auto-migration on app startup (optional)
  - Values: `true`, `1`, `yes` (case insensitive)
  - Default: disabled

## Examples

### Development Workflow
```bash
# Set up environment
export DATABASE_URL="postgres://user:password@localhost:5432/userdb?sslmode=disable"

# Run migrations manually
go run cmd/migrate/main.go -up

# Start the application
go run cmd/main.go
```

### Production Workflow
```bash
# Run migrations separately before deployment
go run cmd/migrate/main.go -up

# Deploy application without auto-migration
go run cmd/main.go
```

### Quick Development (Auto-Migration)
```bash
# Set environment and start with auto-migration
export DATABASE_URL="postgres://user:password@localhost:5432/userdb?sslmode=disable"
export AUTO_MIGRATE=true
go run cmd/main.go
```

## Best Practices

1. **Production**: Always use manual migration command
2. **Development**: Either approach works, auto-migration is convenient
3. **CI/CD**: Run migrations as a separate deployment step
4. **Rollbacks**: Plan rollback strategies before applying migrations

## Migration Files

- Location: `migrations/` directory
- Format: `XXX_description.sql` (e.g., `001_create_users_table.sql`)
- Execution: Alphabetical order
- Tracking: Stored in `migrations` table