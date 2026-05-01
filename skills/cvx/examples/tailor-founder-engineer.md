# Example Agent Prompt: Tailor Founder-Engineer CV

You are working in a `cvx` repository. Tailor the CV for the YC founder-engineer variant without inventing facts.

## Scope

Editable source files:

```txt
cv.yaml
variants/yc-founder-engineer.yaml
prompts/*.md
skills/cvx/SKILL.md
```

Do not edit generated files in `output/**` as source of truth. Do not edit TeX unless the task becomes explicitly visual or layout-related.

## Required Reading

Before changing content, read:

```txt
AGENTS.md
skills/cvx/SKILL.md
cv.yaml
variants/yc-founder-engineer.yaml
prompts/tailor.md
prompts/founder-review.md
```

Identify which existing facts support the YC founder-engineer target. Do not add employers, titles, dates, degrees, metrics, revenue, funding, awards, publications, certifications, links, locations, technologies, or impact claims that are not already supported by source data.

## Task

Tailor the CV toward founder-engineer positioning by using only supported facts. Prefer:

- clearer summary language
- stronger ordering of existing experience and projects
- concise bullets
- product, systems, operations, infrastructure, and AI relevance where already supported
- variant-level emphasis changes when that is the smallest safe edit

If a useful improvement needs missing evidence, write it as one of:

```txt
[Unverified] This claim needs human confirmation.
[Missing] This section has no source data.
[Suggestion] Add a metric here if accurate.
```

## Validation Workflow

Run the standard workflow first:

```bash
cvx check --variant variants/yc-founder-engineer.yaml
```

Then create or update comparable snapshots and render artifacts:

```bash
cp output/current.json output/previous.json 2>/dev/null || true
cvx normalize --variant variants/yc-founder-engineer.yaml --output output/current.json
cvx render --variant variants/yc-founder-engineer.yaml
```

If `output/previous.json` exists, compare the snapshots:

```bash
cvx diff --from output/previous.json --to output/current.json
```

Review:

```txt
output/reports/last-render.json
output/reports/last-diff.md
```

## Report Back

Use this format:

```txt
Changed:
- ...

Did not change:
- Unsupported factual claims.
- Generated output as source of truth.

Validation:
- cvx check --variant variants/yc-founder-engineer.yaml: pass/fail
- cvx normalize --variant variants/yc-founder-engineer.yaml --output output/current.json: pass/fail
- cvx render --variant variants/yc-founder-engineer.yaml: pass/fail
- cvx diff --from output/previous.json --to output/current.json: pass/fail, if used
- render report: pass/fail

Notes:
- ...
```
