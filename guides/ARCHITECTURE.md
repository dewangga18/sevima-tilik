# Architecture Guide

Purpose: keep a React + TypeScript frontend and Go backend simple, independent, and hackathon-friendly inside one monorepo.

## Default Stack

| Layer | Default | Notes |
|---|---|---|
| Frontend | React + TypeScript | Prefer Vite unless framework features are required |
| Backend | Go | HTTP API |
| Database | _____ | Choose based on confirmed PRD |
| Auth | _____ | Choose based on confirmed PRD |
| Runtime | Docker Compose | Local-first reference environment |
| Deployment | Container-ready | Provider selected later |

Do not add infrastructure merely for architectural purity. Optimize for a reliable MVP and clear boundaries.

## Monorepo Boundary

```text
apps/
  web/
    src/
      components/
      features/
      hooks/
      services/
      types/
      utils/
    Dockerfile
  api/
    cmd/
      server/
    internal/
      config/
      domain/
      handler/
      repository/
      service/
    Dockerfile
docker-compose.yml
.github/
  workflows/
```

Rules:

- `apps/web` must build independently without needing `apps/api` source files.
- `apps/api` must build independently without needing `apps/web` source files.
- `docker compose up --build` is the reference local setup.
- Communication between apps happens through an API contract, not filesystem imports.
- Do not create a repository-level `tests/` directory by default.
- If tests are introduced, colocate them with the package/module they verify unless the framework strongly requires otherwise.

## Frontend Responsibilities

Frontend owns:

- UI and interaction state
- Client-side validation for UX
- API requests
- Presentation formatting
- Non-sensitive local state

Frontend must not own secrets or authoritative business/security validation.

## Backend Responsibilities

Backend owns:

- Authoritative validation
- Business rules
- Authentication/authorization enforcement
- Database/external-service access
- Secret-bearing integrations
- AI-agent tool execution when actions require trusted credentials

## Go Layer Direction

Recommended dependency flow:

```text
handler -> service -> repository
              |
            domain
```

- `handler`: HTTP transport, request/response mapping
- `service`: use cases and business logic
- `repository`: storage/integration implementation
- `domain`: core models/contracts
- `config`: runtime configuration only

Do not create interfaces for every type by habit. Add abstraction when it improves boundary clarity, replacement, or testability.

## Data Flow

```text
Browser
  -> React app
  -> Go API
      -> Database
      -> External APIs
      -> AI model / agent tools (if applicable)
```

## AI-Agent Placement

If the product contains a functional agent:

- Keep agent orchestration in backend application/service logic when actions require secrets or trusted execution.
- Define each agent action/tool explicitly.
- Validate tool inputs before execution.
- Require user confirmation for destructive or high-impact actions where appropriate.
- Return observable action results so the agent can decide whether to continue or stop.

## Runtime Configuration

Frontend receives only public configuration such as API base URL.

Backend receives secrets and infrastructure configuration through environment variables.

Never hardcode environment-specific URLs into application logic. Use environment variables so the same containers can run locally or on a future deployment target.

## Decisions Log

| Decision | Reason |
|---|---|
| React + TypeScript frontend | _____ |
| Go backend | _____ |
| Database | _____ |
| Auth | _____ |
| AI provider/agent approach | _____ |

## Out of Scope

- _____
