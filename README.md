# cvx

`cvx` is a repo-native CV workflow for agents.

A skill teaches the agent what to do. A tiny CLI gives it safe, deterministic operations.

## Why Agent-First?

A human-first CV editor optimizes for forms, buttons, and visual controls.

An agent-first CV system optimizes for:

- clear source files
- strict schemas
- auditable changes
- deterministic rendering
- local execution
- version control
- role-specific variants
- factual constraints
- automatic validation

The human should provide intent. The agent should execute safely.

## Positioning

`cvx` is not a resume builder with AI sprinkled on top.

It is a structured CV repo that agents can operate:

- `cv.yaml` contains factual source data.
- `variants/*.yaml` selects and prioritizes facts for a target role.
- `prompts/*.md` defines restrictive agent tasks.
- `skills/cvx/SKILL.md` defines the agent workflow.
- `cvx` validates, normalizes, renders, exports schemas, and reports diffs.

## Why Not Just HTML Mockups?

HTML mockups are useful for presentation exploration. They are not the source of truth.

`cvx` keeps creativity and factual control separate:

- agents may be creative in templates and visual hierarchy
- agents must be constrained around facts, dates, employers, degrees, links, and metrics
- deterministic commands produce artifacts and reports that humans can review

## Core Files

```txt
cv.yaml                  # canonical facts
variants/*.yaml          # target-specific selection and emphasis
prompts/*.md             # restrictive task prompts
skills/cvx/SKILL.md      # agent workflow
templates/html/*.tmpl    # presentation examples
schema/*.schema.json     # editor/agent schema support
output/**                # generated artifacts, not source of truth
```

## Commands

```bash
cvx lint
```

Validate `cv.yaml` and optional variant structure.

```bash
cvx normalize --variant variants/yc-founder-engineer.yaml --output output/current.json
```

Write a canonical JSON snapshot for diffing and review.

```bash
cvx render --variant variants/yc-founder-engineer.yaml
```

Write `output/cv.tex` and `output/reports/last-render.json`.

```bash
cvx diff --from output/previous.json --to output/current.json
```

Write `output/reports/last-diff.md`.

```bash
cvx schema
```

Write JSON Schema files for editor and agent support.

## Agent Workflow

A typical tailoring workflow:

```bash
cvx lint --variant variants/yc-founder-engineer.yaml
cp output/current.json output/previous.json 2>/dev/null || true
cvx normalize --variant variants/yc-founder-engineer.yaml --output output/current.json
cvx render --variant variants/yc-founder-engineer.yaml
cvx diff --from output/previous.json --to output/current.json
```

The agent should then report:

```txt
Changed:
- ...

Did not change:
- Unsupported factual claims.
- Generated output as source of truth.

Validation:
- cvx lint: pass/fail
- cvx normalize: pass/fail
- cvx render: pass/fail
- cvx diff: pass/fail, if used

Notes:
- ...
```

## Factual Accuracy Rule

Do not invent:

- employers
- titles
- dates
- degrees
- metrics
- revenue
- funding
- publications
- awards
- certifications
- technologies
- links
- locations
- impact claims

If something is unclear, mark it:

```txt
[Unverified] This claim needs human confirmation.
[Missing] This section has no source data.
[Suggestion] Add a metric here if accurate.
```

## Development

```bash
go test ./...
go run ./cmd/cvx lint --variant variants/yc-founder-engineer.yaml
go run ./cmd/cvx normalize --variant variants/yc-founder-engineer.yaml --output output/current.json
go run ./cmd/cvx render --variant variants/yc-founder-engineer.yaml
go run ./cmd/cvx schema
```

## Product Direction

The CLI should remain a small deterministic kernel.

Prefer files and conventions when agents can operate safely. Add CLI behavior only when deterministic checking, normalization, rendering, or reporting is needed.
