# Local Runtime & Deployment Guide — Docker First

Purpose: make the React + TypeScript frontend and Go backend easy to run, share, and later deploy without requiring each developer to manually install matching Node and Go versions on the host machine.

For this project, Docker is the default runtime setup. Production hosting is intentionally left open until the MVP is stable.

## Runtime Model

```text
Docker Compose
|
|-- web  -> React + TypeScript dev server
|
`-- api  -> Go HTTP API
```

Both apps remain independently buildable, but local development should prefer one command from the repository root.

## Recommended Repository Shape

```text
apps/
  web/
    src/
    package.json
    package-lock.json
    Dockerfile
    .dockerignore
    .env.example
  api/
    cmd/
      server/
    internal/
    go.mod
    go.sum
    Dockerfile
    .dockerignore
    .env.example
docker-compose.yml
.env.example
.github/
  workflows/
```

## Local Prerequisite

Required on the host:

- Docker Desktop, Docker Engine, or another compatible Docker runtime
- Docker Compose v2 (`docker compose`)

Node.js and Go do not need to be installed on the host when using the Docker workflow.

## One-Command Local Setup

From the repository root:

```bash
cp .env.example .env
docker compose up --build
```

Expected local URLs:

```text
Frontend: http://localhost:5173
API:      http://localhost:8080
Health:   http://localhost:8080/health
```

Stop the stack with:

```bash
docker compose down
```

Rebuild after dependency or Dockerfile changes:

```bash
docker compose up --build
```

## Root Environment Example

Keep non-secret local defaults in `.env.example`:

```env
WEB_PORT=5173
API_PORT=8080
APP_ENV=development
ALLOWED_ORIGIN=http://localhost:5173
```

If database or external-provider credentials are required, add empty placeholders only.

Never commit real secrets.

## Docker Compose

Recommended baseline:

```yaml
services:
  web:
    build:
      context: ./apps/web
      target: development
    ports:
      - "${WEB_PORT:-5173}:5173"
    environment:
      VITE_API_URL: http://localhost:${API_PORT:-8080}
    volumes:
      - ./apps/web:/app
      - web_node_modules:/app/node_modules
    depends_on:
      api:
        condition: service_healthy

  api:
    build:
      context: ./apps/api
      target: development
    ports:
      - "${API_PORT:-8080}:8080"
    environment:
      PORT: 8080
      APP_ENV: ${APP_ENV:-development}
      ALLOWED_ORIGIN: ${ALLOWED_ORIGIN:-http://localhost:5173}
    volumes:
      - ./apps/api:/app
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 5s
      timeout: 2s
      retries: 10

volumes:
  web_node_modules:
```

Adjust environment variables when the confirmed PRD adds database, auth, AI providers, or other integrations.

## Frontend Dockerfile — React + TypeScript

Use a multi-stage file so development and production builds share the same dependency definition.

```dockerfile
FROM node:22-alpine AS base
WORKDIR /app
COPY package*.json ./
RUN npm ci

FROM base AS development
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

FROM base AS build
COPY . .
RUN npm run build

FROM nginx:alpine AS production
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
```

If the project uses another package manager, keep the same stage structure but replace the install/build commands consistently.

## Backend Dockerfile — Go

Use separate development and production targets.

```dockerfile
FROM golang:1.25-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM base AS development
COPY . .
EXPOSE 8080
CMD ["go", "run", "./cmd/server"]

FROM base AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.22 AS production
WORKDIR /app
COPY --from=build /server ./server
EXPOSE 8080
CMD ["./server"]
```

Pin versions to the versions actually used by the project before final submission.

## Docker Ignore Files

Frontend `.dockerignore`:

```text
node_modules
dist
.git
.env
```

Backend `.dockerignore`:

```text
.git
.env
bin
tmp
```

## Container Networking Rule

Inside Compose, containers communicate using service names, not `localhost`.

Example:

```text
web container -> http://api:8080
```

However, browser-side React code runs on the user's machine. Therefore `VITE_API_URL` exposed to the browser should normally use the host-visible address:

```text
http://localhost:8080
```

Do not accidentally put `http://api:8080` into browser runtime configuration.

## Go Server Binding

The API must bind to all interfaces inside the container.

```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}

addr := "0.0.0.0:" + port
log.Fatal(http.ListenAndServe(addr, router))
```

Do not bind only to `localhost` inside Docker.

## Health Endpoint

Provide:

```text
GET /health
```

It should be cheap and should not call AI providers or other expensive external services.

Example response:

```json
{"status":"ok"}
```

## Development Workflow

Normal startup:

```bash
docker compose up
```

Force image rebuild:

```bash
docker compose up --build
```

Run in background:

```bash
docker compose up -d
```

View logs:

```bash
docker compose logs -f
```

View one service:

```bash
docker compose logs -f api
```

Reset containers and volumes when local state is intentionally disposable:

```bash
docker compose down -v
```

Do not use `-v` casually once persistent local database volumes are introduced.

## Native Development Is Optional

Developers may still run either app natively when useful:

```bash
cd apps/web && npm ci && npm run dev
```

```bash
cd apps/api && go mod download && go run ./cmd/server
```

But Docker remains the reference setup used to verify that a clean machine can start the project reliably.

## CI Minimum

CI should prove both source builds and Docker builds remain valid.

```text
web:
  npm ci
  npm run build

api:
  go mod download
  go vet ./...
  go build ./...

docker:
  docker compose build
```

Add lint/typecheck/test steps only when those commands actually exist in the repository.

## Future Deployment

Do not couple the project to a specific cloud provider yet.

The production-ready Docker targets should make later deployment possible on platforms that can run containers. When hosting is selected, add provider-specific instructions only then.

Keep these assumptions portable:

- configuration through environment variables
- no required local filesystem persistence inside containers
- API binds to `0.0.0.0`
- health endpoint exists
- frontend and backend images build independently
- secrets are injected at runtime, not baked into images

## Done When

- [ ] `docker compose up --build` starts the full app from repository root
- [ ] frontend opens at `http://localhost:5173`
- [ ] API opens at `http://localhost:8080`
- [ ] `/health` becomes healthy
- [ ] frontend can call the API successfully
- [ ] frontend and backend Dockerfiles build independently
- [ ] source code changes are visible during local development
- [ ] no real secrets are committed or baked into images
- [ ] a clean machine only needs Docker to run the project
