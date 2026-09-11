# T-14 verification

## Run

Latest user scope override: desktop only, acceptance viewport 1280×720. The mobile observations below are historical evidence from before this decision, not a supported layout or current acceptance requirement. The mobile-specific CSS breakpoint has been removed; desktop CSS is unchanged. SPEC and PLAN now record this decision. The original T13 checklist is preserved as historical test input; its mobile requirements are superseded by the user's instruction.

- From frontend: npm run lint: PASS (ESLint, 2026-09-11).
- From frontend: npm run build: PASS (TypeScript + Vite 8.1.5, 40 modules, built in 430 ms).
- Windows used the installed npm.cmd at C:/Program Files/nodejs/npm.cmd.
- Start: node node_modules/vite/bin/vite.js --host=127.0.0.1 --port=5173 --strictPort.
- Browser: http://127.0.0.1:5173/e2e/T14.html, in-app browser, actual rendered App.
- Harness replaces only the generated client's transport with T13 fixture Responses. Production entry does not import harness or fixtures. No real mutation is performed.

## T13 case evidence

| Cases | Result / observation |
| --- | --- |
| L01 | PASS: transport log GET /api/v1/researches query= body=null |
| L02 | PASS: loading holds until release, refresh disabled; release shows 3 rows |
| L03–L04 | PASS: HTTP 500 and rejected network transport show โหลดรายการไม่สำเร็จ and ลองใหม่; no empty message |
| L05 | PASS: set retry-success then click/Enter ลองใหม่; loading followed by 3 rows |
| L06 | PASS: empty fixture shows ทั้งหมด 0 รายการ and ยังไม่มีงานวิจัย |
| L07–L08 | PASS: five prescribed columns; CN-101, CN-108, CN-120 in order, duplicate โครงการ ก preserved |
| L09 | PASS: source uses key=item.id, no displayed ID or ID input; changes selected by id |
| L10 | PASS by source/data-flow review: complete typed Research[] retained without projection in page state; table uses prescribed fields from each record. Members, dates, budget, metadata are retained in objects, not added as unrequested columns |
| L11 | PASS: created fixture shows 4 rows and เพิ่มโครงการสำเร็จ: CN-125; only CN-125 marked เปลี่ยนแปลงล่าสุด |
| L12 | PASS: updated fixture shows 3 rows, อัปเดตโครงการสำเร็จ: CN-108; only CN-108 marked despite duplicate name |
| L13 | PASS: refresh after updated response issues another GET; 3 data rows remain, no duplicated stale rows |
| L14 | PASS: navigation link focuses main; Space collapses sidebar, Enter expands; aria-expanded and current-page link present |
| L15 | PASS: screenshots inspected at 375×667 and 1280×720; mobile table scrolls inside its region; no document horizontal overflow |
| L16 | PASS: Tab from main reaches refresh then table region; Right scrolls table; retry uses Enter; menu uses Space/Enter. Visible 3px focus outline. No dialog/dropdown in this slice |
| L17 | PASS: computed colors give body/helper 6.91:1, normal cells 12.10:1, headers 10.48:1, changed row/feedback 10.06:1. Error foreground rgb(120,43,32), background rgb(255,243,240). Focus uses #174d40 with 3px outline and 4px offset |
| L18 | PASS by invariant: no animation or transition in any UI state, computed animationName=none and transitionDuration=0s. Reduced-motion CSS additionally disables both. No OS preference was changed or browser media emulation claimed |
| L19 | PASS: long-text fixture renders 30 rows; document scrollWidth 360 at viewport 375, 1265 at viewport 1280 (vertical scrollbar accounts for difference); text wraps inside table cells |
| L20 | PASS: no fetch calls in pages/components; api.listResearches delegates to generated SDK. Generated files unchanged |

Screenshots and accessibility trees were observed in browser tool output. These are manual E2E observations plus the explicitly identified source reviews, not an automated test-suite pass.

## Integration contract

App accepts optional change={id, kind:'created'|'updated', revision}. A successful mutation flow should increment revision and pass the affected internal ID. App remounts the list for that revision, reloads through generated client, then shows feedback only when that record appears in successful data. No mutation forms were implemented.

## Scope

Production changes: src/App.tsx, src/index.css, src/pages/ResearchListPage.tsx, src/components/ResearchTable.tsx.
Test support: e2e/T14.html, e2e/T14.tsx and this record. Original T13 checklist/fixtures and all *_test.go are unchanged in T14.
