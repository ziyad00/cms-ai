#!/bin/bash
set -e

echo "Building Go application..."
go build -o server

echo "Ensuring Python files are in correct location..."
echo "Source directory contents:"
ls -la tools/renderer/
echo "Creating target directory..."
mkdir -p /app/tools/renderer
echo "Copying files from tools/renderer/* to /app/tools/renderer/..."
cp -r tools/renderer/* /app/tools/renderer/
echo "Setting permissions..."
chmod +x /app/tools/renderer/render_pptx.py

echo "Python files copied:"
ls -la /app/tools/renderer/

echo "Checking render_pptx.py content (first 10 lines):"
head -10 /app/tools/renderer/render_pptx.py

echo "Build complete"