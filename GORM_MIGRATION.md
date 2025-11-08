# GORM Migration Guide

This document explains the migration from raw SQL to GORM ORM with clean architecture principles.

## What Changed

### 1. Dependencies Added
- `gorm.io/gorm` - GORM ORM core
- `gorm.io/driver/postgres` - PostgreSQL driver for GORM
- `gorm.io/plugin/optimisticlock` - Optimistic locking plugin
- `github.com/golang-migrate/migrate/v4` - Professional migration tool

### 2. Architecture Pattern

**Clean Architecture Maintained:**
```
Domain Layer (Pure Go)
    ↓
Repository Interface
    ↓
Infrastructure Layer (GORM Models + Mappers)
```

### 3. Key Files Changed

#### New Files:
- `internal/infrastructure/persistence/model/message_model.go` - GORM model
- `internal/infrastructure/persistence/model/mapper.go` - Domain ↔ Model conversion
- `internal/infrastructure/persistence/postgres_gorm.go` - GORM connection setup
- `internal/infrastructure/persistence/message_repository_gorm.go` - GORM repository
- `internal/infrastructure/persistence/gorm_errors.go` - Error mapping helper
- `migrations/000001_create_messages_table.down.sql` - Rollback migration

#### Modified Files:
- `cmd/migrate/main.go` - Now uses golang-migrate CLI
- `cmd/api/main.go` - Uses GORM connection
- `internal/presentation/handler/health_handler.go` - GORM health check
- `Dockerfile` - Includes migrate binary
- `Makefile` - New migration commands
- `README.md` - Updated documentation

### 4. Hybrid Approach Explained

**Simple Queries (GORM ORM):**
```go
func (r *repo) Create(ctx context.Context, message *entity.Message) error {
    model := mapper.ToModel(message)
    return r.db.WithContext(ctx).Create(model).Error
}
```

**Critical Queries (Raw SQL for SKIP LOCKED):**
```go
func (r *repo) FindPendingMessages(ctx context.Context, limit int) {
    query := `SELECT * FROM messages WHERE status = ?
              ORDER BY created_at ASC LIMIT ?
              FOR UPDATE SKIP LOCKED`
    r.db.Raw(query, "pending", limit).Scan(&models)
}
```

### 5. Optimistic Locking

Uses GORM optimisticlock plugin:
```go
type MessageModel struct {
    Version optimisticlock.Version `gorm:"column:version"`
}
```

Automatically increments version and checks on updates.

### 6. Transaction Handling

**GORM Transaction API:**
```go
err := r.db.Transaction(func(tx *gorm.DB) error {
    // All operations in transaction
    return nil  // Auto-commit
})
```

### 7. Migration Commands

```bash
# Run migrations up
make migrate-up

# Rollback last migration
make migrate-down

# Check current version
make migrate-version

# Create new migration
make migrate-create
```

## Benefits

✅ **Type Safety**: GORM models provide compile-time type checking
✅ **SQL Injection Protected**: Parameterized queries by default
✅ **Clean Architecture**: Domain layer still ORM-independent
✅ **Professional Migrations**: Version control with rollback support
✅ **Optimistic Locking**: Built-in with plugin
✅ **Performance**: Critical queries still use raw SQL
✅ **Error Handling**: Standardized error mapping

## Testing

Domain layer tests remain unchanged because domain entities are ORM-independent.

Repository tests now use GORM test database.

## Running the Project

1. **Start services:**
   ```bash
   docker-compose up -d
   ```

2. **Check migration status:**
   ```bash
   docker-compose exec app ./migrate-tool -cmd version -path migrations
   ```

3. **Seed data:**
   ```bash
   docker-compose exec app go run cmd/seed/main.go
   ```

4. **Verify:**
   ```bash
   curl http://localhost:8080/health
   ```

## Rollback Plan

If issues arise, the migration can be rolled back:
```bash
make migrate-down
```

Old raw SQL repository code is preserved in `message_repository_postgres.go` (can be reactivated if needed).
