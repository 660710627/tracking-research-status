# TASKS: Tracking Research Status MVP

ทำตามลำดับทีละหนึ่ง task เท่านั้น ทุก feature ต้องมี test task ที่แดงและครอบเงื่อนไขของ slice ครบก่อนเริ่ม implementation task

## Contract

### T-01 — ปรับ OpenAPI contract
- **สิ่งที่ทำ:** ปรับ `docs/openapi.yaml` ให้ตรงกับ API Contract, schema, validation, status และ error code ใน SPEC ล่าสุดเท่านั้น ครบ health, list, create, update by ID, delete by ID, status และ process โดยไม่เพิ่ม endpoint
- **Dependencies:** ไม่มี
- **DoD:** `npx @redocly/cli lint docs/openapi.yaml` ผ่าน; operation และ response ครบ AC-1–AC-10; schema ใช้ `id`, `continuationOfId`, `status`, `process` และ error format เดียวกัน; ยังไม่มีการแก้โค้ด backend/frontend

## Walking skeleton

### T-02 — เขียน tests สำหรับ health และ global routing
- **สิ่งที่ทำ:** เขียน handler/integration tests ตาม Contract สำหรับ health `200`, `503`, `500` พร้อม JSON response และ global `404 ROUTE_NOT_FOUND`, `405 METHOD_NOT_ALLOWED` รวมทั้งไม่เปิดเผย internal error ตาม AC-10
- **Dependencies:** T-01
- **DoD:** ทุก test ใช้ SQLite คนละฐานข้อมูลใน `t.TempDir()`; `cd backend; go test ./...` แดงเพราะ walking skeleton ยังไม่มี implementation เท่านั้น และทุกเงื่อนไขของ slice มี assertion

### T-03 — Implement health และ global routing
- **สิ่งที่ทำ:** ประกอบ Gin handler → service → repo → SQLite สำหรับ health และ routing เท่านั้น โดยคง layer boundaries ตาม PLAN
- **Dependencies:** T-02
- **DoD:** `cd backend; go test ./...` ผ่าน และ `cd backend; go run ./cmd/server` เริ่ม server ได้โดย route ตรง OpenAPI

### T-04 — สร้าง generated typed API client
- **สิ่งที่ทำ:** ตั้งค่า generator และสร้าง typed API client จาก `docs/openapi.yaml`; แยก generated files จาก handwritten wrapper และห้ามแก้ generated files ด้วยมือ
- **Dependencies:** T-03
- **DoD:** `cd frontend; npm run generate:api` และ `cd frontend; npm run build` ผ่าน; generate ซ้ำแล้วไม่มี diff; `src/pages` และ `src/components` ไม่มี direct `fetch`

## Feature: เพิ่มงานวิจัย — AC-1 ถึง AC-4 และ AC-10 ที่เกี่ยวข้อง

### T-05 — เขียน backend tests สำหรับ create research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ success และ atomic persistence, positive unique/non-reused/immutable ID, nullable/existing/non-existing continuation, immutable continuation, initial status/process, root/continuation title rules, concurrent ID/title creation, Unicode trim, boundary/forbidden characters และ request-body rules ทุกกรณีของ POST
- **Dependencies:** T-04
- **DoD:** ทุก test ใช้ SQLite คนละฐานข้อมูลใน `t.TempDir()`; มี assertion ครบ `201/400/404/409/413/415/422/500` และ error codes ตาม SPEC; `cd backend; go test ./...` แดงเพราะ create slice ยังไม่มี implementation เท่านั้น โดยไม่มีเงื่อนไข AC-1–AC-4 หรือ AC-10 ของ create ถูกเลื่อนไป task อื่น

### T-06 — Implement create research
- **สิ่งที่ทำ:** เพิ่ม schema/constraints สำหรับ identity, continuation, validation, defaults และ title rules; เพิ่ม create SQL ใน repo, กฎธุรกิจ/typed errors ใน service และ POST handler เท่านั้น
- **Dependencies:** T-05
- **DoD:** `cd backend; go test ./...` ผ่าน; concurrent writes รักษา ID/title invariants; response/error ตรง OpenAPI และไม่มี SQL หรือ HTTP logic หลุด layer

