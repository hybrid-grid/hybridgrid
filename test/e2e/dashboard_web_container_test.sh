#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
IMAGE_TAG="hybridgrid-dashboard-web:test"
RUN_ID="hg-dashboard-web-test-$$"
NETWORK_NAME="$RUN_ID"
BACKEND_NAME="$RUN_ID-backend"
FRONTEND_NAME="$RUN_ID-frontend"
TEMP_DIR=$(mktemp -d)

cleanup() {
  docker rm -f "$FRONTEND_NAME" "$BACKEND_NAME" >/dev/null 2>&1 || true
  docker network rm "$NETWORK_NAME" >/dev/null 2>&1 || true
  rm -rf "$TEMP_DIR"
}
trap cleanup EXIT INT TERM

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

assert_contains() {
  haystack=$1
  needle=$2
  description=$3

  case "$haystack" in
    *"$needle"*) ;;
    *) fail "$description" ;;
  esac
}

wait_for_url() {
  url=$1
  attempts=30

  while [ "$attempts" -gt 0 ]; do
    if curl --fail --silent --show-error "$url" >/dev/null 2>&1; then
      return 0
    fi
    attempts=$((attempts - 1))
    sleep 1
  done

  return 1
}

cat >"$TEMP_DIR/backend.conf" <<'EOF'
events {}
http {
  server {
    listen 8081;

    location = /api/v1/stats {
      default_type application/json;
      return 200 '{"source":"runtime-backend"}';
    }
  }
}
EOF

docker build --target hg-dashboard-web -t "$IMAGE_TAG" "$ROOT_DIR"
docker network create "$NETWORK_NAME" >/dev/null

docker run -d \
  --name "$BACKEND_NAME" \
  --network "$NETWORK_NAME" \
  -v "$TEMP_DIR/backend.conf:/etc/nginx/nginx.conf:ro" \
  nginx:alpine >/dev/null

docker run -d \
  --name "$FRONTEND_NAME" \
  --network "$NETWORK_NAME" \
  -p 127.0.0.1::8080 \
  -e "API_BACKEND_URL=http://$BACKEND_NAME:8081" \
  "$IMAGE_TAG" >/dev/null

HOST_PORT=$(docker port "$FRONTEND_NAME" 8080/tcp | awk -F: 'NR == 1 { print $NF }')
BASE_URL="http://127.0.0.1:$HOST_PORT"
wait_for_url "$BASE_URL/health" || fail "frontend health endpoint did not become ready"

health=$(curl --fail --silent --show-error "$BASE_URL/health")
[ "$health" = "OK" ] || fail "frontend health endpoint returned unexpected body"

index_html=$(curl --fail --silent --show-error "$BASE_URL/")
assert_contains "$index_html" '<div id="root"></div>' "root path did not serve the SPA"

deep_route=$(curl --fail --silent --show-error "$BASE_URL/builds/example-build")
assert_contains "$deep_route" '<div id="root"></div>' "deep route did not fall back to the SPA"

stats=$(curl --fail --silent --show-error "$BASE_URL/api/v1/stats")
assert_contains "$stats" '"source":"runtime-backend"' "API request did not reach API_BACKEND_URL"

invalid_output="$TEMP_DIR/invalid-output.log"
if docker run --rm -e 'API_BACKEND_URL=http://backend:8081/path' "$IMAGE_TAG" >"$invalid_output" 2>&1; then
  fail "container accepted API_BACKEND_URL with a path"
fi
assert_contains "$(cat "$invalid_output")" 'API_BACKEND_URL' "invalid URL failure did not identify API_BACKEND_URL"

printf 'PASS: dashboard web container serves the SPA and proxies the runtime backend\n'
