# syntax=docker/dockerfile:1.4

# =============================================================================
# Stage 1: Builder - Compiles all Go binaries
# =============================================================================
FROM golang:1.22-alpine AS builder

# Install git for version info and ca-certificates for HTTPS
RUN apk add --no-cache git ca-certificates tzdata

# Set environment for static compilation
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies with BuildKit cache
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build version info from git
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# Build all three binaries with BuildKit cache
# hg-coord: Coordinator server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /bin/hg-coord ./cmd/hg-coord

# hg-worker: Worker agent
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /bin/hg-worker ./cmd/hg-worker

# hgbuild: CLI tool
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /bin/hgbuild ./cmd/hgbuild

# =============================================================================
# Stage 2: hg-coord - Coordinator image
# =============================================================================
FROM scratch AS hg-coord

# Copy CA certificates for HTTPS (needed for external API calls)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /bin/hg-coord /usr/local/bin/hg-coord

# Copy embedded dashboard assets (if any static files)
# The dashboard uses go:embed, so assets are in binary

# Run as non-root user (numeric UID for scratch)
USER 65534:65534

# Expose gRPC and HTTP ports
EXPOSE 9000 8080

# Health check will be handled by docker-compose/k8s
# scratch doesn't have curl/wget, so we rely on external health probes

ENTRYPOINT ["/usr/local/bin/hg-coord"]
CMD ["serve"]

# =============================================================================
# Stage 3: hg-worker - Worker image
# =============================================================================
FROM scratch AS hg-worker

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /bin/hg-worker /usr/local/bin/hg-worker

USER 65534:65534

# Expose gRPC and metrics ports
EXPOSE 50052 9090

ENTRYPOINT ["/usr/local/bin/hg-worker"]
CMD ["serve"]

# =============================================================================
# Stage 4: hgbuild - CLI image (for CI/CD pipelines)
# =============================================================================
FROM scratch AS hgbuild

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /bin/hgbuild /usr/local/bin/hgbuild

USER 65534:65534

ENTRYPOINT ["/usr/local/bin/hgbuild"]

# =============================================================================
# Stage 5: All-in-one image (for development/testing)
# =============================================================================
FROM alpine:3.19 AS all-in-one

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S hybridgrid && adduser -S hybridgrid -G hybridgrid

COPY --from=builder /bin/hg-coord /usr/local/bin/
COPY --from=builder /bin/hg-worker /usr/local/bin/
COPY --from=builder /bin/hgbuild /usr/local/bin/

USER hybridgrid

# Default to coordinator
ENTRYPOINT ["/usr/local/bin/hg-coord"]
CMD ["serve"]
