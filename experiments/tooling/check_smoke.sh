#!/bin/bash
# Check one smoke cell directory (plan 3.2 item 3). Usage: check_smoke.sh <dir>
. /d/bench-work/env.sh
d=$1
echo "== $d"
tail -n 2 "$d.log"
cat "$d/results.csv"
echo "failures:"; tail -n +2 "$d/failures.csv"
f=$(ls "$d"/tasks_*_round1.jsonl)
python3 - "$f" "$d" <<'PY'
import json, sys, glob, csv
tasks=[json.loads(l) for l in open(sys.argv[1]) if l.strip()]
done=[t for t in tasks if t.get("event")=="task_completed"]
print("task_completed", len(done), "failed", sum(1 for t in done if not t.get("success")))
ev=[t for t in tasks if t.get("event")=="injected_event"]
for e in ev: print("event", e.get("kind"), e.get("dispatch_count"), e.get("detail"), e.get("ts"))
on=[e for e in ev if e.get("kind")=="drift_on"]
if on:
    c=int(on[0]["dispatch_count"]); post=sum(1 for t in done if int(t["dispatch_seq"])>c)
    print("first drift_on", c, "in[101,160]", 101<=c<=160, "post completed", post)
PY
ls "$d" | head -40 | tr '\n' ' '; echo
