# Clean Code Guide

Purpose: keep implementation readable and easy to change under hackathon time pressure.

## General Rule

Prefer boring, explicit code over clever abstraction.

## Functions

- One clear responsibility per function.
- Keep orchestration separate from low-level details.
- Prefer early returns over deeply nested conditionals.
- Side effects should be obvious from the function name.
- Extract functions when it improves meaning, reuse, or readability — not to satisfy an arbitrary line count.

## Naming

Use names that explain intent.

Good:

```text
getUserByID
calculateTotalPrice
validateEmail
createRecommendation
```

Boolean names should normally read naturally with:

```text
is...
has...
can...
should...
```

Avoid vague names such as:

```text
handleData
process
helper
doStuff
manager
```

Use common technical abbreviations only when they are unambiguous (`id`, `url`, `api`, `db`).

## Dependencies

- Pass infrastructure dependencies into services/components instead of hiding them in mutable global state.
- Keep configuration access near application setup.
- Avoid service-locator/singleton patterns when simple dependency injection is enough.
- Wrap non-deterministic inputs such as current time or randomness when deterministic behavior matters.

## Errors

- Never use empty catch/recover blocks.
- Add context when returning/wrapping backend errors.
- Do not expose internal stack traces or sensitive details to clients.
- UI errors must tell the user what happened and, when possible, what action they can take.

## Frontend

- Keep network calls out of presentational components.
- Prefer feature/domain hooks or service modules for remote operations.
- Keep derived state derived; do not duplicate it unnecessarily.
- Avoid large components that mix fetching, transformation, layout, and unrelated interactions.
- Use existing design-system components before adding new ones.

## Go Backend

- Pass `context.Context` through request-scoped operations that perform I/O.
- Return errors rather than logging-and-swallowing them in lower layers.
- Log at an application boundary where request/operation context is available.
- Keep HTTP concerns inside handlers; keep core business logic out of handlers.
- Keep database-specific details inside repository/integration code.

## Testability Without Test Theater

The project does not require tests merely to satisfy structure.

Still write code that can be tested later:

- Separate pure logic from I/O.
- Inject external dependencies at meaningful boundaries.
- Avoid hidden global mutable state.
- Keep business rules callable independently of HTTP/UI code.

If tests are added, place them near the code they verify rather than creating a repository-level `tests/` folder by default.

## Documentation

Comments should explain **why**, constraints, or non-obvious behavior.

Do not comment code that already explains itself.
