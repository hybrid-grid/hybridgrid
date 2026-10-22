import json, glob, os, re, collections
B='D:/bench-work/out/official/'
bad=0; cells=0; sched=collections.defaultdict(set); per_build=collections.Counter()
for f in glob.glob(B+'E*/tasks_*_round*.jsonl'):
    name=os.path.basename(f); m=re.match(r'tasks_(.+)_round(\d+)\.jsonl$', name); arm,rnd=m.groups()
    if rnd=='0': continue
    cells+=1; cnt=collections.Counter()
    for l in open(f):
        x=json.loads(l)
        if x.get('event')!='task_completed': continue
        cnt[x['build_id']]+=1; sched[arm].add(x.get('scheduler'))
        if not x.get('success') or x.get('from_cache'): bad+=1
    for n in cnt.values(): per_build[n]+=1
print('cells', cells, '| tasks per build distribution', dict(per_build), '| failed or cache-hit tasks', bad)
for a in sorted(sched): print('  %-28s -> %s' % (a, sorted(sched[a])))
