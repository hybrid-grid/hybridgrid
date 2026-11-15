import json, csv, statistics as st, sys
from datetime import datetime
from scipy.stats import spearmanr
def ts(s): return datetime.fromisoformat(s.replace('Z','+00:00')[:26]+'+00:00').timestamp() if '.' in s else datetime.fromisoformat(s.replace('Z','+00:00')).timestamp()
for w in sys.argv[1:]:
    d=f'D:/bench-work/out/pilot-tn3/runner_{w}/'
    T=[json.loads(l) for l in open(d+'tasks_leastloaded_round1.jsonl') if 'task_completed' in l]
    r=list(csv.DictReader(open(d+'results.csv')))[0]
    end=float(r['end_ts']); start=float(r['start_ts'])
    last_disp=max(ts(t['dispatch_ts']) for t in T)
    fin=max(T,key=lambda t: ts(t['ts']))
    sizes=sorted(t['preprocessed_size_bytes'] for t in T)
    rank=sum(s<=fin['preprocessed_size_bytes'] for s in sizes)/len(sizes)
    ct=[t['compile_time_ms'] for t in T]
    print(f"== {w}: makespan {float(r['cell_makespan_s']):.1f}s, tasks {len(T)}, tail(end-last dispatch) {end-last_disp:.1f}s = {100*(end-last_disp)/(end-start):.0f}% of makespan")
    print(f"   last finisher {fin['worker_address']} ({fin['worker_cpu_millis']}m), size rank {100*rank:.0f}th pct, compile {fin['compile_time_ms']} ms")
    print(f"   compile ms median {st.median(ct)} max {max(ct)}; spearman size~time {spearmanr([t['preprocessed_size_bytes'] for t in T],ct).correlation:.2f}")
    top=sorted(T,key=lambda t:-t['preprocessed_size_bytes'])[:max(1,len(T)//20)]
    print(f"   top5% by size -> workers: {sorted(t['worker_cpu_millis'] for t in top)}")
