# TASKS: Tracking Research Status MVP

ทำตามลำดับทีละหนึ่ง task เท่านั้น งานทดสอบต้องมาก่อนงาน implementation ของ slice เดียวกันเสมอ

## Contract

### T-01 — ปรับ OpenAPI contract
- **สิ่งที่ทำ:** ปรับ `docs/openapi.yaml` ให้ตรง SPEC ล่าสุดสำหรับ health, list, create, update, delete, status และ process รวม schema, validation และ error code โดยไม่เพิ่ม endpoint
- **Dependencies:** ไม่มี
- **DoD:** `npx @redocly/cli lint docs/openapi.yaml` ผ่าน; ครบ AC-1–AC-10; ยังไม่มีการแก้ backend/frontend

## Walking skeleton

### T-02 — เขียน tests สำหรับ health และ global routing
- **สิ่งที่ทำ:** เขียน handler/integration tests สำหรับ health `200/503/500`, JSON response, unknown route `404` และ wrong method `405` โดยไม่เปิดเผย internal error
- **Dependencies:** T-01
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ error format ของ slice; `cd backend; go test ./...` แดงเพราะ skeleton ยังไม่มี implementation

### T-03 — Implement health และ global routing
- **สิ่งที่ทำ:** ประกอบ Gin handler → service → repo → SQLite สำหรับ health และ global routing เท่านั้น
- **Dependencies:** T-02
- **DoD:** `cd backend; go test ./...` ผ่าน; `cd backend; go run ./cmd/server` เริ่มได้และ route ตรง OpenAPI

### T-04 — สร้าง generated typed API client
- **สิ่งที่ทำ:** ตั้งค่า generator และสร้าง typed API client จาก `docs/openapi.yaml`; แยก generated files จาก wrapper และไม่แก้ generated files ด้วยมือ
- **Dependencies:** T-03
- **DoD:** `cd frontend; npm run generate:api` และ `cd frontend; npm run build` ผ่าน; generate ซ้ำไม่มี diff; pages/components ไม่มี direct `fetch`

## Feature: เพิ่มงานวิจัย — AC-1 ถึง AC-4 และ AC-10 ที่เกี่ยวข้อง

### T-05 — เขียน backend tests สำหรับ create research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ success/atomic persistence, positive unique/non-reused/immutable ID, continuation, initial status/process, title rules, concurrent create, Unicode validation และ request-body rules ทุกกรณีของ POST
- **Dependencies:** T-04
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ `201/400/404/409/413/415/422/500` และทุกเงื่อนไข AC-1–AC-4/AC-10 ของ POST; `cd backend; go test ./...` แดงเพราะ create ยังไม่มี implementation

### T-06 — Implement create research
- **สิ่งที่ทำ:** เพิ่ม schema/constraints, create SQL, service rules/typed errors และ POST handler เท่านั้น
- **Dependencies:** T-05
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite บังคับ identity, foreign key และ title invariants; response/error ตรง OpenAPI

## Feature: แสดงรายการงานวิจัย — AC-5 และ AC-10 ที่เกี่ยวข้อง

### T-07 — เขียน backend tests สำหรับ list researches
- **สิ่งที่ทำ:** เขียน repo/service/handler tests สำหรับ `200`, JSON array/content type, fields ตาม contract, sort title/id, empty, body/query rejection และ database failure
- **Dependencies:** T-06
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ AC-5/AC-10 ของ GET; `cd backend; go test ./...` แดงเพราะ list ยังไม่มี implementation

### T-08 — Implement list researches
- **สิ่งที่ทำ:** เพิ่ม list SQL, service orchestration/typed errors และ GET handler เท่านั้น
- **Dependencies:** T-07
- **DoD:** `cd backend; go test ./...` ผ่าน; response เรียงและมี fields ตรง OpenAPI; layer boundaries ตรง PLAN

### T-09 — สร้าง UI รายการงานวิจัย
- **สิ่งที่ทำ:** อ่าน `.agents/skills/frontend-design/SKILL.md` แล้วสร้าง page/components รายการผ่าน generated client โดยใช้ ID แยกรายการชื่อซ้ำ
- **Dependencies:** T-08
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E: loading, error เมื่อ API ล้มเหลว, empty เมื่อ `[]`, success แสดงตาม API; responsive, keyboard, focus และ contrast ตรง skill; ไม่มี direct `fetch`

### T-10 — สร้าง UI เพิ่มงานวิจัย
- **สิ่งที่ทำ:** อ่าน `.agents/skills/frontend-design/SKILL.md` แล้วสร้าง form title, description, continuationOfId ผ่าน generated client และ refresh รายการหลังบันทึกสำเร็จ
- **Dependencies:** T-09
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E: loading/submitting ป้องกันส่งซ้ำ, error ครบ validation/not-found/conflict/API failure, empty ตรวจฟิลด์จำเป็น, success แสดง record ใหม่และข้อมูลที่ normalize แล้ว; ไม่มี direct `fetch`

