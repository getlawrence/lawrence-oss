# Production Dockerfile for Lawrence OSS
# Multi-stage build with Go backend and React frontend

# =============================================================================
# Stage 1: Build Go Backend
# =============================================================================
FROM golang:1.24-bookworm AS backend-builder

# Install build dependencies (including gcc/g++ for CGO, SQLite, and DuckDB)
RUN apt-get update && apt-get install -y \
    git \
    ca-certificates \
    tzdata \
    gcc \
    g++ \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for DuckDB
RUN CGO_ENABLED=1 GOOS=linux go build -a -o lawrence ./cmd/all-in-one

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
FROM debian:bookworm-slim

# Install runtime dependencies (including sqlite and C++ libs for DuckDB)
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    curl \
    libsqlite3-0 \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1001 lawrence && \
    useradd -u 1001 -g lawrence -s /bin/bash -m lawrence

# Set working directory
WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/lawrence .

# Copy frontend build
COPY --from=frontend-builder /app/dist ./ui/dist

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