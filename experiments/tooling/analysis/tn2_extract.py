import pandas as pd, numpy as np
B='D:/bench-work/out/official/'
def fam(s, d): return pd.read_csv(B+f'{s}/{d}/pairs.csv')
rows=[('H2.1a','E2-perm','analysis_d_vs_hg','time_to_adapt_rank','hybrid-linucb-d-g098','hybrid-linucb'),
      ('H2.1b','E2-perm','analysis_d_vs_hg','drift_post_share','hybrid-linucb-d-g098','hybrid-linucb'),
      ('H2.2','E2-trans','analysis_d_vs_hg','target_after_recovery_share','hybrid-linucb-d-g098','hybrid-linucb'),
      ('H2.3a','E2-perm','analysis','makespan_s','hybrid-linucb-d-g098','leastloaded'),
      ('H2.3b','E2-onoff','analysis','makespan_s','hybrid-linucb-d-g098','leastloaded')]
print("== audit: family sizes and n")
for h,s,d,m,t,b in rows:
    p=fam(s,d); f=p[p.metric==m]
    r=f[(f.treatment==t)&(f.baseline==b)].iloc[0]
    c=pd.read_csv(B+f'{s}/analysis/cells.csv')
    nan=c[c.arm.isin([t,b])][m].isna().sum()
    print(f"{h} {s} {m}: family size {len(f)}, n={r.n}, n_zero={r.n_zero}, NaN cells={nan}")
print("\n== confirmatory")
for h,s,d,m,t,b in rows:
    r=fam(s,d); r=r[(r.metric==m)&(r.treatment==t)&(r.baseline==b)].iloc[0]
    print(f"{h} {s} {m}: T={r.treatment_median:.4g} B={r.baseline_median:.4g} diff={r.median_diff:.4g} CI=[{r.ci_lo:.4g},{r.ci_hi:.4g}] p_raw={r.p_raw:.3g} p_holm={r.p_holm:.3g} r={r.r_rb:.3f}")
c=pd.read_csv(B+'E2-perm/analysis/cells.csv')
print("\ncensored time_to_adapt (E2-perm):"); print(c.groupby('arm').time_to_adapt_censored.sum().to_string())
t=pd.read_csv(B+'E2-trans/analysis/cells.csv')
print("\nE2-trans recovery tasks NaN per arm:"); print(t.groupby('arm').target_after_recovery_share.apply(lambda x: x.isna().sum()).to_string())
