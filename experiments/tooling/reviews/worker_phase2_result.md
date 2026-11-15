Implemented T1–T7 in [campaign.sh](D:/bench-work/campaign.sh), [supervisor.sh](D:/bench-work/supervisor.sh), [supervisor_prompt.md](D:/bench-work/supervisor_prompt.md), new [stop_campaign.sh](D:/bench-work/stop_campaign.sh) and [analysis_marker.sh](D:/bench-work/analysis_marker.sh). Added fake based [campaign tests](D:/bench-work/tests/test_campaign.sh), [supervisor tests](D:/bench-work/tests/test_supervisor.sh), and the per-file hash record in [CHANGELOG-phase2.md](D:/bench-work/CHANGELOG-phase2.md). No frozen code or `out/` files were changed.

**Validation:** Git Bash failed before reading any script (`CreateFileMapping ... Win32 error 5`). Thus `bash -n` and both Bash test suites produced no PASS/FAIL case lines. `shellcheck` is unavailable. The checks that did run produced:

```text
PASS hash campaign.sh
PASS hash supervisor.sh
PASS hash supervisor_prompt.md
PASS hash stop_campaign.sh
PASS hash analysis_marker.sh
PASS hash tests/test_campaign.sh
PASS hash tests/test_supervisor.sh
```

The changelog’s SHA-256 is `de9a7b7248d466405f758c7c8aba441a56ddc462226d44f63d695c74ee7e5ea1`. The remaining validation is to run the two test scripts and `bash -n` in a working Git Bash environment. Real Docker signals were not tested. The plan’s manual creation of markers in official E1 outputs was omitted because your instruction forbids modifying `out/`.