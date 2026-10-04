#!/bin/sh
set -eu

backend_url=${API_BACKEND_URL:-}
line_breaks=$(printf '%s' "$backend_url" | wc -l | tr -d ' ')
if [ "$line_breaks" -ne 0 ] || ! printf '%s' "$backend_url" | grep -Eq '^https?://(\[[0-9A-Fa-f:.]+\]|[A-Za-z0-9._-]+)(:[0-9]{1,5})?$'; then
  echo >&2 "ERROR: API_BACKEND_URL must be an http:// or https:// origin without credentials, query, fragment, or path"
  exit 1
fi

backend_authority=${backend_url#*://}
case "$backend_authority" in
  \[*\]:*) backend_port=${backend_authority##*:} ;;
  \[*\]) backend_port= ;;
  *:*) backend_port=${backend_authority##*:} ;;
  *) backend_port= ;;
esac

if [ -n "$backend_port" ] && { [ "$backend_port" -lt 1 ] || [ "$backend_port" -gt 65535 ]; }; then
  echo >&2 "ERROR: API_BACKEND_URL port must be between 1 and 65535"
  exit 1
fi
