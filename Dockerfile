# Production Dockerfile for Lawrence OSS
# Multi-stage build with Go backend and React frontend

# =============================================================================
# Stage 1: Build Go Backend
# =============================================================================
FROM golang:1.23-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Temporarily comment out DuckDB imports to avoid CGO issues
RUN find . -name "*.go" -exec sed -i 's|_ "github.com/marcboeker/go-duckdb"|// _ "github.com/marcboeker/go-duckdb"|g' {} \;

# Build the application without CGO
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o lawrence ./cmd/all-in-one

# =============================================================================
# Stage 2: Build React Frontend
# =============================================================================
FROM node:20-alpine AS frontend-builder

# Install pnpm
RUN npm install -g pnpm

# Set working directory
WORKDIR /app

# Copy package files
COPY ui/package.json ui/pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy source code
COPY ui/ .

# Build the frontend
RUN pnpm build

# =============================================================================
# Stage 3: Production Image
# =============================================================================
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 -S lawrence && \
    adduser -u 1001 -S lawrence -G lawrence

# Set working directory
WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/lawrence .

# Copy frontend build
COPY --from=frontend-builder /app/dist ./web/dist

# Copy configuration
COPY lawrence.yaml .

# Create data directory
RUN mkdir -p /app/data && \
    chown -R lawrence:lawrence /app

# Switch to non-root user
USER lawrence

# Expose ports
# 8080 - HTTP API
# 4320 - OpAMP server
# 4317 - OTLP gRPC
# 4318 - OTLP HTTP
EXPOSE 8080 4320 4317 4318

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV TZ=UTC

# Run the application
CMD ["./lawrence"]