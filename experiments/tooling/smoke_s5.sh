#!/bin/bash
# S5 (plan 3.2 item 4): real Docker stop/resume test of the phase-2 campaign
# tooling. Uses a test copy of campaign.sh (REPS=2, one arm, only E2-perm) in
# out/pilot-e2/S5-campaign. Pilot data, never official.
. /d/bench-work/env.sh
T=$BENCH_OUT/pilot-e2/S5-campaign
rm -rf "$T"; mkdir -p "$T"
sed -e 's/^REPS=20$/REPS=2/' \
    -e 's/^A7=.*/A7="hybrid-linucb"/' \
    -e 's/^D_ARMS=.*/D_ARMS="hybrid-linucb"/' \
    -e 's/^ORDER=.*/ORDER=(E2-perm)/' \
    /d/bench-work/campaign.sh > "$T/campaign_s5.sh"
grep -E '^(REPS|A7|D_ARMS|ORDER)=' "$T/campaign_s5.sh"
printf 'E2_LAMBDA=SKIP\n' > "$T/PHASE2_APPROVED"
launch() { CAMPAIGN=$T EXPECT_COMMIT=1b8339d BUILDER_CPUS=2.0 bash "$T/campaign_s5.sh" >> "$T/stdout.log" 2>&1 & echo $! > "$T/launcher.pid"; }

echo "$(date '+%F %T') S5 launch 1"
launch
# Wait until block 2 is running (runner log announces round 2), then 60 s more.
for i in $(seq 1 240); do
    grep -q "round 2" "$T/E2-perm/runner.log" 2>/dev/null && break
    sleep 5
done
sleep 60
echo "$(date '+%F %T') S5 stop (block 2 in progress)"
t0=$SECONDS
CAMPAIGN=$T bash /d/bench-work/stop_campaign.sh "S5 stop test"; echo "stop rc=$? after $((SECONDS - t0)) s"
sleep 3
echo "containers (all states): $(docker ps -aq --filter label=com.docker.compose.project=hgexp | wc -l)"
echo "runner procs: $(grep -l bench_exp.sh /proc/[0-9]*/cmdline 2>/dev/null | wc -l)"
echo "pid files: $(ls "$T"/*.pid 2>/dev/null | tr '\n' ' ')"
cp "$T/E2-perm/results.csv" "$T/results_after_stop.csv"
rm -f "$T/STOP_REQUESTED"
echo "$(date '+%F %T') S5 relaunch"
launch
wait_pid=$(cat "$T/launcher.pid")
while kill -0 "$wait_pid" 2>/dev/null; do sleep 10; done
echo "$(date '+%F %T') S5 campaign finished"
tail -n 25 "$T/campaign.log"
echo "--- results.csv"; cat "$T/E2-perm/results.csv"
echo "--- results after stop"; cat "$T/results_after_stop.csv"
echo "--- interruptions"; cat "$T/E2-perm/interruptions.csv" 2>/dev/null
echo "--- reruns"; cat "$T/E2-perm/reruns.csv" 2>/dev/null || echo "(none)"
echo "--- archive"; find "$T/E2-perm/archive" -type f | sed "s#$T/##"
echo "containers at end: $(docker ps -aq --filter label=com.docker.compose.project=hgexp | wc -l)"
echo "S5 DONE"
