# T-02 health/global-routing test contract

## Expected production boundary for T-03

- `handler.HealthChecker`: interface with `CheckHealth(context.Context) error`.
- `handler.NewRouter(HealthChecker) http.Handler`: builds the HTTP router.
- Typed service errors expose `ErrorCode() string`; `SERVICE_UNAVAILABLE` maps to 503. Unclassified failures map to 500.

These symbols describe the target production API only. T-02 does not provide production implementations.

## AC-13 coverage mapping

| AC-13 / T-02 condition | Automated test |
|---|---|
| Healthy service and reachable isolated SQLite return 200 JSON `{"status":"ok"}` | `TestHealthReturnsOKWhenDatabaseIsAvailable` |
| Database unavailable returns 503 with `SERVICE_UNAVAILABLE` | `TestHealthReturnsServiceUnavailableWhenDatabaseIsUnavailable` |
| Unexpected health failure returns 500 with `INTERNAL_ERROR` | `TestHealthReturnsInternalErrorWithoutLeakingDetails` |
| Error responses use exactly the JSON error envelope with a non-empty message | `requireErrorResponse` is applied to every 4xx/5xx case |
| Database paths, SQL details, and wrapped internal causes are not exposed | Both health failure tests use secret/path-bearing causes and assert their absence |
| Unknown route returns JSON 404 `ROUTE_NOT_FOUND` without invoking health | `TestUnknownRouteReturnsJSONNotFoundWithoutCallingHealth` |
| Unsupported method on a known route returns JSON 405 `METHOD_NOT_ALLOWED` without invoking health | `TestWrongMethodReturnsJSONMethodNotAllowedWithoutCallingHealth` |
| Persistence tests use SQLite in `t.TempDir()` and never the runtime `library.db` | `openIsolatedSQLite` and `TestSQLiteHealthFixtureUsesOneTemporaryDatabase` |

All HTTP cases use `net/http/httptest`. No test is skipped and no runtime database or uploaded file is touched.
