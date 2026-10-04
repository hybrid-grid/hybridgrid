# syntax=docker/dockerfile:1.4

# =============================================================================
# Stage 0: UI Builder - Builds the dashboard's React frontend
# =============================================================================
FROM node:20-alpine AS ui-builder

WORKDIR /app/internal/observability/ui/web

# Copy package manifests first for better layer caching
COPY internal/observability/ui/web/package.json internal/observability/ui/web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-fund --no-audit

COPY internal/observability/ui/web/ ./
RUN npm run build

# =============================================================================
# Stage 1: Builder - Compiles all Go binaries
# =============================================================================
FROM golang:1.25-alpine AS builder

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

# Pull in the pre-built dashboard frontend so hg-dashboard's go:embed
# directive (internal/observability/ui/server.go) has web/dist to embed
COPY --from=ui-builder /app/internal/observability/ui/web/dist ./internal/observability/ui/web/dist

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

# hg-dashboard: Standalone dashboard service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /bin/hg-dashboard ./cmd/hg-dashboard

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
# Stage 5: hg-dashboard - Standalone dashboard image
# =============================================================================
FROM scratch AS hg-dashboard

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /bin/hg-dashboard /usr/local/bin/hg-dashboard

USER 65534:65534

# Dashboard HTTP port (SPA, REST API, WebSocket)
EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/hg-dashboard"]
CMD ["serve"]

# =============================================================================
# Stage 6: hg-dashboard-web - Standalone frontend image
# =============================================================================
FROM nginxinc/nginx-unprivileged:1.27-alpine AS hg-dashboard-web

ENV API_BACKEND_URL=http://hg-dashboard:8081 \
    NGINX_ENVSUBST_FILTER=^API_BACKEND_URL$

COPY --from=ui-builder /app/internal/observability/ui/web/dist /usr/share/nginx/html
COPY --chmod=755 internal/observability/ui/web/docker/10-validate-backend.sh /docker-entrypoint.d/10-validate-backend.sh
COPY internal/observability/ui/web/docker/default.conf.template /etc/nginx/templates/default.conf.template

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/health || exit 1

# =============================================================================
# Stage 7: All-in-one image (for development/testing)
# =============================================================================
FROM alpine:3.19 AS all-in-one

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S hybridgrid && adduser -S hybridgrid -G hybridgrid

COPY --from=builder /bin/hg-coord /usr/local/bin/
COPY --from=builder /bin/hg-worker /usr/local/bin/
COPY --from=builder /bin/hgbuild /usr/local/bin/
COPY --from=builder /bin/hg-dashboard /usr/local/bin/

USER hybridgrid

# Default to coordinator
ENTRYPOINT ["/usr/local/bin/hg-coord"]
CMD ["serve"]
