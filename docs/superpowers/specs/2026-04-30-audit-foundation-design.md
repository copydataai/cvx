# Audit Foundation Design

## Goal

Add the first auditability layer for `cvx` so agents can validate, render, compare, and explain changes without relying on hidden state or generated artifacts as source of truth.

This stage does not add PDF generation, preview improvements, or template registry work. It focuses on reports, warnings, schema export, and structured diffs.

## Scope

Implement:

- `cvx render` writes `output/reports/last-render.json` on every render attempt.
- `cvx schema` writes `schema/cv.schema.json` and `schema/variant.schema.json`.
- `cvx diff --from a.json --to b.json` writes `output/reports/last-diff.md`.
- CV validation warnings for agent-relevant quality issues.
- Tests for warnings, render report generation, schema export, and diff generation.

Do not implement:

- PDF compilation.
- Preview server changes.
- Template registry or `--template-name`.
- AI API integration.
- SaaS dependencies.

## Architecture

### `internal/cv`

Keep factual data models and validation here.

Add warning generation:

- Empty bullets.
- Overly long bullets.
- Duplicate bullets.
- Missing link URLs.
- Suspicious metric language without source marker.
- Too many skills.
- Too many projects.

Warnings are non-fatal. Hard validation remains limited to structural requirements that make the CV impossible or unsafe to process.

### `internal/report`

Add a focused package for audit output types and writers.

Primary types:

- `RenderReport`
- `RenderArtifact`
- `ValidationResult`

The render report should include:

- Timestamp.
- Input file path.
- Output TeX path.
- Intended PDF path, even before PDF generation exists.
- Variant path if supplied.
- Engine name, currently `tex-only`.
- Validation success/failure.
- Validation errors.
- Warnings.
- Section order.

`cvx render` should write a report for both success and failure when it can determine enough context to do so.

### `internal/schema`

Add static JSON Schema generation. Prefer static Go maps or embedded JSON strings over reflection-heavy schema generation. Static schemas are easier to review, stable for agents, and avoid unnecessary dependencies.

`cvx schema` writes:

- `schema/cv.schema.json`
- `schema/variant.schema.json`

### `internal/diff`

Add markdown diff generation for CV-shaped JSON files.

The first version compares normalized CV snapshots and reports:

- Changed top-level sections.
- Added bullets.
- Removed bullets.
- Changed wording where bullet counts align but text differs.

Inputs are JSON files for this stage. YAML diff input can be added later if useful.

Command:

```bash
cvx diff --from output/previous.json --to output/current.json
```

Output:

```txt
output/reports/last-diff.md
```

## Data Flow

Render flow:

```txt
cv.yaml
  -> load YAML
  -> validate hard errors
  -> collect warnings
  -> optionally load/apply variant
  -> render TeX
  -> write output/cv.tex
  -> write output/reports/last-render.json
```

Schema flow:

```txt
cvx schema
  -> create schema directory
  -> write cv schema
  -> write variant schema
```

Diff flow:

```txt
from.json + to.json
  -> decode CV snapshots
  -> compare sections and bullets
  -> write markdown report
```

## Error Handling

- CLI errors should be explicit and include the failing path where relevant.
- Render validation failures should prevent TeX output but still write `last-render.json` when possible.
- Report write failures should fail the command because auditability is part of the command contract.
- Diff should fail if either input path is missing or cannot be decoded as JSON.

## Testing

Add Go tests for:

- Warning collection.
- Render report JSON creation.
- Schema file creation and valid JSON decoding.
- Diff markdown generation for added, removed, and changed bullets.
- Existing lint/render tests continue to pass.

Verification commands:

```bash
go test ./...
go run ./cmd/cvx lint --variant variants/yc-founder-engineer.yaml
go run ./cmd/cvx render --variant variants/yc-founder-engineer.yaml
go run ./cmd/cvx schema
```

## Commit Strategy

Use micro commits:

1. Add warning model and tests.
2. Add render report writer and wire into render.
3. Add schema export command.
4. Add diff command and tests.
5. Update CLI help and docs if needed.

Each commit should compile and pass tests.