## Feature: แก้ไขงานวิจัย — AC-2 ถึง AC-4, AC-6 และ AC-10 ที่เกี่ยวข้อง

### T-11 — เขียน backend tests สำหรับ update research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests สำหรับ full replacement, immutable fields, title rules, Unicode validation, positive/not-found ID, body/media/size/JSON, rollback, concurrent conflict และ database failure
- **Dependencies:** T-10
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ `200/400/404/409/413/415/422/500` และทุกเงื่อนไข AC ของ PUT; `cd backend; go test ./...` แดงเพราะ update ยังไม่มี implementation

### T-12 — Implement update research
- **สิ่งที่ทำ:** เพิ่ม update transaction/constraint mapping, service validation/title rules/typed errors และ PUT handler เท่านั้น
- **Dependencies:** T-11
- **DoD:** `cd backend; go test ./...` ผ่าน; mutation atomic, immutable fields คงเดิม และ response/error ตรง OpenAPI

### T-13 — สร้าง UI แก้ไขงานวิจัย
- **สิ่งที่ทำ:** อ่าน `.agents/skills/frontend-design/SKILL.md` แล้วเพิ่ม edit flow ที่ใช้ ID และส่งเฉพาะ title/description ผ่าน generated client
- **Dependencies:** T-12
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E: loading, error ครบ validation/not-found/conflict/API failure, empty ตรวจฟิลด์, success แสดงค่าล่าสุดโดย ID และ immutable fields ไม่เปลี่ยน; ไม่มี direct `fetch`

## Feature: ลบงานวิจัย — AC-7 และ AC-10 ที่เกี่ยวข้อง

### T-14 — เขียน backend tests สำหรับ delete research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests สำหรับ `204` body ว่าง, หายจาก list, not-found/delete ซ้ำ, ID ไม่ใช้ซ้ำ, positive ID, parent continuation restriction, body/query rejection, concurrent delete และ database failure
- **Dependencies:** T-13
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ `204/400/404/409/422/500` และทุกเงื่อนไข AC-7/AC-10 ของ DELETE; `cd backend; go test ./...` แดงเพราะ delete ยังไม่มี implementation

### T-15 — Implement delete research
- **สิ่งที่ทำ:** เพิ่ม delete SQL/constraint mapping, service errors และ DELETE handler เท่านั้น
- **Dependencies:** T-14
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite บังคับ parent restriction/non-reused ID; success ไม่มี body และ response/error ตรง OpenAPI

### T-16 — สร้าง UI ลบงานวิจัย
- **สิ่งที่ทำ:** อ่าน `.agents/skills/frontend-design/SKILL.md` แล้วเพิ่ม delete action ด้วย ID พร้อม confirmation และ refresh list ผ่าน generated client
- **Dependencies:** T-15
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E: loading/deleting ป้องกันกดซ้ำ, error ครบ not-found/continuation/API failure, empty เมื่อลบรายการสุดท้าย, success ลบ record ตรง ID, cancel ไม่เปลี่ยนข้อมูล; ไม่มี direct `fetch`

## Feature: ปรับสถานะ — AC-4, AC-8 และ AC-10 ที่เกี่ยวข้อง

### T-17 — เขียน tests สำหรับ status skeleton
- **สิ่งที่ทำ:** เขียน tests ของ route, handler dependency, service input/interface และ repo contract สำหรับ `PATCH /api/v1/researches/{id}/status` โดยใช้ test double ตรวจว่า HTTP request ถูกส่งต่อเป็น typed input และผลลัพธ์ถูกส่งกลับตาม contract; ยังไม่ทดสอบ transition rules
- **Dependencies:** T-16
- **DoD:** `cd backend; go test ./...` แดงเพราะ status contracts/route ยังไม่มี; ไม่ใส่ business rule, SQL update หรือ status transition tests; ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()` เมื่อแตะ persistence

### T-18 — Implement status skeleton
- **สิ่งที่ทำ:** เพิ่มเฉพาะ route, handler dependency, service input/interface, repo contract และ generated-client exposure ของ status โดยไม่เพิ่ม transition rule, SQL mutation หรือ endpoint นอก OpenAPI
- **Dependencies:** T-17
- **DoD:** `cd backend; go test ./...` และ `cd frontend; npm run generate:api` ผ่าน; T-17 ผ่านและ status contract เชื่อม HTTP → service → repo ได้โดยไม่ทำ mutation

### T-19 — เขียน backend tests สำหรับ status transition
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ transition ปกติทีละขั้น, terminal jump, ห้ามข้าม/ย้อน, idempotency, terminal lock, process/ID/ข้อมูลอื่นคงเดิม, positive/not-found ID, body/media/size/JSON, atomic concurrency และ database failure
- **Dependencies:** T-18
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ `200/400/404/409/413/415/422/500`, `INVALID_STATUS_TRANSITION`, `PROJECT_ALREADY_ENDED` และทุกเงื่อนไข AC-4/AC-8/AC-10 ของ status; `cd backend; go test ./...` แดงเพราะ transition implementation ยังไม่มี

### T-20 — Implement status transition
- **สิ่งที่ทำ:** เพิ่ม status constraint/trigger และ atomic SQL update, transition rules/typed errors ใน service และ PATCH handler behavior เท่านั้น
- **Dependencies:** T-19
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite และ service บังคับ transition/terminal invariants แบบ atomic; response/error ตรง OpenAPI

## Feature: ปรับกระบวนการ — AC-4, AC-9 และ AC-10 ที่เกี่ยวข้อง

### T-21 — เขียน tests สำหรับ process skeleton
- **สิ่งที่ทำ:** เขียน tests ของ route, handler dependency, service input/interface และ repo contract สำหรับ `PATCH /api/v1/researches/{id}/process` โดยใช้ test double; ยังไม่ทดสอบ process rules
- **Dependencies:** T-20
- **DoD:** `cd backend; go test ./...` แดงเพราะ process contracts/route ยังไม่มี; ไม่ใส่ business rule, SQL update หรือ process transition tests; ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()` เมื่อแตะ persistence

