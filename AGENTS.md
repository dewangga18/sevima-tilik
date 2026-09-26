# Agent Rules

This repository uses small, focused guides. Read only the guides relevant to the current task, but always respect the global rules below.

## Required Reading Order

Before implementation:

1. `guides/PRD.md` — discover missing requirements, confirm the product direction, prioritize scope, and create the implementation plan.
   Product-specific scope lives in `docs/PRODUCT.md`; execution status and phase checkpoints live in `docs/IMPLEMENTATION_PLAN.md`. Read both before implementation; do not treat the PRD guide as a filled product PRD.
2. `guides/ARCHITECTURE.md` — respect the agreed stack and boundaries.
3. `guides/CLEAN_CODE.md` — apply implementation rules.
4. For UI work: `guides/DESIGN_SYSTEM.md` + `guides/DESIGN_DIRECTION.md`.
5. For auth, input, storage, API, AI tools/actions, or secrets: `guides/SECURITY.md`.
   For API work, also read `docs/API_CONTRACT.md` and update it in the same change whenever an endpoint or its behavior changes.
6. For local runtime/Docker/config/CI/deployment portability: `guides/DEPLOYMENT.md`.
7. For commits: `guides/GIT_CONVENTION.md`.

## Hackathon Execution Rules

- Do not code before the PRD direction is explicitly confirmed and a prioritized implementation plan exists.
- Execute the highest incomplete priority first: P0 -> P1 -> P2 -> P3.
- Prefer vertical slices that produce observable end-to-end behavior.
- Keep the app runnable and demoable after every completed phase.
- Complete the current phase's demo checkpoint before starting lower-priority polish.
- If time becomes constrained, remove optional scope before weakening the core demo path.
- Re-plan when requirements materially change instead of silently expanding scope.

## Must

- Prefer the smallest implementation that satisfies the confirmed MVP.
- Preserve the monorepo boundary: frontend and backend must remain independently buildable and containerizable.
- Reuse existing components, patterns, utilities, and dependencies before creating new ones.
- Ask before introducing a new dependency, architectural pattern, or major rewrite.
- Keep secrets in environment variables only.
- Surface errors explicitly; never silently swallow failures.
- Before adding or changing an API endpoint, define its contract in `docs/API_CONTRACT.md`; reconcile the document with the final implementation before committing. Include method/path, auth/role/environment restrictions, request/response fields, status codes, safe errors, and retry/idempotency behavior. Endpoint work is not complete until its contract is updated in the same commit.
- Keep changes scoped to the requested task and the active implementation phase.

## Must Not

- Do not invent requirements when the PRD is ambiguous.
- Do not turn nice-to-have features into MVP requirements without confirmation.
- Do not start P3 polish while critical P0/P1 items remain incomplete.
- Do not finish an entire backend or frontend layer in isolation when a smaller vertical slice can validate the product sooner.
- Do not hardcode production URLs, secrets, colors, spacing, or typography values when configuration/tokens exist.
- Do not create duplicate UI components for minor variants.
- Do not create a repository-level `tests/` folder by default.
- Do not delete or rewrite working code without a clear reason and explicit approval for large changes.

## Output Style

- Lead with the result or direct answer.
- Keep explanations compact unless detail is requested.
- State the current phase/priority when reporting implementation progress.
- State blockers and the smallest viable fallback when one exists.
- Prefer concrete file names, commands, and decisions over generic advice.
