# Exploratory figure 6.4(b): rolling 20-decision share sent to worker-5 by dispatch_seq,
# median over 20 blocks, one line per arm, median onset/offset markers. Writes SVG.
import json, glob, re, statistics as st, sys
B='D:/bench-work/out/official/'
OUT=sys.argv[1]
arms=['leastloaded','sed','hybrid-linucb','hybrid-linucb-d-g095','hybrid-linucb-d-g098','hybrid-linucb-d-g099','hybrid-linucb-d-arm-g098']
lab={'leastloaded':'LeastLoaded','sed':'SED','hybrid-linucb':'HG-LinUCB','hybrid-linucb-d-g095':'D γ=0.95','hybrid-linucb-d-g098':'D γ=0.98','hybrid-linucb-d-g099':'D γ=0.99','hybrid-linucb-d-arm-g098':'D γ=0.98 (arm)'}
col={'leastloaded':'#d62728','sed':'#ff7f0e','hybrid-linucb':'#1f77b4','hybrid-linucb-d-g095':'#9ecae1','hybrid-linucb-d-g098':'#3182bd','hybrid-linucb-d-g099':'#08519c','hybrid-linucb-d-arm-g098':'#2ca02c'}
W=20; N=590
def series(s,a):
    curves=[];ons=[];offs=[]
    for f in glob.glob(B+f'{s}/tasks_{a}_round*.jsonl'):
        if re.search(r'_round0\.jsonl$',f): continue
        T=[json.loads(l) for l in open(f) if l.strip()]
        ev=[t for t in T if t.get('event')=='injected_event' and t.get('kind') in('drift_on','drift_off') and not str(t.get('detail','')).startswith('cell_end')]
        if ev: ons.append(int(ev[0]['dispatch_count'])+1)
        if len(ev)>1 and ev[1]['kind']=='drift_off': offs.append(int(ev[1]['dispatch_count'])+1)
        hit=[0]*(N+1)
        for t in T:
            if t.get('event')=='task_completed': hit[int(t['dispatch_seq'])]=int(t['worker_address']=='worker-5:50051')
        curves.append([sum(hit[i-W+1:i+1])/W for i in range(W,N+1)])
    med=[st.median(c[i] for c in curves) for i in range(len(curves[0]))]
    return med, (st.median(ons) if ons else None), (st.median(offs) if offs else None)
panels=['E2-perm','E2-trans','E2-onoff']
PW,PH,ML,MT,GAP=560,220,50,30,60
H=MT+len(panels)*(PH+GAP)+60; Wt=ML+PW+200
o=[f'<svg xmlns="http://www.w3.org/2000/svg" width="{Wt}" height="{H}" font-family="sans-serif" font-size="12">','<rect width="100%" height="100%" fill="white"/>']
for pi,s in enumerate(panels):
    y0=MT+pi*(PH+GAP)
    X=lambda i: ML+PW*(i-W)/(N-W); Y=lambda v: y0+PH-PH*v/0.5
    o.append(f'<text x="{ML}" y="{y0-10}" font-weight="bold">{s}: rolling share of decisions sent to worker-5 (window 20, median of 20 blocks)</text>')
    o.append(f'<rect x="{ML}" y="{y0}" width="{PW}" height="{PH}" fill="none" stroke="#999"/>')
    for v in (0,0.1,0.2,0.3,0.4,0.5):
        o.append(f'<line x1="{ML}" x2="{ML+PW}" y1="{Y(v)}" y2="{Y(v)}" stroke="#eee"/><text x="{ML-8}" y="{Y(v)+4}" text-anchor="end">{v:.1f}</text>')
    for i in (100,200,295,400,500,590):
        o.append(f'<text x="{X(i)}" y="{y0+PH+15}" text-anchor="middle">{i}</text>')
    o.append(f'<line x1="{X(295)}" x2="{X(295)}" y1="{y0}" y2="{y0+PH}" stroke="#bbb" stroke-dasharray="2,3"/>')
    onm=[];offm=[]
    for a in arms:
        med,on,off=series(s,a); onm.append(on); offm+= [off] if off else []
        pts=' '.join(f'{X(W+i):.1f},{Y(v):.1f}' for i,v in enumerate(med))
        o.append(f'<polyline fill="none" stroke="{col[a]}" stroke-width="1.4" points="{pts}"/>')
    on=st.median([x for x in onm if x]); o.append(f'<line x1="{X(on)}" x2="{X(on)}" y1="{y0}" y2="{y0+PH}" stroke="black" stroke-dasharray="5,3"/><text x="{X(on)+3}" y="{y0+12}">onset</text>')
    if s=='E2-trans' and offm:
        off=st.median(offm); o.append(f'<line x1="{X(off)}" x2="{X(off)}" y1="{y0}" y2="{y0+PH}" stroke="black" stroke-dasharray="1,3"/><text x="{X(off)+3}" y="{y0+26}">recovery</text>')
    if s=='E2-onoff': o.append(f'<text x="{X(on)+3}" y="{y0+26}">on/off phases every ~14 s</text>')
    o.append(f'<text x="{ML+PW/2}" y="{y0+PH+32}" text-anchor="middle">dispatch_seq (build 2 starts at 296)</text>')
for k,a in enumerate(arms):
    o.append(f'<line x1="{ML+PW+20}" x2="{ML+PW+45}" y1="{MT+20+k*18}" y2="{MT+20+k*18}" stroke="{col[a]}" stroke-width="3"/><text x="{ML+PW+50}" y="{MT+24+k*18}">{lab[a]}</text>')
o.append(f'<text x="{ML}" y="{H-15}" fill="#555">Exploratory figure (plan 6.4b), not pre-registered as a test. Source: task logs in D:/bench-work/out/official.</text></svg>')
open(OUT,'w',encoding='utf-8').write('\n'.join(o)); print('wrote',OUT)