### T-22 — Implement process skeleton
- **สิ่งที่ทำ:** เพิ่มเฉพาะ route, handler dependency, service input/interface, repo contract และ generated-client exposure ของ process โดยไม่เพิ่ม transition rule, SQL mutation หรือ endpoint นอก OpenAPI
- **Dependencies:** T-21
- **DoD:** `cd backend; go test ./...` และ `cd frontend; npm run generate:api` ผ่าน; T-21 ผ่านและ process contract เชื่อม HTTP → service → repo ได้โดยไม่ทำ mutation

### T-23 — เขียน backend tests สำหรับ process transition
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบแปดขั้น, เดินทีละขั้น, ห้ามข้าม/ย้อน/ไปต่อจากขั้นสุดท้าย, idempotency, status ไม่เปลี่ยน/ไม่ auto-complete, terminal lock, ID/ข้อมูลอื่นคงเดิม, positive/not-found ID, body/media/size/JSON, atomic concurrency และ database failure
- **Dependencies:** T-22
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ `200/400/404/409/413/415/422/500`, `INVALID_PROCESS_TRANSITION`, `PROJECT_ALREADY_ENDED` และทุกเงื่อนไข AC-4/AC-9/AC-10 ของ process; `cd backend; go test ./...` แดงเพราะ process transition implementation ยังไม่มี

### T-24 — Implement process transition
- **สิ่งที่ทำ:** เพิ่ม process constraint/trigger และ atomic SQL update, transition rules/typed errors ใน service และ PATCH handler behavior เท่านั้น
- **Dependencies:** T-23
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite และ service บังคับลำดับ/final/terminal invariants แบบ atomic; response/error ตรง OpenAPI

## Hardening

### T-25 — รัน quality gates บนเครื่อง
- **สิ่งที่ทำ:** รัน OpenAPI validation, backend tests/lint/security, generated-client drift, frontend checks และ CRUD E2E ตาม skill; defect ทุกข้อเริ่มด้วย failing regression test
- **Dependencies:** T-24
- **DoD:** `npx @redocly/cli lint docs/openapi.yaml`; `cd backend; go test ./...`; `cd backend; golangci-lint run`; `cd backend; gosec ./...`; `cd backend; govulncheck ./...`; `cd frontend; npm run generate:api`; `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่านทั้งหมด; CRUD UI ครบ loading/error/empty/success

### T-26 — เพิ่ม CI ด้วย GitHub Actions
- **สิ่งที่ทำ:** สร้าง workflow clean checkout ที่ pin Go/Node และรัน quality gates เดียวกับ T-25 รวม generated-client drift
- **Dependencies:** T-25
- **DoD:** GitHub Actions รัน OpenAPI lint, backend tests/lint/security, client generation/drift, frontend lint/build ผ่าน และยืนยันว่า CI ล้มเมื่อมี test แดง

### T-27 — Cross-agent review ด้วย session ใหม่
- **สิ่งที่ทำ:** ให้ agent ใน session ใหม่ตรวจ AGENTS, SPEC, PLAN, TASKS, OpenAPI, AC coverage, layer boundaries, database invariants, security และ CRUD UI พร้อมระดับ findings
- **Dependencies:** T-26
- **DoD:** findings ระดับสูงถูกแก้โดยมี failing regression test ก่อน fix; reviewer ยืนยัน endpoint/error code, AC tests, isolated SQLite, no direct fetch และ CRUD UI states; rerun T-25 และ CI ผ่าน
