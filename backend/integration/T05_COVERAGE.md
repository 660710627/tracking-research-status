# T-05: create persistence contract and coverage

## Expected production API (declared before writing tests)

- `db.Migrate(ctx context.Context, database *sql.DB) error`: explicitly install schema; opening the runtime database does not migrate or clear it.
- `repo.NewResearchRepository(database *sql.DB) *repo.ResearchRepository`.
- `(*repo.ResearchRepository).Create(ctx context.Context, input repo.CreateResearchInput) (int64, error)`: atomically persist a project, both members and one contract; return generated ID only after commit, zero on failure.
- `repo.CreateResearchInput`: Title, ResearchKind, ProjectType, ResponsibleProjectUnit, ResponsibleBudgetUnit, StartDate, EndDate, ThaiAbstract, EnglishAbstract, Objectives, Keywords (string); IsSubsidized (bool); ContinuationOfID (*int64); BudgetAmount (float64); Members ([]repo.ResearchMemberInput); Contract (repo.ResearchContractInput). No ID/status/process input.
- `repo.ResearchMemberInput`: FullName, Email, Affiliation, Role (string), ContributionPercent (float64).
- `repo.ResearchContractInput`: FundingType, FundingSourceName, ContractNumber, ContractNumberKey, StoragePath, OriginalFilename, ContentType (string), SizeBytes (int64).
- `repo.ErrContinuationNotFound`, `repo.ErrTitleAlreadyExists`, `repo.ErrContractNumberAlreadyExists`, `repo.ErrConstraintViolation`, `repo.ErrInternal`: errors identifiable with `errors.Is`.

Internal persistence dates are canonical Gregorian `YYYY-MM-DD`; amounts are decimal baht and percentages, never silently rounded. Service owns BE date conversion and Unicode normalization; fixtures supply already-normalized titles/emails and contract keys. Database still enforces constraints. These are internal APIs, not new HTTP endpoints. Relation columns in tests are `research_id`.

## Scope mapping

| AC / individual persistence requirement | Test |
|---|---|
| AC-01 all project fields, both personnel records, funding and single contract metadata, initial status/process | CreateCompleteAggregate |
| AC-02 positive/generated unique ID, no caller ID/status/process; no reuse after deleting highest ID | CreateIdentity |
| AC-02 ID immutable when any SQL writer attempts change | CreateImmutableAndForeignKeys |
| AC-02 simultaneous creates keep unique IDs | CreateConcurrent/distinct |
| AC-03 root has null parent, continuation has existing parent, missing parent typed error | CreateContinuation, CreateRejectedAggregate |
| AC-03 completed/terminated parent accepted, multiple children, unrelated or duplicated child title | CreateContinuation |
| AC-03 root duplicate against root or child, Unicode-trimmed/case-sensitive comparison | CreateTitleRules |
| AC-03 immutable parent, self-FK restrict update/delete | CreateImmutableAndForeignKeys |
| AC-03/13 concurrent duplicate root accepts one; concurrent children allowed | CreateConcurrent |
| AC-04 required/length/control/title rules at SQL boundary, enums, boolean, valid dates/minimum year, budget positivity/precision | CreateSQLConstraints |
| AC-05 separate members, all fields, exactly two roles, distinct normalized emails, positive percentages and total 100, invalid member rolls back aggregate | CreateCompleteAggregate, CreateRejectedAggregate, CreateSQLConstraints |
| AC-06 contract metadata one-to-one, PDF type/size constraints; no partial metadata on error | CreateSQLConstraints, CreateRollback |
| AC-13 atomic project/member/contract write on SQL failure including COMMIT, canceled/closed DB errors | CreateRollback, CreateRollbackAtCommit, CreateUnavailable |
| AC-08 uniqueness is also a create invariant: normalized contract key, concurrent duplicate contract | CreateContractUniqueness, CreateConcurrent |

UI, HTTP parsing/error envelopes, raw input normalization, PDF parsing/filesystem publish/cleanup and startup recovery belong to the already scheduled T-06/T-07 tests, not this repo-only scope. Status/process transition matrices remain in their own slices; T-05 only uses terminal SQL fixtures to verify continuation creation. No list/update/delete production methods are required by these tests.

## Verification

Verification:

- `cd backend; go test ./...` → exit 1, RED in `internal/repo_test`: undefined `repo.ResearchRepository`, `db.Migrate`, `repo.NewResearchRepository`, `repo.CreateResearchInput`, `repo.ResearchMemberInput`, `repo.ResearchContractInput`. Handler health tests remain PASS.
- `go test '-gcflags=github.com/660710627/my-research/internal/repo=-e' ./...` → exit 1; expanded diagnostics report the planned missing production symbols/types/errors, without test syntax errors. This also covers the final commit-failure test. The flag requires quoting in PowerShell.
- `gofmt -l` on the initial two new test files produced no output; `gofmt -w` was also applied after adding the commit-failure test.
- Initial sandbox execution could not access Go standard library/build cache; rerunning with approved access produced the feature-related RED above. The access failure is not counted as RED evidence.

Tests have not reached runtime assertions: the RED is compilation against the documented future internal API, explicitly permitted by TASKS. Each test/subtest provisions one SQLite database in its own `t.TempDir()` once implementation exists; no real PDF or database is touched. No production implementation, stub, skipped test, or existing test modification was added.
