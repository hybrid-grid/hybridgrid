- MAJOR — `supervisor_prompt.md:10` treats `analysis.done` presence as proof E1 data is unchanged. That marker hashes only analyzer code and analysis outputs, not raw experiment inputs (`analysis_marker.sh:45-64`). Therefore modified/deleted E1 results or task logs can evade supervision. Require verification against `out/official/manifest_E1.sha256`.

The `next_round()` and `current_block()` code changes otherwise look sound. Note S5 exercised the older copied script (`out/pilot-e2/S5-campaign/campaign_s5.sh:52-53`), not the new helper.

NO-GO