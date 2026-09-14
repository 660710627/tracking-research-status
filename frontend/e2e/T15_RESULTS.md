# T-15 verification — implementation complete, final E2E blocked

Run date: 2026-09-14. Desktop-only scope follows the user's decision and docs/PLAN.md.

## Environment

- Frontend: Vite on http://127.0.0.1:5173.
- Backend: real handler/service/repository/SQLite stack through go run ./integration/t15 on port 8080.
- Isolated data root: C:\Users\Lenovo\AppData\Local\Temp\research-t15-3143642796; the application library.db was not used.
- Generated PDF fixtures cover valid, invalid, encrypted, password-protected, zero-page and exact size boundaries.
- The browser transport harness forwards normal requests to the real backend. Delay/network/500 modes add no production endpoint.

## Browser and persistence evidence

- E01-E07: PASS. Complete form/actions and labels; one multipart POST returned 201 and survived reload; literal null for BUDGET; internal parent ID for continuation; initial status/process correct; duplicate submit blocked; 500/network errors retained values/file without leaking internals.
- E08: PARTIAL. Same-name, different-name and multiple-child continuations passed against the real backend. Completed/terminated parent variants still require isolated fixture mutation.
- E09-E10: PASS. Missing selected parent returned 404 CONTINUATION_NOT_FOUND, retained the form, cleared only the stale choice and created no child. Duplicate title/contract behavior and normalization produced the expected 409 field errors.
- E11-E19: PASS. Required fields, Unicode/code-point limits, forbidden controls, invalid enums, Buddhist-calendar boundaries, budget decimals, member shape/email and contribution boundaries were exercised. Invalid values caused no POST or real-backend 422; valid boundaries passed.
- E20-E22: PASS. Selection alone caused no upload; cancel confirmation caused no POST. Empty/fake/malformed/zero-page/encrypted/password PDFs failed at contractFile. Exactly 20 MiB returned 201; one-byte-over returned 413 and left the form usable.
- E23: UI PASS / real failure injection blocked. Simulated 500 displayed no success and retained the form. Passing backend tests cover persistence compensation and atomicity. A temporary database trigger for the final browser-to-real-backend repetition was not authorized.
- E24-E26: PASS. Pristine and dirty cancel/back behavior, confirmation focus return, main/alert focus, labels, tab order and visible focus rule passed.
- E25: PASS. Parent loading, error/retry and empty states are distinct and never select a fallback ID.
- E27-E28: PASS for desktop at 1280x720. Document/main scroll and client widths were 1265/1265 and 1045/1045. Measured text contrast was 6.91:1 to 12.1:1. Reduced motion disables animation, transitions and smooth scrolling.
- E29: PASS. Unicode-trimmed/case-folded duplicate contract number returned 409, marked the field and retained the form.

## Commands

- cd frontend; npm run lint — PASS, exit 0.
- cd frontend; npm run build — PASS, exit 0; TypeScript and Vite completed, 41 modules transformed.
- cd backend; go test ./... — PASS, exit 0 after build-cache access; handler, repo and service passed.
- rg for direct fetch in frontend/src/pages and frontend/src/components — no matches.

## T-15 files

- frontend/src/pages/CreateResearchPage.tsx: form, validation, generated-client multipart submit, continuation selection, error retention, pending state, focus and dirty confirmation.
- frontend/src/App.tsx: list/create transition and creation feedback.
- frontend/src/pages/ResearchListPage.tsx: create entry action.
- frontend/src/index.css: form, error, focus, desktop layout and reduced-motion styles.
- frontend/e2e/T15.html and frontend/e2e/T15.tsx: test-only transport harness.
- backend/integration/t15/main.go, commands.go and pdf_fixtures.go: isolated real-backend runner, audit/fault controls and PDF fixtures.
- frontend/e2e/T15_RESULTS.md: this run evidence.

No *_test.go file was modified for T-15. Other dirty worktree entries predate this task and are not claimed here.
