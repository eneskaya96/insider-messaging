# Insider Messaging System

An automatic message sending system built with Go that processes and sends messages via webhook at configurable intervals using a custom scheduler implementation.

## Features

- **Custom Scheduler**: Native Go implementation using goroutines and channels (no cron packages)
- **GORM ORM**: Type-safe database operations with clean architecture
- **Professional Migrations**: golang-migrate/migrate for version control and rollbacks
- **Batch Processing**: Processes messages in configurable batch sizes with worker pool pattern
- **FIFO Queue**: Messages processed in order of creation with database-level locking
- **Atomic Operations**: Transaction-based processing with optimistic locking to prevent race conditions
- **Hybrid Approach**: GORM for simple queries, raw SQL for critical operations (SKIP LOCKED)
- **Redis Caching**: Caches successfully sent messages with metadata
- **Rate Limiting**: Built-in rate limiting for webhook calls
- **Error Handling**: Comprehensive error handling with retry logic
- **Health Checks**: Liveness and readiness endpoints for container orchestration
- **API Documentation**: Auto-generated Swagger/OpenAPI documentation
- **Clean Architecture**: DDD principles with clear separation of concerns (Domain independent from ORM)

## Architecture

```
├── cmd/
│   ├── api/              # Application entry point
│   ├── migrate/          # Database migration tool
│   └── seed/             # Database seeding tool
├── internal/
│   ├── domain/           # Business logic layer
│   │   ├── entity/       # Domain entities
│   │   ├── valueobject/  # Value objects with validation
│   │   └── repository/   # Repository interfaces
│   ├── application/      # Application services
│   │   ├── service/      # Use case implementations
│   │   └── dto/          # Data transfer objects
│   ├── infrastructure/   # External dependencies
│   │   ├── persistence/  # GORM + PostgreSQL implementation
│   │   │   └── model/    # Database models (separate from domain)
│   │   ├── cache/        # Redis implementation
│   │   ├── http/         # HTTP client for webhooks
│   │   └── scheduler/    # Custom message scheduler
│   └── presentation/     # API layer
│       ├── handler/      # HTTP handlers
│       ├── middleware/   # HTTP middleware
│       └── router/       # Route definitions
├── pkg/                  # Shared packages
│   ├── config/          # Configuration management
│   ├── logger/          # Structured logging
│   └── errors/          # Custom error types
└── migrations/          # SQL migration files
```

## Requirements

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 16
- Redis 7

## Quick Start

### 1. Clone the repository

```bash
git clone <repository-url>
cd insider-messaging
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your webhook URL and other configurations
```

### 3. Start services with Docker Compose

```bash
docker-compose up -d
```

This command will:
- Start PostgreSQL database
- Start Redis cache
- Build and run the application
- Run migrations automatically
- Start the scheduler

### 4. Seed the database

```bash
# Inside the container
docker-compose exec app ./main seed

# Or locally
make seed
```

### 5. Access the application

- **API**: http://localhost:8080
- **Swagger Documentation**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health

## Configuration

All configuration is managed through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | postgres |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | Database user | messaging_user |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | messaging_db |
| `REDIS_HOST` | Redis host | redis |
| `REDIS_PORT` | Redis port | 6379 |
| `APP_PORT` | Application port | 8080 |
| `MESSAGE_BATCH_SIZE` | Messages per cycle | 2 |
| `MESSAGE_INTERVAL_MINUTES` | Processing interval | 2 |
| `MESSAGE_CHAR_LIMIT` | Max message length | 160 |
| `MESSAGE_WORKER_COUNT` | Worker goroutines | 5 |
| `WEBHOOK_URL` | Webhook endpoint | - |
| `WEBHOOK_AUTH_KEY` | Auth key header | - |
| `SEED_MESSAGE_COUNT` | Test messages to create | 100 |

## API Endpoints

### Scheduler Management

- `POST /api/v1/scheduler/start` - Start automatic message sending
- `POST /api/v1/scheduler/stop` - Stop automatic message sending
- `GET /api/v1/scheduler/status` - Get scheduler status and statistics

### Message Management

