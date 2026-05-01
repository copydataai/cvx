# Installing The cvx Agent Skill

The `cvx` agent skill lives in:

```txt
skills/cvx
```

It is a local instruction package. Installing it means placing that directory where your agent runtime looks for skills, or giving the same files to a generic agent as context.

## Copy The Skill

Use a copy when you want a stable snapshot of the skill.

```bash
mkdir -p ~/.codex/skills
cp -R skills/cvx ~/.codex/skills/cvx
```

If your agent uses another skills directory, replace `~/.codex/skills` with that directory.

## Symlink The Skill

Use a symlink while developing or testing changes in this repository.

```bash
mkdir -p ~/.codex/skills
ln -s "$(pwd)/skills/cvx" ~/.codex/skills/cvx
```

If a previous copy or symlink exists, inspect it before replacing it.

## Use With Generic Agents

For agents that do not support local skill directories, include these files in the conversation or agent context:

```txt
AGENTS.md
skills/cvx/SKILL.md
skills/cvx/README.md
prompts/tailor.md
prompts/verify.md
cv.yaml
variants/<target>.yaml
```

For a complete tailoring prompt, use:

```txt
skills/cvx/examples/tailor-founder-engineer.md
```

## Use Repository Prompts

The `prompts/*.md` files are task-specific guardrails. They can be printed with the CLI when available:

```bash
cvx prompt tailor
cvx prompt verify
cvx prompt compress
cvx prompt ats-review
cvx prompt recruiter-review
cvx prompt founder-review
```

You can also paste the prompt file directly into an agent conversation. Prompts do not replace `cv.yaml`; they constrain how the agent should edit the structured source.

## Recommended Check

After content edits, run:

```bash
cvx check --variant variants/yc-founder-engineer.yaml
```

For manual snapshot comparison, run:

```bash
cp output/current.json output/previous.json 2>/dev/null || true
cvx normalize --variant variants/yc-founder-engineer.yaml --output output/current.json
cvx render --variant variants/yc-founder-engineer.yaml
cvx diff --from output/previous.json --to output/current.json
```

Rendering writes:

```txt
output/reports/last-render.json
```
