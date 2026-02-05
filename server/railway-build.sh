#!/bin/bash
set -e

echo "Building Go application..."
go build -o server

echo "Ensuring Python files are in correct location..."
mkdir -p /app/tools/renderer
cp -r tools/renderer/* /app/tools/renderer/

echo "Python files copied:"
ls -la /app/tools/renderer/

echo "Build complete"