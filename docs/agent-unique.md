# Agent-Unique Direction

`cvx` treats agent edits as auditable transactions over professional identity data.

Agent-unique layers:

- Source data: `cv.yaml`
- Provenance registry: `sources` and bullet `sources`
- Targeting: `variants/*.yaml`
- Session transactions: `.cvx/sessions/<id>`
- Policy gates: `cvx policy`
- Multi-agent roles: `prompts/roles/*.md`
- Adversarial benchmarks: `benchmarks/adversarial/*.md`

Recommended multi-agent flow:

```bash
cvx session start --goal "Tailor for YC founder engineer"
cvx session snapshot --label before --variant variants/yc-founder-engineer.yaml
# Tailor agent edits structured files.
cvx session snapshot --label after --variant variants/yc-founder-engineer.yaml
cvx session verify --variant variants/yc-founder-engineer.yaml
cvx session report
```
