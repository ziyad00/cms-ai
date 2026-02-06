# Use Go image for building
FROM golang:1.21-bullseye AS builder

WORKDIR /app
COPY . .

# Build Go application
RUN go mod download
RUN go build -o server

# Final stage with Python and runtime
FROM ubuntu:22.04

# Install Python and dependencies
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Create app directory
WORKDIR /app

# Copy Go binary
COPY --from=builder /app/server /app/server

# Copy Python requirements and install
COPY requirements.txt /app/requirements.txt
RUN pip3 install -r requirements.txt

# Copy olama Python renderer files
COPY tools/renderer/ /app/tools/renderer/
RUN chmod +x /app/tools/renderer/render_pptx.py

# Debug: Verify files are copied
RUN echo "=== DOCKERFILE: Checking copied files ===" && \
    ls -la /app/tools/renderer/ && \
    echo "=== DOCKERFILE: Checking render_pptx.py content ===" && \
    head -10 /app/tools/renderer/render_pptx.py

EXPOSE 8080
CMD ["./server"]