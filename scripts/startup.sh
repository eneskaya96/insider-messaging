#!/bin/bash

set -e

echo "Starting Insider Messaging System..."

# Wait for PostgreSQL
echo "Waiting for PostgreSQL..."
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - executing migrations"

# Run migrations
go run cmd/migrate/main.go

# Optionally run seed if flag is set
if [ "$RUN_SEED" = "true" ]; then
  echo "Running database seed..."
  go run cmd/seed/main.go
fi

# Start the application
echo "Starting application..."
exec ./main
