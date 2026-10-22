# Exploratory appendix 6.4(a) of ke-hoach-thi-nghiem-2.md (definition fixed before unblinding).
import json, glob, os, statistics as st, re
B='D:/bench-work/out/official/E2-perm/'
idx=[]; dropped_tasks=0; dropped_cells=0; slow5=[]; slow4=[]
for f in sorted(glob.glob(B+'tasks_*_round*.jsonl')):
    if re.search(r'_round0\.jsonl$', f): continue
    T=[json.loads(l) for l in open(f) if l.strip()]
    on=next(t for t in T if t.get('event')=='injected_event' and t.get('kind')=='drift_on')
    onset=int(on['dispatch_count'])+1
    cell={}
    for t in T:
        if t.get('event')!='task_completed': continue
        ok=t.get('success') and (t.get('compile_time_ms') or 0)>0 and (t.get('preprocessed_size_bytes') or 0)>0
        if not ok: dropped_tasks+=1; continue
        w=t['worker_address']
        if w not in ('worker-5:50051','worker-4:50051'): continue
        r=t['compile_time_ms']/(t['preprocessed_size_bytes']/1e5)
        cell.setdefault((w, int(t['dispatch_seq'])>=onset),[]).append(r)
    if len(cell)<4: dropped_cells+=1; continue
    m={k:st.median(v) for k,v in cell.items()}
    s5=m[('worker-5:50051',True)]/m[('worker-5:50051',False)]; s4=m[('worker-4:50051',True)]/m[('worker-4:50051',False)]
    slow5.append(s5); slow4.append(s4); idx.append(s5/s4)
q=lambda v: (st.median(v), st.quantiles(v,n=4)[0], st.quantiles(v,n=4)[2])
print("cells used", len(idx), "dropped cells", dropped_cells, "dropped tasks", dropped_tasks)
print("worker-5 post/pre median r: median %.2f IQR [%.2f, %.2f]"%q(slow5))
print("worker-4 post/pre median r: median %.2f IQR [%.2f, %.2f]"%q(slow4))
print("index (w5 / w4):           median %.2f IQR [%.2f, %.2f]"%q(idx))
