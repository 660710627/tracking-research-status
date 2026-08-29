# PLAN: Tracking Research Status MVP

## Scope

- Research records contain project data, members, one funding-contract PDF, continuation, status, and process. `description` is not used.
- Create/update use `multipart/form-data`; status/process changes remain JSON. `docs/openapi.yaml` is the API source of truth.

## Architecture

- **Backend:** Gin handler → service → repo → SQLite. Handlers know HTTP only; services own business and file-lifecycle rules; repositories own SQL and transactions only.
- **SQLite:** enforce identity, foreign keys, continuation/title rules, enums, precision, transition/terminal rules, and one contract per research with constraints/triggers. Enable foreign keys on every connection.
- **Files:** validate and stage the PDF; publish it with contract metadata only after the transaction succeeds. Clean staged files on failure and remove stored files on deletion.
- **Frontend:** React pages → components → generated typed client. Components never call `fetch`; selected PDFs remain in browser memory until successful submit.

## Errors and testing

- Repo typed errors → service domain errors → handler HTTP status/error code exactly as SPEC.
- Use `go test`, `httptest`, and one SQLite database in `t.TempDir()` per test.
- Cover multipart/PDF lifecycle, project-member and funding rules, continuation, CRUD, transitions, and concurrent mutations.
- UI checks cover loading, error, empty, success, validation, cancellation, keyboard/focus, contrast, and responsive behavior.

## Delivery order

1. Keep SPEC and OpenAPI aligned; regenerate the typed client.
2. Write failing backend tests, then implement schema/repo/service/handler per vertical slice.
3. Build UI flows through the generated client.
4. Run local quality gates, CI, and independent review.
