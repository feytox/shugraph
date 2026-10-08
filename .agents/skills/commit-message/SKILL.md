---
name: commit-message
description: "Use when writing a git commit message in this repo. Format: [GSD-<id>] <type>: <summary>."
---

`[GSD-<id>] <type>: <summary>`

- `<id>`: from the branch name (`git branch --show-current`); if the branch has no `GSD-<n>`, ask.
- `<type>`: `feat` | `fix` | `docs` | `refactor` | `test` | `ci` | `chore`.
- `<summary>`: English, imperative, lowercase, no trailing dot.

Example: `[GSD-5] feat: initialize frontend workspace and test page`
