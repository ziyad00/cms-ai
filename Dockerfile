# Build Go server
FROM golang:1.23 AS go-build
WORKDIR /app
COPY server/ .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Build Next.js web
FROM node:18 AS web-build
WORKDIR /app
COPY web/ .
RUN npm install
RUN npm run build

# Final image with Node.js runtime
FROM node:18-alpine
RUN apk --no-cache add ca-certificates python3 py3-pip
RUN pip3 install --break-system-packages python-pptx httpx asyncio

# Copy Go server
COPY --from=go-build /app/server /usr/local/bin/server
# Copy SQL migrations for runtime
COPY --from=go-build /app/migrations /app/server/migrations

# Copy Next.js app with node_modules
WORKDIR /app/web
COPY --from=web-build /app /app/web
RUN npm ci --production

# Copy tools
COPY tools/ /app/tools/
COPY scripts/start.sh /app/scripts/start.sh
RUN chmod +x /app/scripts/start.sh
# Make Python renderer script executable
RUN chmod +x /app/tools/renderer/render_pptx.py
# Verify files are copied correctly
RUN ls -la /app/tools/renderer/ && echo "Files copied successfully"

WORKDIR /app
EXPOSE 3000 8080
CMD ["/app/scripts/start.sh"]
