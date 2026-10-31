#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUNNER="$ROOT/scripts/bench_exp.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
cat > "$TMP/docker" <<'SH'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$DOCKER_CALL_LOG"
if [[ $1 == compose && $2 == -p && $4 == -f && $6 == config && $7 == --images ]]; then
    printf '%s\n' hgexp-coordinator:latest hgexp-builder:latest
elif [[ $1 == image && $2 == inspect && $3 == --format ]]; then
    case $5 in
        sha256:coord|hgexp-coordinator:latest) printf '%s\n' sha256:coord ;;
        sha256:builder) [[ ${MISSING_BUILDER:-0} == 0 ]] && printf '%s\n' sha256:builder ;;
        hgexp-builder:latest) printf '%s\n' "${CURRENT_BUILDER:-sha256:builder}" ;;
    esac
else
    exit 70
fi
SH
chmod +x "$TMP/docker"
export PATH="$TMP:$PATH" DOCKER_CALL_LOG="$TMP/calls"
OUT_DIR="$TMP"
PROJECT=hgexp
COMPOSE_FILE="$TMP/compose.yml"
cat > "$TMP/meta.json" <<'JSON'
{"image_ids":{"coordinator":"sha256:coord","builder":"sha256:builder"}}
JSON
die() { printf 'ERROR: %s\n' "$*" >&2; return 1; }
compose() { docker compose -p "$PROJECT" -f "$COMPOSE_FILE" "$@"; }
source <(awk '/^verify_resume_images\(\) \{/ {copy=1} /^FAIL_REASON=/ {copy=0} copy' "$RUNNER")

verify_resume_images
if CURRENT_BUILDER=sha256:changed verify_resume_images 2> "$TMP/error"; then
    printf 'expected changed image rejection\n' >&2; exit 1
fi
grep -Fq 'builder recorded=sha256:builder' "$TMP/error"
if MISSING_BUILDER=1 verify_resume_images 2> "$TMP/error"; then
    printf 'expected missing image rejection\n' >&2; exit 1
fi
grep -Fq 'recorded_exists=missing' "$TMP/error"
! grep -Fq ' build' "$DOCKER_CALL_LOG"
printf 'bench_exp resume image tests passed\n'
