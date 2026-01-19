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

# Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates python3 py3-pip
RUN pip3 install --break-system-packages python-pptx
COPY --from=go-build /app/server /usr/local/bin/server
COPY --from=web-build /app/.next /app/web/.next
COPY --from=web-build /app/public /app/web/public
COPY --from=web-build /app/package.json /app/web/
COPY --from=web-build /app/next.config.mjs /app/web/
COPY tools/renderer/ /app/tools/renderer/
EXPOSE 8080
CMD ["/usr/local/bin/server"]
