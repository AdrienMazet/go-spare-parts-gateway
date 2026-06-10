# Multi-stage build for Go services
# Usage:
#   docker build -t spare-parts-api:local --build-arg SERVICE=. .
#   docker build -t offer-price-worker:local --build-arg SERVICE=./cmd/offer-price-worker .
#   docker build -t spare-parts-provider:local --build-arg SERVICE=./cmd/provider .

# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy go mod and sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy entire source
COPY . .

# Build argument for service path
ARG SERVICE=.

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/app ${SERVICE}

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/app /app/app
COPY --from=builder /build/internal/db/migrations /app/internal/db/migrations

ENTRYPOINT ["/app/app"]
