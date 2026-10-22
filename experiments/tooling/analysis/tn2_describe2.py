import pandas as pd, json, glob, re, statistics as st
B='D:/bench-work/out/official/'
order=['leastloaded','sed','hybrid-linucb','hybrid-linucb-d-g095','hybrid-linucb-d-g098','hybrid-linucb-d-g099','hybrid-linucb-d-arm-g098']
for s in ['E2-perm','E2-trans','E2-onoff']:
    c=pd.read_csv(B+f'{s}/analysis/cells.csv')
    g=c.groupby('arm')
    t=pd.DataFrame({'b1':g.build_1_makespan_s.median(),'b2':g.build_2_makespan_s.median(),'pre':g.drift_pre_all_share.median(),
       'post':g.drift_post_share.median(),'early':g.drift_post_early_share.median(),'late':g.drift_post_late_share.median(),
       'excess':g.excess_target_share_post.median(),'tta':g.time_to_adapt_rank.median(),'cens':g.time_to_adapt_censored.sum(),
       'w5last':g.apply(lambda x:(x.last_finisher_address=='worker-5:50051').mean())}).reindex(order)
    print('==',s); print(t.round(3).to_string())
    if s=='E2-onoff':
        print('on n med', c.drift_on_phases_n.median(), 'off n med', c.drift_off_phases_n.median(), 'boundary nonempty', (c.drift_build_boundary_phases.fillna('')!='').sum())
# onoff event stats
ons=[];durs={'on':[],'off':[]}
for f in glob.glob(B+'E2-onoff/events_*_round*.csv'):
    if re.search(r'_round0\.csv$',f): continue
    e=pd.read_csv(f); e=e[e.kind.isin(['drift_on','drift_off'])&~e.detail.astype(str).str.startswith('cell_end')]
    ons.append((e.kind=='drift_on').sum())
print('E2-onoff on-events per cell: median', st.median(ons), 'min', min(ons), 'max', max(ons), 'cells', len(ons))
# trans exactly one pair
k=0;n=0
for f in glob.glob(B+'E2-trans/events_*_round*.csv'):
    if re.search(r'_round0\.csv$',f): continue
    e=pd.read_csv(f); e=e[e.kind.isin(['drift_on','drift_off'])&~e.detail.astype(str).str.startswith('cell_end')]
    n+=1; k+= (list(e.kind)==['drift_on','drift_off'])
print('E2-trans exactly one pair', k, '/', n)
