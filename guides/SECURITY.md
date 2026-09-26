# Security Guide

Purpose: provide a practical security floor for a hackathon web app without pretending the MVP is production-hardened.

## Trust Boundary

Treat all browser/client input as untrusted.

Authoritative validation and permission checks belong on the Go backend.

## Input and Output

- Validate request shape, type, range, length, and allowed values server-side.
- Use parameterized queries / safe ORM APIs; never concatenate raw SQL with user input.
- Avoid rendering untrusted HTML. If HTML is required, sanitize it explicitly.
- Validate file type and size; do not trust file names or MIME headers alone.

## Auth and Authorization

- Never store plaintext passwords.
- Prefer a proven identity provider when auth is not the product's core innovation.
- Enforce authorization on backend resources/actions, not only in UI visibility.
- Use secure cookie settings when cookie-based sessions are used.
- Apply CSRF protection when the chosen auth/session model requires it.

## Secrets

- Never commit `.env`, API tokens, private keys, credentials, or service-account secrets.
- Client-side `VITE_*` variables are public after build; never put secrets there.
- Secret-bearing AI/API integrations should run through the backend.

## API Protection

- Use HTTPS in production.
- Configure CORS to the known frontend origin(s).
- Add rate limiting to abuse-sensitive endpoints such as auth, expensive AI requests, or public write operations when relevant.
- Return safe client errors; do not expose stack traces, SQL errors, or internal configuration.

## AI-Agent Actions

When an AI agent can execute actions:

- Validate every tool/action input before execution.
- Apply the same authorization checks as non-AI requests.
- Do not let model-generated arguments bypass server rules.
- Require explicit user confirmation for destructive/high-impact actions where appropriate.
- Avoid sending unnecessary personal/sensitive data to external models.

## Before Submission

- [ ] No secrets are committed
- [ ] Production CORS is restricted appropriately
- [ ] Backend validates important inputs
- [ ] Protected actions enforce authorization
- [ ] Error responses do not leak internals
- [ ] Expensive/abusable endpoints have reasonable safeguards where needed
