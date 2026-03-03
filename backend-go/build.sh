#!/bin/bash
set -e

echo "Building Sentinel..."

cd "$(dirname "$0")"

if [ ! -f "go.mod" ]; then
    echo "Error: go.mod not found. Are you in the backend-go directory?"
    exit 1
fi

echo "Downloading dependencies..."
go mod download

echo "Building binary..."
CGO_ENABLED=0 go build -o sentinel .

echo "Build complete: ./sentinel"
