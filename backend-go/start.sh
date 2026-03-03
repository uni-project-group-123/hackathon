#!/bin/bash
set -e

PORT=${PORT:-5000}

cd "$(dirname "$0")"

if [ ! -f "./sentinel" ]; then
    echo "Binary not found. Running build.sh first..."
    ./build.sh
fi

echo "Starting Sentinel on port $PORT..."

exec ./sentinel
