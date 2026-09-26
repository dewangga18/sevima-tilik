# Git Convention

Use Conventional Commit style:

```text
<type>(<scope>): <subject>
```

Scope is optional when the change clearly spans the repository.

## Types

- `feat` — new capability
- `fix` — bug fix
- `refactor` — behavior-preserving code restructuring
- `style` — visual/formatting change without logic change
- `perf` — performance improvement
- `docs` — documentation only
- `test` — tests only
- `chore` — tooling, dependency, CI, build, maintenance

## Rules

- One logical change per commit.
- Use imperative, lowercase subject text.
- Do not end the subject with a period.
- Avoid vague messages such as `update`, `changes`, `fix stuff`, or `wip`.
- Add a body only when context or trade-offs are not obvious from the diff.
- Make a checkpoint commit before a risky large rewrite/deletion.

## Examples

```text
feat(agent): add recommendation action tools
feat(web): add result comparison view
fix(api): validate missing recommendation input
refactor(api): separate request mapping from service logic
style(web): improve mobile dashboard spacing
chore(ci): add web and api build checks
```
