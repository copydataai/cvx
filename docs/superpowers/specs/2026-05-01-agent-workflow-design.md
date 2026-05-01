# Agent Workflow Design

## Goal

Make `cvx` primarily an agent workflow contract with a small deterministic CLI kernel. The CLI should not become a large app wrapper. It should provide only the checks and reproducible artifacts that agents should not hand-roll.

## Product Shape

`cvx` has two layers:

1. Agent layer: `AGENTS.md`, `skills/cvx/SKILL.md`, and `prompts/*.md` tell agents how to tailor, compress, verify, review, and report without inventing facts.
2. Kernel layer: the Go CLI validates, normalizes, renders TeX, exports schemas, and writes reports/diffs.

Agents should edit files directly when safe. The CLI exists where deterministic behavior matters: validation, normalized snapshots, render reports, schema export, and diff reports.

## Scope

Implement:

- Repo-local `skills/cvx/SKILL.md` for agent workflows.
- Prompt files for tailor, compress, verify, ATS review, recruiter review, and founder review.
- HTML template examples as plain files, not wired into rendering yet.
- `cvx normalize` to write a canonical JSON snapshot for `cvx diff`.
- README positioning and command usage.
- Tests for normalization.

Do not implement:

- Full render backend abstraction.
- HTML rendering command.
- PDF compilation.
- Preview server.
- AI API calls or SaaS dependencies.

## CLI Rule

Only add a command if it produces deterministic output an agent should not improvise. `cvx normalize` qualifies because it produces the JSON snapshots that `cvx diff` already expects.

## Normalize Behavior

Command:

```bash
cvx normalize --input cv.yaml --variant variants/yc-founder-engineer.yaml --output output/current.json
```

Behavior:

- Load and validate `cv.yaml`.
- Optionally load and validate a variant.
- Apply variant project include/exclude behavior.
- Preserve canonical fields and section data as JSON.
- Write indented JSON with a trailing newline.
- Fail on validation errors.

## Documentation Rule

Docs should position `cvx` as:

```txt
A repo-native CV workflow for agents: a skill teaches the agent what to do; a tiny CLI gives it safe, deterministic operations.
```
