#!/bin/sh

set -e

echo "Starting Insider Messaging System..."

# Wait for PostgreSQL
echo "Waiting for PostgreSQL..."
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - executing migrations"

# Run migrations using golang-migrate CLI
./migrate-tool -cmd up -path migrations

# Run seed data automatically on first startup
echo "Seeding database with test data..."
./seed-tool || echo "Seed failed or already exists (this is normal on restart)"

# Start the application
echo "Starting application..."
exec ./main
