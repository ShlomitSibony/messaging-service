#!/bin/bash

set -e

echo "Starting the application..."
echo "Environment: ${ENV:-development}"

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Build the application
echo "Building the application..."
go build -o messaging-service cmd/server/main.go

# Start the application
echo "Starting messaging service..."
./messaging-service 