## Feature: แสดงรายการงานวิจัย — AC-5 และ AC-10 ที่เกี่ยวข้อง

### T-07 — เขียน backend tests สำหรับ list researches
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ `200`, JSON content type/array, fields ทั้งหกและไม่มี field เกิน, เรียง title แล้ว id, empty `[]`, shared list, body/query rejection, database failure และ error format
- **Dependencies:** T-06
- **DoD:** ทุก test ใช้ SQLite คนละฐานข้อมูลใน `t.TempDir()`; มี assertion ครบทุกเงื่อนไข AC-5 และ AC-10 ของ list; `cd backend; go test ./...` แดงเพราะ list slice ยังไม่มี implementation เท่านั้น

### T-08 — Implement list researches
- **สิ่งที่ทำ:** เพิ่ม list SQL ใน repo, orchestration/typed errors ใน service และ GET handler เท่านั้น
- **Dependencies:** T-07
- **DoD:** `cd backend; go test ./...` ผ่าน; response เรียงและมี fields ตรง OpenAPI; layer boundaries ตรง PLAN

### T-09 — สร้าง UI รายการงานวิจัย
- **สิ่งที่ทำ:** อ่านและยึด `docs/DESIGN_BRIEF.md` กับ `.agents/skills/frontend-design/SKILL.md` แล้วสร้าง React page/components สำหรับรายการผ่าน generated client โดยใช้ ID แยกรายการที่ชื่อซ้ำ
- **Dependencies:** T-08
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E checklist ผ่านครบ: loading ระหว่างรอ, error เมื่อ API ล้มเหลว, empty เมื่อได้ `[]`, success แสดงข้อมูลตามลำดับจาก API; layout/responsive/accessibility ตรง Design Brief; ไม่มี direct `fetch`

### T-10 — สร้าง UI เพิ่มงานวิจัย
- **สิ่งที่ทำ:** อ่านและยึด `docs/DESIGN_BRIEF.md` กับ `.agents/skills/frontend-design/SKILL.md` แล้วสร้าง form สำหรับ title, description และ continuationOfId ผ่าน generated client พร้อม refresh list เมื่อสำเร็จ
- **Dependencies:** T-09
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E checklist ผ่านครบ: loading/submitting ป้องกันส่งซ้ำ, error แสดง validation/not-found/conflict/API failure, empty ตรวจฟิลด์จำเป็น, success แสดง record ใหม่และค่าที่ normalize แล้ว; dialog/form/focus/feedback ตรง Design Brief; ไม่มี direct `fetch`

## Feature: แก้ไขงานวิจัย — AC-2 ถึง AC-4, AC-6 และ AC-10 ที่เกี่ยวข้อง

### T-11 — เขียน backend tests สำหรับ update research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ full replacement ของ title/description, ID/continuation/status/process คงเดิม, root/continuation title rules, Unicode validation, positive/not-found ID, body/media/size/JSON ทุกกรณี, rollback, concurrent conflict และ database failure
- **Dependencies:** T-10
- **DoD:** ทุก test ใช้ SQLite คนละฐานข้อมูลใน `t.TempDir()`; มี assertion ครบ `200/400/404/409/413/415/422/500` และทุกเงื่อนไขของ AC ที่ใช้กับ PUT; `cd backend; go test ./...` แดงเพราะ update slice ยังไม่มี implementation เท่านั้น

### T-12 — Implement update research
- **สิ่งที่ทำ:** เพิ่ม update transaction/constraint mapping ใน repo, validation/title rules/typed errors ใน service และ `PUT /api/v1/researches/{id}` handler เท่านั้น
- **Dependencies:** T-11
- **DoD:** `cd backend; go test ./...` ผ่าน; mutation atomic, immutable fields คงเดิม และ response/error ตรง OpenAPI

