#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
source_dir="$script_dir/cmake-smoke"
if [ "$#" -eq 0 ]; then
	build_dir=$(mktemp -d)
	trap 'rm -rf -- "$build_dir"' EXIT
else
	build_dir=$1
	case "$build_dir" in
		''|/|.|..) printf 'unsafe build directory: %s\n' "$build_dir" >&2; exit 1 ;;
	esac
	rm -rf -- "$build_dir"
fi

CC=hgcc CXX=hgcxx cmake -S "$source_dir" -B "$build_dir" -G Ninja -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
hgbuild ninja -C "$build_dir" -j4
"$build_dir/hg_cmake_smoke"
printf 'compile commands: '
grep -c '"command"' "$build_dir/compile_commands.json"
