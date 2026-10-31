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
    if [[ ${IMAGES_UNTAGGED:-0} == 1 ]]; then printf '%s\n' hgexp-coordinator hgexp-builder
    else printf '%s\n' hgexp-coordinator:latest hgexp-builder:latest; fi
elif [[ $1 == image && $2 == inspect && $3 == --format ]]; then
    case $5 in
        sha256:coord|hgexp-coordinator:latest) printf '%s\n' sha256:coord ;;
        sha256:builder) [[ ${MISSING_BUILDER:-0} == 0 ]] && printf '%s\n' sha256:builder ;;
        hgexp-builder:latest|hgexp-builder) [[ ${NO_BUILDER_IMAGE:-0} == 0 ]] || exit 1; printf '%s\n' "${CURRENT_BUILDER:-sha256:builder}" ;;
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
source <(awk '/^require_existing_images\(\) \{/ {copy=1} /^FAIL_REASON=/ {copy=0} copy' "$RUNNER")

verify_resume_images
# Docker Compose 29.x prints image names without the :latest tag; resume must still work.
IMAGES_UNTAGGED=1 verify_resume_images
if IMAGES_UNTAGGED=1 CURRENT_BUILDER=sha256:changed verify_resume_images 2> "$TMP/error"; then
    printf 'expected changed image rejection with untagged names\n' >&2; exit 1
fi
if CURRENT_BUILDER=sha256:changed verify_resume_images 2> "$TMP/error"; then
    printf 'expected changed image rejection\n' >&2; exit 1
fi
grep -Fq 'builder recorded=sha256:builder' "$TMP/error"
if MISSING_BUILDER=1 verify_resume_images 2> "$TMP/error"; then
    printf 'expected missing image rejection\n' >&2; exit 1
fi
grep -Fq 'recorded_exists=missing' "$TMP/error"
! grep -Fq ' build' "$DOCKER_CALL_LOG"
# REUSE_IMAGES=1 starts a fresh directory on existing images and never builds.
require_existing_images
IMAGES_UNTAGGED=1 require_existing_images
if NO_BUILDER_IMAGE=1 require_existing_images 2> "$TMP/error"; then
    printf 'expected missing reusable image rejection\n' >&2; exit 1
fi
grep -Fq 'REUSE_IMAGES=1 but image hgexp-builder:latest does not exist' "$TMP/error"
! grep -Fq ' build' "$DOCKER_CALL_LOG"
grep -Fq 'if [[ $REUSE_IMAGES == 1 ]]; then require_existing_images; else compose build; fi' "$RUNNER"
printf 'bench_exp resume image tests passed\n'
