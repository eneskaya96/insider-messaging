# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod ./

# Install golang-migrate CLI and swag
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy source code
COPY . .

# Download dependencies and tidy after copying all source files
RUN go mod download
RUN go mod tidy

# Generate Swagger documentation
RUN swag init -g cmd/api/main.go -o ./docs

# Build the application, migration tool, and seed tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate-tool cmd/migrate/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o seed-tool cmd/seed/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata postgresql-client

WORKDIR /root/

# Copy the binaries from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrate-tool .
COPY --from=builder /app/seed-tool .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/scripts/startup.sh .

RUN chmod +x startup.sh

# Expose port
EXPOSE 8080

# Run startup script
CMD ["./startup.sh"]
