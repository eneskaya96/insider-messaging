#!/bin/bash

set -e

echo "Running tests..."

# Run unit tests
echo "Running unit tests..."
go test -v -race ./internal/domain/...

# Run integration tests if available
if [ -d "tests" ]; then
    echo "Running integration tests..."
    go test -v -race ./tests/...
fi

echo "All tests passed!"
