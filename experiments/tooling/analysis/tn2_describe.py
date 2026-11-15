import pandas as pd
pd.set_option('display.width',250); pd.set_option('display.max_columns',30)
B='D:/bench-work/out/official/'
order=['leastloaded','sed','hybrid-linucb','hybrid-linucb-d-g095','hybrid-linucb-d-g098','hybrid-linucb-d-g099','hybrid-linucb-d-arm-g098']
cols={'E2-perm':['makespan_s','drift_pre_all_share','drift_post_share','drift_post_early_share','drift_post_late_share','time_to_adapt_rank','excess_target_share_post'],
      'E2-trans':['makespan_s','drift_pre_all_share','drift_post_share','target_after_recovery_share','time_to_adapt_rank'],
      'E2-onoff':['makespan_s','drift_pre_all_share','drift_post_share','drift_on_phases_share','drift_off_phases_share']}
for s,cs in cols.items():
    c=pd.read_csv(B+f'{s}/analysis/cells.csv')
    print(f"\n== {s} medians"); print(c.groupby('arm')[cs].median().reindex(order).round(4).to_string())
    print("last finisher worker-5 share:", c.assign(w5=c.last_finisher_address.eq('worker-5:50051')).groupby('arm').w5.mean().reindex(order).round(2).to_dict())
    p=pd.read_csv(B+f'{s}/analysis/pairs.csv'); p=p[p.metric=='makespan_s']
    print(f"-- {s} makespan family (Holm over 10)"); print(p[['treatment','baseline','median_diff','ci_lo','ci_hi','p_raw','p_holm','r_rb']].round(4).to_string(index=False))
    d=pd.read_csv(B+f'{s}/analysis_d_vs_hg/pairs.csv')
    keep=['makespan_s','drift_post_share','time_to_adapt_rank','target_after_recovery_share','drift_on_phases_share']
    d=d[d.metric.isin(keep)]
    print(f"-- {s} D vs HG (Holm over 4)"); print(d[['metric','treatment','median_diff','p_raw','p_holm','r_rb']].round(4).to_string(index=False))
c=pd.read_csv(B+'E2-lambda/analysis/cells.csv')
lo=['leastloaded','sed','hybrid-linucb-l0','hybrid-linucb-l025','hybrid-linucb','hybrid-linucb-l100','hybrid-linucb-d-g098-l0','hybrid-linucb-d-g098-l025','hybrid-linucb-d-g098','hybrid-linucb-d-g098-l100']
print("\n== E2-lambda medians"); print(c.groupby('arm')[['makespan_s','drift_pre_all_share','drift_post_share','time_to_adapt_rank']].median().reindex(lo).round(4).to_string())