### T-13 — สร้าง UI แก้ไขงานวิจัย
- **สิ่งที่ทำ:** อ่านและยึด `docs/DESIGN_BRIEF.md` กับ `.agents/skills/frontend-design/SKILL.md` แล้วเพิ่ม edit flow ที่ใช้ ID ใน path และส่งเฉพาะ title/description ผ่าน generated client
- **Dependencies:** T-12
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E checklist ผ่านครบ: loading ระหว่างโหลด/บันทึก, error ครอบ validation/not-found/conflict/API failure, empty ตรวจฟิลด์จำเป็น, success แสดงค่าล่าสุดโดย ID และ field immutable ไม่เปลี่ยน; dialog/form/focus/feedback ตรง Design Brief; ไม่มี direct `fetch`

## Feature: ลบงานวิจัย — AC-7 และ AC-10 ที่เกี่ยวข้อง

### T-14 — เขียน backend tests สำหรับ delete research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ `204` body ว่าง, หายจาก list, not-found/delete ซ้ำ, ID ไม่ถูกใช้ซ้ำ, positive ID, parent ที่มี continuation, database restriction, body/query rejection, concurrent delete และ database failure
- **Dependencies:** T-13
- **DoD:** ทุก test ใช้ SQLite คนละฐานข้อมูลใน `t.TempDir()`; มี assertion ครบ `204/400/404/409/422/500` และทุกเงื่อนไข AC-7/AC-10 ของ delete; `cd backend; go test ./...` แดงเพราะ delete slice ยังไม่มี implementation เท่านั้น

### T-15 — Implement delete research
- **สิ่งที่ทำ:** เพิ่ม delete SQL/constraint mapping ใน repo, not-found/continuation typed errors ใน service และ `DELETE /api/v1/researches/{id}` handler เท่านั้น
- **Dependencies:** T-14
- **DoD:** `cd backend; go test ./...` ผ่าน; parent restriction และ non-reused ID ถูกบังคับที่ SQLite; success ไม่มี body และ response/error ตรง OpenAPI

### T-16 — สร้าง UI ลบงานวิจัย
- **สิ่งที่ทำ:** อ่านและยึด `docs/DESIGN_BRIEF.md` กับ `.agents/skills/frontend-design/SKILL.md` แล้วเพิ่ม delete action ด้วย ID พร้อม confirmation และ refresh list ผ่าน generated client
- **Dependencies:** T-15
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E checklist ผ่านครบ: loading/deleting ป้องกันกดซ้ำ, error ครอบ not-found/continuation/API failure, empty เมื่อรายการสุดท้ายถูกลบ, success นำ record ที่มี ID ตรงกันออก; confirmation/focus/destructive feedback ตรง Design Brief; cancel แล้วข้อมูลไม่เปลี่ยนและไม่มี direct `fetch`

## Feature: ปรับสถานะ — AC-4, AC-8 และ AC-10 ที่เกี่ยวข้อง

### T-17 — เขียน backend tests สำหรับ status transition
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ transition ปกติทีละขั้น, ข้ามไป terminal จากทุกสถานะปกติ, ห้ามข้าม/ย้อน, same-value idempotency, terminal idempotency/lock, process คงเดิม, ID/ข้อมูลอื่นคงเดิม, positive/not-found ID, body/media/size/JSON ทุกกรณี, atomic concurrent transition และ database failure
- **Dependencies:** T-16
- **DoD:** ทุก test ใช้ SQLite คนละฐานข้อมูลใน `t.TempDir()`; มี assertion ครบ `200/400/404/409/413/415/422/500`, `INVALID_STATUS_TRANSITION`, `PROJECT_ALREADY_ENDED` และทุกเงื่อนไข AC-4/AC-8/AC-10 ของ status; `cd backend; go test ./...` แดงเพราะ status slice ยังไม่มี implementation เท่านั้น

### T-18 — Implement status transition
- **สิ่งที่ทำ:** เพิ่ม status constraints/atomic update ใน repo, transition rules/typed errors ใน service และ `PATCH /api/v1/researches/{id}/status` handler เท่านั้น
- **Dependencies:** T-17
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite และ service บังคับ transition/terminal invariants และ response/error ตรง OpenAPI

