# AGENTS.md

This repository is designed for AI agents.

## Mission

Help the human maintain, tailor, validate, and render a professional CV without corrupting factual accuracy.

The agent should optimize clarity, relevance, structure, and presentation.

The agent must not invent facts.

## Golden rule

Edit structured source first.

Prefer:

```txt
cv.yaml
variants/*.yaml
prompts/*.md
```

Avoid editing generated TeX unless the task is explicitly visual/layout-related.

## Forbidden behavior

Do not invent employers, titles, dates, degrees, metrics, revenue, funding, publications, awards, certifications, technologies, links, locations, or claims of impact.

If a metric is missing, do not create one.

Bad:

```txt
Increased lead conversion by 40%.
```

Good:

```txt
Built lead management workflows for cleaning operators.
```

## Allowed behavior

You may improve grammar, bullet clarity, ordering, concision, section emphasis, role relevance, ATS readability, template spacing, compile errors, YAML structure, and validation schemas.

## Required workflow

Before changing content:

1. Read `cv.yaml`.
2. Read the target variant, if one exists.
3. Identify which facts support the target.
4. Edit the smallest necessary set of files.

After changing content:

```bash
cvx lint
cvx render
```

Rendering writes an audit report:

```txt
output/reports/last-render.json
```

If a variant is used:

```bash
cvx lint --variant variants/yc-founder-engineer.yaml
cvx render --variant variants/yc-founder-engineer.yaml
```

When comparing two CV snapshots:

```bash
cvx diff --from output/previous.json --to output/current.json
```

When editor or agent schema files are needed:

```bash
cvx schema
```

## Reporting format

```txt
Changed:
- ...

Did not change:
- ...

Validation:
- cvx lint: pass/fail
- cvx render: pass/fail
- render report: pass/fail
- cvx diff: pass/fail, if used
- cvx schema: pass/fail, if used

Notes:
- ...
```

## Handling uncertainty

If something is unclear, do not guess. Use one of:

```txt
[Unverified] This claim needs human confirmation.
[Missing] This section has no source data.
[Suggestion] Add a metric here if accurate.
```

## File ownership

### Agent-editable

```txt
cv.yaml
variants/*.yaml
prompts/*.md
templates/**/*.tex.njk
internal/**
cmd/**
README.md
```

### Generated

```txt
output/**
.cvx/**
```

Do not treat generated files as source of truth.
