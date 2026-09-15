# T-11 list researches backend test coverage

## Expected typed production API

- `(*repo.ResearchRepository).List(context.Context) ([]repo.Research, error)` returns a non-nil empty slice when no rows exist. `repo.Research`, `repo.ResearchMember`, and `repo.ResearchContract` contain only values needed by the list response; contract storage paths, normalized keys, and file bytes are excluded.
- `service.NewResearchListService(repository).List(context.Context) ([]service.Research, error)` maps Gregorian stored dates to Buddhist `DD/MM/YYYY`, preserves every public field, and maps repository failures to `INTERNAL_ERROR` without leaking their text.
- `handler.WithResearchLister(...)` installs `GET /api/v1/researches` on `handler.NewRouter`.

These APIs are intentionally referenced by the tests before they exist. The compile failure is the RED state for T-11 and defines the interface T-12 must implement.

## Acceptance-criteria mapping

| Scope | Tests |
| --- | --- |
| AC-07 complete fields, two members, contract metadata, no binary/path/description | `TestListResearchesReturnsCompleteAggregatesSortedByTitleThenID`, `TestResearchListServiceFormatsDatesAndPreservesCompleteData`, `TestGETResearchesReturnsArrayWithContractShapeAndOrder` |
| AC-07 title ascending, ID ascending for duplicate titles | `TestListResearchesReturnsCompleteAggregatesSortedByTitleThenID`, `TestGETResearchesReturnsArrayWithContractShapeAndOrder` |
| AC-07 empty database is `[]` | `TestListResearchesEmptyAndDatabaseFailure/empty`, `TestGETResearchesEmptyIsJSONArray` |
| AC-13 non-empty GET body is 400 | `TestGETResearchesRejectsBodyAndQuery/body` |
| AC-13 unsupported query is 422 with `fieldErrors` | `TestGETResearchesRejectsBodyAndQuery/query` |
| AC-13 database failures are sanitized 500 errors | `TestListResearchesEmptyAndDatabaseFailure/database_failure`, `TestResearchListServiceSanitizesRepositoryFailure`, `TestGETResearchesSanitizesDatabaseFailure` |

No test is skipped and no production file is changed by T-11.

## RED result

Command: `cd backend; go test ./...`

Result: RED (`exit 1`) because the target production API does not exist yet:

- `(*repo.ResearchRepository).List` is undefined.
- `repo.Research`, `repo.ResearchMember`, and `repo.ResearchContract` are undefined.
- `service.NewResearchListService` is undefined.
- `handler.WithResearchLister` is undefined.

The existing `cmd/server` and `internal/db` packages compiled; the repo, service, and handler test packages failed to build only at the missing T-12 capability boundary.
