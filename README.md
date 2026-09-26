# Project Name

One-sentence description of the product and its primary user.

## Core Features

- _____
- _____
- _____

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React + TypeScript |
| Backend | Go |
| Database | _____ |
| Auth | _____ |
| Local Runtime | Docker Compose |
| Deployment | Container-ready, provider TBD |

Architecture rationale lives in `guides/ARCHITECTURE.md`.

## Repository Structure

```text
apps/
  web/      # React + TypeScript
  api/      # Go API
guides/     # project and agent rules
docker-compose.yml
.github/
  workflows/
AGENTS.md
README.md
```

## Local Setup

Docker is the reference setup. A clean machine should only need Docker + Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

Then open:

```text
Frontend: http://localhost:5173
API:      http://localhost:8080
Health:   http://localhost:8080/health
```

Stop the stack with:

```bash
docker compose down
```

Native Node/Go setup is optional and documented in `guides/DEPLOYMENT.md`.

## Environment Variables

Frontend example:

```env
VITE_API_URL=http://localhost:8080
```

Backend example:

```env
PORT=8080
APP_ENV=development
DATABASE_URL=
ALLOWED_ORIGIN=http://localhost:5173
```

Never commit real secrets.

## Runtime & Deployment

The current priority is a reproducible Docker-based local environment. Cloud hosting is selected only after the MVP is stable.

Container setup and future deployment rules: `guides/DEPLOYMENT.md`.

## Project Guides

| File | Scope |
|---|---|
| `guides/PRD.md` | discovery, scope, P0-P3 prioritization, phased implementation plan |
| `guides/ARCHITECTURE.md` | stack, boundaries, data flow, structure |
| `guides/CLEAN_CODE.md` | implementation quality rules |
| `guides/DESIGN_SYSTEM.md` | reusable UI component structure |
| `guides/DESIGN_DIRECTION.md` | project-specific visual direction |
| `guides/SECURITY.md` | security floor |
| `guides/DEPLOYMENT.md` | Docker-first local runtime, env, CI/CD, deployment portability |
| `guides/GIT_CONVENTION.md` | commit format |
