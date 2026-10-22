Created [campaign_tn3.sh](D:/bench-work/campaign_tn3.sh), [supervisor_prompt_tn3.md](D:/bench-work/supervisor_prompt_tn3.md), the permitted [supervisor_tn3.sh](D:/bench-work/supervisor_tn3.sh) wrapper, and [test_campaign_tn3.sh](D:/bench-work/tests/test_campaign_tn3.sh). I appended the TN3 section to [CHANGELOG-phase2.md](D:/bench-work/CHANGELOG-phase2.md). The frozen TN1–2 files were not changed.

Run the tests in Git Bash:

```bash
bash /d/bench-work/tests/test_campaign_tn3.sh
```

After the TN3 commit and pilot thresholds are frozen, run the campaign with:

```bash
TN3_REPO=/d/bench-work/hybridgrid-tn3 EXPECT_COMMIT=<frozen-sha> \
  P90_SYN=<positive-integer> P90_ABSL=<positive-integer> \
  bash /d/bench-work/campaign_tn3.sh
```

The unchanged STRICT preflight requires the clone’s **local branch name** to be `feat/exp-readiness`; otherwise it will stop the campaign. For a clean TN3 stop, create `out/official-tn3/STOP_REQUESTED`.

I checked the embedded Python syntax and confirmed the unchanged scripts’ hashes. I did not run Bash tests, Docker, or the runner while the pilot is active.