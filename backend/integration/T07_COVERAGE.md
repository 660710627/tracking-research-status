# T-07 — POST multipart and create E2E

## Expected internal API

Tests use T-05 `db.Migrate`, `repo.NewResearchRepository`, T-06 `service.NewFileContractStore` and `service.NewResearchService`. New composition option: `handler.WithResearchCreator(creator)` accepted by `handler.NewRouter(health, ...options)`. The creator is the T-06 research service. The existing single-argument NewRouter health contract must continue working. The option registers only the existing POST `/api/v1/researches`; no upload endpoint or other feature is introduced. Production implementation is not part of T-07.

Every integration subtest creates one SQLite database in `t.TempDir()` and a separate temporary storage directory. Tests exercise real planned service/repo/storage rather than success mocks. Fixture SQL is limited to failure injection and inspection. Tests from T-05/T-06 are not edited.

## Mapping

| Requirement | Tests |
|---|---|
| AC-01 complete multipart → 201 JSON, every field, members, contract metadata, initial status/process, persisted record/PDF | POSTCreateSuccess |
| AC-02 internal server ID, reject ID/status/process and storage/server fields supplied by client | POSTCreateUnknownParts |
| AC-03 parent missing 404, duplicate root title/contract 409, continuation ID preserved and terminal parent accepted | POSTCreateBusinessErrors, POSTCreateContinuation |
| AC-04 required parts each once; bool/null/decimal types; no partial writes | POSTCreateMissingAndDuplicateParts, POSTCreateScalars |
| AC-05 exactly one member JSON array, malformed/non-array/trailing JSON, duplicate/unknown member keys and wrong member types | POSTCreateMembers |
| AC-06 PDF bytes and media, size 20 MiB and whole-body 21 MiB inclusive/one-byte-over; unknown content length; no published/staging files after rejected create | POSTCreatePDF, POSTCreateBodyLimit, POSTCreateBodyBoundary |
| Multipart media type case-insensitive, quoted boundary, member JSON part with/without browser Blob filename | POSTCreateEncodings |
| AC-13 malformed boundary, wrong media 415, request query 422, validation field errors, internal error suppression 500 | POSTCreateTransport, POSTCreateDatabaseFailure |
| AC-01/02/03/04/05/06/14 UI create and cross-cutting states | frontend/e2e/T07_CREATE.md |

T-05/T-06 cover detailed business validation matrices and atomic persistence. These integration tests add wire-format decoding/error mapping and real full-stack create; E2E cases repeat the validation rules from the user's perspective. Runtime assertions must be verified after T-10 and E2E after T-15 before closing the create slice.

## Results

- `cd backend; go test ./...` → exit 1 / RED. T-07 `internal/handler_test` reports missing `handler.WithResearchCreator` and the future router option signature, plus missing T-05/T-06 migration/repo/service/storage APIs. No handler test syntax errors reported. Existing T-05 repo and T-06 service packages also remain RED. Runtime assertions have not executed; TASKS permits RED from missing planned production symbols.
- `gofmt -l backend/internal/handler/research_create_test.go` (repository root) → no output after formatting.
- `rg -n '\bfetch\s*\(' frontend/src -g '*.tsx'` → no matches; current React entry/page component does not call fetch directly.
- Frontend server started successfully; in-app browser showed the placeholder heading/text at both 375×667 and 1280×720. No create control/form exists. Keyboard Tab did not reveal create controls. Screenshot and accessibility evidence appears in this task's browser tool results; textual evidence and all 29 scenarios are in `frontend/e2e/T07_CREATE.md`.
- E2E RED is confirmed at entry into the missing create flow. Subsequent loading/error/empty/success/submit/cancel assertions remain unexecuted until UI implementation; no GREEN claim is made. Initial connection-refused and npm argument errors were resolved before observation and are not RED evidence.

Only one new Go test file and two test documents were added. No production implementation, previous test edit, real database mutation, or uploaded-file change was made.
