# T-06 — create validation and PDF orchestration

## Internal API expected by tests (before implementation)

- `service.CreateResearchInput`: JSON-tagged request fields matching OpenAPI. Text/enums/dates are strings, `IsSubsidized *bool`, `ContinuationOfID *int64`, `ContinuationOfIDPresent bool` (not serialized), `BudgetAmount json.Number`, `ProjectMembers []service.ResearchMemberInput` with decimal `ContributionPercent json.Number`; `ContractFile *service.ContractUpload` (not serialized). No ID/status/process fields.
- `service.ContractUpload{Filename string, Reader io.Reader}`; no HTTP or multipart types. Handler owns malformed scalar decoding and duplicate/unknown parts in T-07. Presence flag distinguishes required literal null from an absent parent part.
- `service.NewResearchService(creator, store)`; creator has `Create(context.Context, repo.CreateResearchInput) (int64,error)` from T-05. `Create(context.Context, service.CreateResearchInput)` returns a JSON-serializable Research and error.
- Store interface: `Stage(context.Context,string,io.Reader) (service.StagedContract,error)`, `Publish(context.Context,service.StagedContract) (service.PublishedContract,error)`, `Discard(context.Context,service.StagedContract) error`, `Remove(context.Context,service.PublishedContract) error`.
- `StagedContract{Token, Filename string; SizeBytes int64}` and `PublishedContract{Path, Filename string; SizeBytes int64}` are internal only. Stage validates PDF and actual byte limit, cleans partial staging before returning error. Publish failures leave the staged token discardable. Remove failure must remain discoverable for startup reconciliation.
- `service.NewFileContractStore(root string)` returns store and error. Managed directories `staging/` and `contracts/` under root. `Reconcile(ctx, referencedPaths []string) error` runs before serving traffic, clears upload staging and unreferenced published files, preserves referenced files, rejects incomplete referenced state. Reference paths come from authoritative metadata, not requests. Loading that snapshot from SQLite belongs to composition, not the file adapter.
- Typed errors implement `ErrorCode() string`; validation errors additionally implement `FieldErrors() []service.FieldError` where FieldError has Field/Message. Client-facing error strings must not expose storage/SQL details. Existing repo errors map to matching contract codes.

Service converts BE dates to T-05 canonical Gregorian dates, validates decimals without silently rounding and normalizes text/contract key before calling repo. Publish occurs before repo.Create commits; failure compensates by deleting the new published file. No change to T-05 API or HTTP endpoints.

## AC mapping (create service/storage scope)

| Requirement | Tests |
|---|---|
| AC-01 complete create, initial status/process, ID and normalized values in response | CreateServiceSuccess |
| AC-03 required null/positive parent and kind relationship; parent/conflict errors, Unicode trimmed title and case preserved | CreateServiceValidation, CreateServiceSuccess, CreateServiceFailures |
| AC-04 every required string, Unicode whitespace, 1/1000/1001 code points, title slash/control, general control vs newline/tab | CreateServiceText |
| AC-04 enums, boolean presence/false/true, parent null/presence | CreateServiceValidation, CreateServiceSuccess, CreateServiceValidValues |
| AC-04 real BE dates and format, minimum year, leap-day anniversary | CreateServiceValidation, CreateServiceValidDates |
| AC-04 budget >0, two decimal places, no rounding, invalid/nonfinite values | CreateServiceValidation, CreateServiceValidValues |
| AC-05 both members' name/email/affiliation, role count, distinct trimmed case-insensitive emails, individual precision/range, total 100 | CreateServiceText, CreateServiceValidation, CreateServiceSuccess |
| AC-04/05/13 field-path errors, invalid input never stages or persists, input retained | CreateServiceValidation, CreateServiceText |
| AC-06 required real PDF, 1+ pages, reject corrupt/truncated/encrypted/empty/header-only, content not filename, actual 20 MiB boundary | FileContractValidation |
| AC-06 no upload on selection alone/cancel before submit | FileContractLifecycle (constructor idle); actual browser cancel remains T-07 E2E |
| AC-06 publish before commit, do not report success until both ready, stage/publish/repo/cleanup/read failures, cancellation | CreateServiceSuccess, CreateServiceFailures, CreateServiceWaitsForCommit, FileContractReadFailure, CreateServiceCanceledAfterPublish |
| AC-06 storage PDF errors retain validation/payload code through service and never call repo | CreateServicePDFErrors |
| AC-06/13 response/error never exposes internal path/key or SQL, no retries of mutation | CreateServiceSuccess, CreateServiceFailures |
| AC-06/SQL cleanup after crash: uncommitted published orphan, stale upload, retain committed PDF; repeat recovery and missing reference failures | FileContractRecovery |
| AC-06 generated storage path; no overwrite from duplicate client filenames; metadata filename constraints | FileContractLifecycle, FileContractFilename |

PDF bytes are built as test fixtures only; no production feature/stub is introduced. Storage tests use `t.TempDir()`. Service tests use in-memory test doubles and do not access SQLite; real repo persistence/atomicity is T-05. Multipart 21 MiB aggregate limit and HTTP scalar parsing are T-07. Update/delete recovery and old-file retention are tested in their own scheduled slices.

## Verification

- `cd backend; go test ./...` → exit 1 / RED. New `internal/service_test` fails compilation because `service.StagedContract`, `service.PublishedContract`, `service.CreateResearchInput` and related planned APIs are not implemented. Prior T-05 `internal/repo_test` also remains RED (`repo.CreateResearchInput`, `repo.ResearchRepository`, `db.Migrate`, etc.). Health handler tests PASS.
- `go test '-gcflags=github.com/660710627/my-research/internal/service=-e' ./internal/service` → exit 1. Expanded diagnostics identify missing planned repo/service/storage symbols, including `NewResearchService` and `NewFileContractStore`; no syntax error reported.
- `gofmt -w internal/service/research_create_test.go internal/service/contract_store_test.go` completed successfully.

Runtime assertions have not executed because implementation is absent. This is the missing-production-symbol RED explicitly permitted by TASKS, not a dependency or network failure. Only the two new test files and this coverage document are introduced by T-06; T-05 tests and production files are unchanged.
