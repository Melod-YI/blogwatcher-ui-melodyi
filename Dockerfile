# Build stage
FROM golang:1.25.6-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary (CGO disabled - modernc.org/sqlite is pure Go)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o blogwatcher-ui ./cmd/server

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

# Create non-root user for security
RUN adduser -D -u 1000 appuser

# Copy binary from builder
COPY --from=builder /build/blogwatcher-ui /app/

# Create data directory for SQLite database
RUN mkdir -p /home/appuser/.blogwatcher && chown -R appuser:appuser /home/appuser

USER appuser

EXPOSE 8080

ENV PORT=8080

ENTRYPOINT ["/app/blogwatcher-ui"]
