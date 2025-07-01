# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server ./cmd/server

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable and migrations
COPY --from=builder /build/server /app/server
COPY --from=builder /build/migrations /app/migrations
COPY --from=builder /build/config.yaml /app/config.yaml

# Use an unprivileged user
USER appuser

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/server"]