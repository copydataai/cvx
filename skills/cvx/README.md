# cvx Agent Skill

This directory packages the `cvx` workflow as a local agent skill. It is meant for agents that can read local skill instructions, and for generic agents that can be given the same instructions as prompt context.

## What Is Included

```txt
skills/cvx/SKILL.md                         # core agent workflow
skills/cvx/README.md                        # installation and use notes
skills/cvx/examples/tailor-founder-engineer.md # complete example prompt
```

## Install For Codex-Style Local Skills

Copy or symlink this directory into the skills directory used by your agent runtime.

Copy example:

```bash
mkdir -p ~/.codex/skills
cp -R skills/cvx ~/.codex/skills/cvx
```

Symlink example:

```bash
mkdir -p ~/.codex/skills
ln -s "$(pwd)/skills/cvx" ~/.codex/skills/cvx
```

Use a symlink while developing the skill, because edits in this repository are reflected immediately wherever the symlink points. Use a copy when you want a stable snapshot.

If your agent uses a different skills directory, replace `~/.codex/skills` with that path.

## Use With Generic Agents

For agents without a local skill loader, paste or attach these files as context:

1. `AGENTS.md`
2. `skills/cvx/SKILL.md`
3. The relevant prompt from `prompts/*.md`
4. `cv.yaml`
5. The target variant from `variants/*.yaml`, if any

Then ask the agent to edit only structured source files unless the task is explicitly about templates or rendering.

## Standard Commands

Prefer the bundled workflow when a variant is available:

```bash
cvx check --variant variants/yc-founder-engineer.yaml
```

Use the explicit workflow when you need separate snapshots or a manual diff:

```bash
cvx lint --variant variants/yc-founder-engineer.yaml
cp output/current.json output/previous.json 2>/dev/null || true
cvx normalize --variant variants/yc-founder-engineer.yaml --output output/current.json
cvx render --variant variants/yc-founder-engineer.yaml
cvx diff --from output/previous.json --to output/current.json
```

Rendering writes the audit report at:

```txt
output/reports/last-render.json
```

## Factual Safety

The skill does not authorize the agent to invent facts. It should improve clarity, ordering, grammar, emphasis, and presentation while preserving supported source data.

If a claim is not supported, use one of these labels instead of turning it into a fact:

```txt
[Unverified] This claim needs human confirmation.
[Missing] This section has no source data.
[Suggestion] Add a metric here if accurate.
```
