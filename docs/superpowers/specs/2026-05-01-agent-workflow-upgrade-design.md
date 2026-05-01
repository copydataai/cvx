# Agent Workflow Upgrade Design

## Goal

Implement the complete next improvement set while preserving the product principle: the skill is primary and the CLI stays a small deterministic kernel.

## Scope

Implement:

- `cvx check` orchestration for lint, normalize, render, and optional diff.
- Bullet provenance fields with backward-compatible YAML scalar bullets.
- Richer normalized snapshots containing input, variant, target, section order, and filtered CV.
- `cvx prompt <name>` for printing repo prompts.
- HTML rendering as a lightweight backend option.
- `examples/` with realistic fake source data.
- CLI UX cleanup: help handling, version command, generated file hygiene.

Do not implement:

- AI API calls.
- SaaS dependencies.
- Browser preview server.
- PDF compilation.

## Design Notes

Bullets become structured internally but remain YAML-compatible with the existing scalar list style. Agents can optionally use object bullets when provenance is needed.

Normalize outputs a snapshot object instead of a bare CV. Diff accepts both the new snapshot shape and legacy CV JSON for compatibility.

HTML rendering is intentionally simple and file-based. It proves the render model can feed non-LaTeX presentation without introducing a large plugin system.