## Feature: ปรับกระบวนการ — AC-4, AC-9 และ AC-10 ที่เกี่ยวข้อง

### T-19 — เขียน backend tests สำหรับ process transition
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบแปดขั้น, เดินได้ทีละขั้น, ห้ามข้าม/ย้อน/ไปต่อจากขั้นสุดท้าย, same-value idempotency, status ไม่เปลี่ยน, ไม่เปลี่ยน status อัตโนมัติเมื่อครบ, terminal lock, ID/ข้อมูลอื่นคงเดิม, positive/not-found ID, body/media/size/JSON ทุกกรณี, atomic concurrent transition และ database failure
- **Dependencies:** T-18
- **DoD:** ทุก test ใช้ SQLite คนละฐานข้อมูลใน `t.TempDir()`; มี assertion ครบ `200/400/404/409/413/415/422/500`, `INVALID_PROCESS_TRANSITION`, `PROJECT_ALREADY_ENDED` และทุกเงื่อนไข AC-4/AC-9/AC-10 ของ process; `cd backend; go test ./...` แดงเพราะ process slice ยังไม่มี implementation เท่านั้น

### T-20 — Implement process transition
- **สิ่งที่ทำ:** เพิ่ม process constraints/atomic update ใน repo, transition rules/typed errors ใน service และ `PATCH /api/v1/researches/{id}/process` handler เท่านั้น
- **Dependencies:** T-19
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite และ service บังคับลำดับ/terminal invariants และ response/error ตรง OpenAPI

## Hardening

### T-21 — รัน quality gates บนเครื่อง
- **สิ่งที่ทำ:** รัน OpenAPI validation, backend tests/lint/security, generated-client drift, frontend checks, CRUD E2E checklist และตรวจ UI เทียบ `docs/DESIGN_BRIEF.md`; หากพบ defect ให้เพิ่ม failing regression test ก่อนแก้
- **Dependencies:** T-20
- **DoD:** `npx @redocly/cli lint docs/openapi.yaml`; `cd backend; go test ./...`; `cd backend; golangci-lint run`; `cd backend; gosec ./...`; `cd backend; govulncheck ./...`; `cd frontend; npm run generate:api`; `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่านทั้งหมด; generated client ไม่มี drift; E2E loading/error/empty/success ของ list/create/update/delete ผ่าน; visual/responsive/accessibility review ตรง `docs/DESIGN_BRIEF.md`

### T-22 — เพิ่ม CI ด้วย GitHub Actions
- **สิ่งที่ทำ:** สร้าง workflow ที่ pin Go/Node, validate OpenAPI, ตรวจ generated-client drift และรัน quality gates เดียวกับ T-21 บน clean checkout
- **Dependencies:** T-21
- **DoD:** GitHub Actions รัน OpenAPI lint, `go test ./...`, `golangci-lint run`, `gosec ./...`, `govulncheck ./...`, `npm run generate:api`, drift check, `npm run lint` และ `npm run build`; workflow ผ่าน และพิสูจน์ว่า fail เมื่อทำ test ตัวอย่างให้แดง

### T-23 — Cross-agent review ด้วย session ใหม่
- **สิ่งที่ทำ:** เปิด session ใหม่ให้ agent ที่ไม่มีบริบท implementation ตรวจ AGENTS, SPEC, PLAN, TASKS, OpenAPI, `docs/DESIGN_BRIEF.md`, `.agents/skills/frontend-design/SKILL.md`, AC coverage, layer boundaries, database invariants, security และ CRUD UI flows พร้อมจัดระดับ findings
- **Dependencies:** T-22
- **DoD:** reviewer ยืนยันว่า endpoint/error code ตรง SPEC, ทุกเงื่อนไข AC มี test, test database แยกกัน, ไม่มี direct `fetch`, business rules อยู่ถูก layer และ CRUD UI ตรง Design Brief/skill โดยไม่เปลี่ยน scope; findings ระดับสูงถูกแก้โดยมี failing regression test ก่อน fix; rerun T-21 และ CI ผ่านทั้งหมด