- `GET /api/v1/messages/sent` - List sent messages (paginated)
- `GET /api/v1/messages/:id` - Get message details
- `GET /api/v1/messages/stats` - Get message statistics
- `POST /api/v1/messages` - Create a new message

### Health & Monitoring

- `GET /health` - Application health check
- `GET /ready` - Readiness probe
- `GET /live` - Liveness probe

## Scheduler Implementation

The scheduler uses a **custom Go implementation** without any cron packages:

```go
// Key features:
- time.Ticker for interval-based triggering
- Worker pool pattern with configurable workers
- Graceful shutdown with context cancellation
- SELECT FOR UPDATE SKIP LOCKED for atomic message selection
- Optimistic locking to prevent double-sending
```

### Processing Flow

1. Every N minutes, scheduler triggers a processing cycle
2. Fetches batch of pending messages using SKIP LOCKED
3. Distributes messages to worker pool
4. Each worker:
   - Marks message as processing
   - Sends via webhook with rate limiting
   - Updates status (sent/failed)
   - Caches to Redis on success
5. Failed messages retry up to MAX_RETRIES

## Database Schema & Migrations

### GORM + golang-migrate Approach

This project uses a **hybrid approach**:
- **GORM**: For type-safe database operations and model definitions
- **golang-migrate**: For professional migration version control
- **Separate Models**: Infrastructure models are separate from domain entities (Clean Architecture)

### Migration Management

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

### Schema

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    sent_at TIMESTAMP,
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    last_error TEXT,
    error_code VARCHAR(50),
    webhook_message_id VARCHAR(255),
    webhook_response TEXT,
    version BIGINT DEFAULT 0  -- Optimistic locking with GORM plugin
);

-- Indexes for FIFO and efficient querying
CREATE INDEX idx_messages_pending_fifo ON messages(created_at)
    WHERE status = 'pending';
```

### Clean Architecture with GORM

```go
// Domain Entity (No GORM tags - Pure business logic)
type Message struct {
    id          uuid.UUID
    phoneNumber *valueobject.PhoneNumber
    // ... domain logic
}

// Infrastructure Model (GORM tags)
type MessageModel struct {
    ID          uuid.UUID `gorm:"primaryKey"`
    PhoneNumber string    `gorm:"column:phone_number"`
    Version     optimisticlock.Version
    // ... GORM-specific tags
}

// Mapper converts between Domain and Infrastructure
func ToEntity(model *MessageModel) *entity.Message
func ToModel(entity *entity.Message) *MessageModel
```

## Development

### Run locally

```bash
# Install dependencies
go mod download

# Run migrations
make migrate-up

# Seed database
make seed

# Run application
make run
```

### Run tests

```bash
# All tests
make test

# With coverage
make test-cover
```

### Generate Swagger docs

```bash
make swagger
```

## Docker Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down

# Rebuild and restart
docker-compose up -d --build

# Run migration inside container
docker-compose exec app go run cmd/migrate/main.go

# Run seed inside container
docker-compose exec app go run cmd/seed/main.go
```

## Error Handling

The system handles various error scenarios:

| Error Type | Behavior |
|------------|----------|
| Network timeout | Retry with exponential backoff |
| Rate limit | Respect webhook rate limits |
| Invalid response | Mark as failed, log details |
| Database lock | Skip locked rows, process available |
| Concurrent updates | Optimistic locking prevents conflicts |

## Monitoring & Observability

- **Structured Logging**: JSON logs with zap
- **Health Endpoints**: Database and Redis connectivity checks
- **Metrics**: Processing statistics via status endpoint
- **Error Tracking**: Detailed error codes and messages

## Production Considerations

1. **Database Connection Pooling**: Configured via `DB_MAX_OPEN_CONNS`
2. **Graceful Shutdown**: Waits for in-flight requests
3. **Rate Limiting**: Prevents webhook overload
4. **Optimistic Locking**: Prevents race conditions
5. **Redis Caching**: Optional, degrades gracefully if unavailable

## Testing

Comprehensive test coverage including:

- Unit tests for domain entities and value objects
- Integration tests for repositories
- E2E tests for API endpoints
- Scheduler concurrency tests

## License

MIT License

## Support

For issues and questions, please open an issue on GitHub.
