# Agent Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the agent workflow layer and one deterministic snapshot command while keeping the CLI small.

**Architecture:** Documentation and prompts live as plain repo files. `cvx normalize` lives in a focused `internal/normalize` package and reuses existing CV and variant validation logic. HTML templates are examples only and are not executed by the CLI in this stage.

**Tech Stack:** Go 1.22, YAML input via existing `internal/cv`, JSON output via standard library, Markdown docs/prompts, plain HTML templates.

---

### Task 1: Agent Skill and Prompts

- Create `skills/cvx/SKILL.md`.
- Create prompt files under `prompts/`: `tailor.md`, `compress.md`, `verify.md`, `ats-review.md`, `recruiter-review.md`, `founder-review.md`.
- Keep all prompt language restrictive around factual accuracy.
- Commit as `docs: add cvx agent skill and prompts`.

### Task 2: Normalize Command

- Create `internal/normalize/normalize.go` and `internal/normalize/normalize_test.go`.
- Add `cvx normalize [--input cv.yaml] [--variant variants/name.yaml] [--output output/current.json]`.
- Reuse validation and variant project filtering semantics.
- Write indented JSON with trailing newline.
- Add help text.
- Commit as `feat: add normalized cv snapshots`.

### Task 3: HTML Template Examples

- Create `templates/html/minimal.html.tmpl` and `templates/html/founder.html.tmpl`.
- Document that these are presentation examples for agents/design models and are not wired into CLI rendering yet.
- Commit as `docs: add html template examples`.

### Task 4: README and Final Verification

- Create/update `README.md` with positioning, workflow, command usage, and why this is not a resume builder.
- Run `go test ./...`, `cvx lint`, `cvx normalize`, `cvx render`, `cvx schema`.
- Commit as `docs: explain agent-first workflow`.
