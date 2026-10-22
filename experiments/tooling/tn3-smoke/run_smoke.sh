#!/bin/bash
# TN3 C0 smoke: cmake-smoke, abseil-cpp (CMake+Ninja) and CPython (make) through the new hgbuild.
set -u
cd /d/bench-work/tn3-smoke
dc() { MSYS_NO_PATHCONV=1 docker compose -p hg3smoke "$@"; }
tasks() { dc exec -T coordinator sh -c 'echo "tasks=$(grep -c task_completed /tmp/tasks.jsonl) failed=$(grep -c "\"success\":false" /tmp/tasks.jsonl)"'; }
reset() { dc exec -T coordinator sh -c ': > /tmp/tasks.jsonl'; }
dc build 2>&1 | tail -1; dc up -d 2>&1 | tail -1; sleep 8
echo "== cmake-smoke"; reset
dc exec -T builder sh -c 'sh /smoke/cmake-smoke.sh 2>&1 | tail -3'; tasks
echo "== abseil"; dc exec -T builder sh -c 'rm -rf /tmp/absl && CC=hgcc CXX=hgcxx cmake -S /src/abseil-cpp -B /tmp/absl -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_STANDARD=17 -DABSL_BUILD_TESTING=OFF -DABSL_PROPAGATE_CXX_STD=ON -DCMAKE_EXPORT_COMPILE_COMMANDS=ON >/dev/null 2>&1; echo compile_cmds=$(grep -c "\"command\"" /tmp/absl/compile_commands.json)'
reset; dc exec -T builder sh -c 's=$(date +%s); hgbuild ninja -C /tmp/absl -j8 >/tmp/absl.log 2>&1; echo "rc=$? secs=$(( $(date +%s) - s ))"; tail -1 /tmp/absl.log'; tasks
echo "== cpython (make, as TN1-2)"
dc exec -T builder sh -c 'cd /tmp && rm -rf cpython && git clone -q --depth=1 --branch=v3.14.0 https://github.com/python/cpython.git && cd cpython && ./configure --disable-test-modules --disable-perf-trampoline >/dev/null 2>&1 && echo configured'
reset; dc exec -T builder sh -c 'cd /tmp/cpython && s=$(date +%s); hgbuild make -j8 >/tmp/py.log 2>&1; echo "rc=$? secs=$(( $(date +%s) - s ))"; ./python -c "import sys; print(sys.version.split()[0])"'; tasks
dc cp coordinator:/tmp/tasks.jsonl 'D:/bench-work/out/pilot-tn3/cpython_smoke.jsonl' >/dev/null 2>&1
dc down 2>&1 | tail -1
echo SMOKE DONE
