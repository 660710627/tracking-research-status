# TASKS: Tracking Research Status MVP

ทำตามลำดับทีละหนึ่ง task เท่านั้น งาน test ต้องมาก่อน implementation ของ slice เดียวกัน

## Contract และ walking skeleton

### T-01 — ปรับ OpenAPI contract
- **สิ่งที่ทำ:** ทำ `docs/openapi.yaml` ให้ตรง AC-1–AC-10: research data, members, PDF contract, multipart create/update, health, list, delete, status และ process โดยไม่เพิ่ม endpoint
- **Dependencies:** ไม่มี
- **DoD:** `npx @redocly/cli lint docs/openapi.yaml` ผ่าน; ยังไม่มีโค้ด backend/frontend

### T-02 — เขียน tests สำหรับ health และ global routing
- **สิ่งที่ทำ:** เขียน handler/integration tests สำหรับ health `200/503/500`, JSON error envelope, unknown route `404` และ wrong method `405`
- **Dependencies:** T-01
- **DoD:** ทุก persistence test ใช้ SQLite ใหม่ใน `t.TempDir()`; `cd backend; go test ./...` แดงเพราะยังไม่มี implementation

### T-03 — Implement health และ global routing
- **สิ่งที่ทำ:** ประกอบ Gin handler → service → repo → SQLite สำหรับ health และ global routing เท่านั้น
- **Dependencies:** T-02
- **DoD:** `cd backend; go test ./...` ผ่าน; `cd backend; go run ./cmd/server` เริ่มได้

### T-04 — สร้าง generated typed API client
- **สิ่งที่ทำ:** ตั้งค่าและ generate client จาก OpenAPI; แยก generated files ออกจาก wrapper และห้ามแก้ generated files ด้วยมือ
- **Dependencies:** T-03
- **DoD:** `cd frontend; npm run generate:api`; `cd frontend; npm run build` ผ่าน; generate ซ้ำไม่มี diff; components ไม่มี direct `fetch`

## Feature: เพิ่มงานวิจัย — AC-1 ถึง AC-4 และ AC-10 ที่เกี่ยวข้อง

### T-04A — เตรียม domain model และ test seams สำหรับ create research
- **สิ่งที่ทำ:** เพิ่มเฉพาะ Go domain types และ handler/service/repo interfaces สำหรับข้อมูลโครงการ, members, funding contract metadata, PDF staging และ typed errors ที่ T-05 ต้องใช้; ห้ามเพิ่ม schema, validation rule, file write หรือพฤติกรรม endpoint
- **Dependencies:** T-04
- **DoD:** `cd backend; go test ./...` ผ่าน; ไม่มีการแก้ `*_test.go`; tests ของ T-05 สามารถ compile โดยใช้ domain model/test seams เหล่านี้

### T-04B — เตรียม concrete adapter skeleton สำหรับ create research
- **สิ่งที่ทำ:** เพิ่มเฉพาะ method/constructor และ dependency wiring ขั้นต่ำใน handler → service → repo รวมทั้ง inject `ContractStager` เข้า service เพื่อให้ test ของ T-05 เรียก concrete seams และตรวจ file lifecycle ได้; method ต้องคืน typed `not implemented` error โดยห้าม parse multipart, เพิ่ม route, เพิ่ม schema/SQL, validation หรือ file I/O
- **Dependencies:** T-04A
- **DoD:** `cd backend; go test ./...` ผ่าน; ไม่มีการแก้ `*_test.go`; handler/service/repo concrete seams สำหรับ create ใหม่ compile ได้ และยังไม่มีพฤติกรรม create ใหม่

### T-05 — เขียน backend tests สำหรับ create research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ multipart required parts/unknown-or-duplicate parts, JSON members, member roles/percent/email, field/date/budget/enum validation, PDF type/20 MiB, staged-file cleanup, atomic persistence, continuation/title/identity rules, initial status/process, concurrent create และ error mapping
- **Dependencies:** T-04B
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()` และ isolated file storage; ครบ AC-1–AC-4/AC-10 ของ POST; `cd backend; go test ./...` แดงเพราะ create ยังไม่มี implementation

### T-06 — Implement create research
- **สิ่งที่ทำ:** ล้างข้อมูลเก่าตาม SPEC และเพิ่ม schema/constraints/triggers, staged PDF storage, create SQL, service typed errors และ POST handler เท่านั้น
- **Dependencies:** T-05
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite บังคับ identity, FK, member/contract/title invariants และสร้างข้อมูล/PDF แบบ atomic; response/error ตรง OpenAPI

### T-07 — สร้าง UI เพิ่มงานวิจัย
- **สิ่งที่ทำ:** อ่าน `.agents/skills/frontend-design/SKILL.md`; สร้าง form multipart สำหรับข้อมูลโครงการ, บุคลากร, แหล่งทุน, วันที่, งบประมาณ, เนื้อหา และ PDF โดยเก็บไฟล์ใน browser จน submit
- **Dependencies:** T-06
- **DoD:** `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่าน; E2E: loading/submitting, validation/API error, cancel ไม่อัปโหลดไฟล์, success แสดงข้อมูลที่ normalize, keyboard/focus/contrast/responsive; ไม่มี direct `fetch`

## Feature: แสดงรายการและค้นหา — AC-5 และ AC-10 ที่เกี่ยวข้อง

### T-08 — เขียน backend tests สำหรับ list researches
- **สิ่งที่ทำ:** เขียน repo/service/handler tests สำหรับ `200`, JSON array/content type, fields/contract metadata, title/id sort, empty, body/query rejection และ database failure
- **Dependencies:** T-07
- **DoD:** ทุก persistence test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ AC-5/AC-10 ของ GET; `cd backend; go test ./...` แดงเพราะ list ยังไม่มี implementation

### T-09 — Implement list researches
- **สิ่งที่ทำ:** เพิ่ม list SQL สำหรับ research/member/contract, service orchestration/typed errors และ GET handler เท่านั้น
- **Dependencies:** T-08
- **DoD:** `cd backend; go test ./...` ผ่าน; response เรียงและตรง OpenAPI; layer boundaries ตรง PLAN

### T-10 — สร้าง UI รายการงานวิจัย
- **สิ่งที่ทำ:** อ่าน skill; สร้างหน้า list ผ่าน generated client แสดงข้อมูลตาม contract และ ID แยกรายการชื่อซ้ำ
- **Dependencies:** T-09
- **DoD:** `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่าน; E2E: loading, API error/retry, empty, success, responsive/keyboard/focus/contrast; ไม่มี direct `fetch`

### T-11 — สร้าง UI ค้นหาและกรองงานวิจัย
- **สิ่งที่ทำ:** อ่าน skill; เพิ่ม search/filter ฝั่ง client จากรายการที่ API ส่ง สำหรับรหัส ชื่อ สถานะ และกระบวนการ พร้อมเปิด/ยุบและล้างตัวกรอง
- **Dependencies:** T-10
- **DoD:** `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่าน; E2E: loading/error/empty/success, ผลค้นหาเป็นศูนย์, clear filter, keyboard/focus/contrast/responsive

## Feature: แก้ไขงานวิจัย — AC-2 ถึง AC-4, AC-6 และ AC-10 ที่เกี่ยวข้อง

### T-12 — เขียน backend tests สำหรับ update research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ full multipart replacement, immutable ID/continuation/researchKind/status/process, member/contract replacement, PDF replacement rollback, field validation, title rules, positive/not-found ID, media/size/multipart errors, concurrency และ database failure
- **Dependencies:** T-11
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()` และ isolated file storage; ครบทุก AC ของ PUT; `cd backend; go test ./...` แดงเพราะ update ยังไม่มี implementation

### T-13 — Implement update research
- **สิ่งที่ทำ:** เพิ่ม update transaction/constraint mapping, replacement file lifecycle, service validation/typed errors และ PUT handler เท่านั้น
- **Dependencies:** T-12
- **DoD:** `cd backend; go test ./...` ผ่าน; mutation และ PDF replacement atomic, immutable fields คงเดิม, response/error ตรง OpenAPI

### T-14 — สร้าง UI แก้ไขงานวิจัย
- **สิ่งที่ทำ:** อ่าน skill; สร้าง edit flow ที่เติมข้อมูลเดิม, แก้ field ที่อนุญาต, เลือก PDF ใหม่ และส่ง multipart ผ่าน generated client
- **Dependencies:** T-13
- **DoD:** `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่าน; E2E: loading, validation/not-found/conflict/API error, cancel ไม่บันทึกข้อมูล/ไฟล์, success แสดงค่าล่าสุด, immutable fields เป็น read-only, keyboard/focus/contrast/responsive

## Feature: ลบงานวิจัย — AC-7 และ AC-10 ที่เกี่ยวข้อง

### T-15 — เขียน backend tests สำหรับ delete research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests สำหรับ `204` body ว่าง, หายจาก list, file cleanup, not-found/delete ซ้ำ, ID ไม่ใช้ซ้ำ, continuation restriction, body/query rejection, concurrent delete และ database failure
- **Dependencies:** T-14
- **DoD:** ทุก test ใช้ SQLite ใหม่ใน `t.TempDir()` และ isolated file storage; ครบ AC-7/AC-10; `cd backend; go test ./...` แดงเพราะ delete ยังไม่มี implementation

### T-16 — Implement delete research
- **สิ่งที่ทำ:** เพิ่ม delete SQL/constraint mapping, transaction file cleanup, service errors และ DELETE handler เท่านั้น
- **Dependencies:** T-15
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite บังคับ parent restriction/non-reused ID; success ไม่มี body และไฟล์ถูกลบหลัง commit

### T-17 — สร้าง UI ลบงานวิจัย
- **สิ่งที่ทำ:** อ่าน skill; เพิ่ม delete action ด้วย ID พร้อม confirmation และ refresh list ผ่าน generated client
- **Dependencies:** T-16
- **DoD:** `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่าน; E2E: loading/deleting, not-found/continuation/API error, empty เมื่อลบรายการสุดท้าย, success/cancel, keyboard/focus/contrast/responsive

## Feature: ปรับสถานะ — AC-4, AC-8 และ AC-10 ที่เกี่ยวข้อง

### T-18 — เขียน backend tests สำหรับ status transition
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ normal transition ทีละขั้น, terminal jump, ห้ามข้าม/ย้อน, idempotency, terminal lock, fields อื่นคงเดิม, positive/not-found ID, JSON/media/size/body errors, atomic concurrency และ database failure
- **Dependencies:** T-17
- **DoD:** ทุก persistence test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ AC-4/AC-8/AC-10, `200/400/404/409/413/415/422/500`; `cd backend; go test ./...` แดงเพราะ transition ยังไม่มี implementation

### T-19 — Implement status transition
- **สิ่งที่ทำ:** เพิ่ม status constraint/trigger, atomic SQL update, service typed errors และ PATCH handler เท่านั้น
- **Dependencies:** T-18
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite และ service บังคับ transition/terminal invariants แบบ atomic; response/error ตรง OpenAPI

### T-20 — สร้าง UI ปรับสถานะงานวิจัย
- **สิ่งที่ทำ:** อ่าน skill; สร้างหน้าค้นหา/เลือกงาน แสดงสถานะปัจจุบันและเฉพาะสถานะที่เปลี่ยนได้ พร้อม confirmation และเรียก generated client
- **Dependencies:** T-19
- **DoD:** `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่าน; E2E: loading, empty search, not-found/transition/terminal/API error, confirm/cancel, success refresh, keyboard/focus/contrast/responsive

## Feature: ปรับกระบวนการ — AC-4, AC-9 และ AC-10 ที่เกี่ยวข้อง

### T-21 — เขียน backend tests สำหรับ process transition
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบแปดขั้น, เดินทีละขั้น, ห้ามข้าม/ย้อน/ต่อจากขั้นสุดท้าย, idempotency, status ไม่เปลี่ยน, terminal lock, fields อื่นคงเดิม, positive/not-found ID, JSON/media/size/body errors, atomic concurrency และ database failure
- **Dependencies:** T-20
- **DoD:** ทุก persistence test ใช้ SQLite ใหม่ใน `t.TempDir()`; ครบ AC-4/AC-9/AC-10, `200/400/404/409/413/415/422/500`; `cd backend; go test ./...` แดงเพราะ transition ยังไม่มี implementation

### T-22 — Implement process transition
- **สิ่งที่ทำ:** เพิ่ม process constraint/trigger, atomic SQL update, service typed errors และ PATCH handler เท่านั้น
- **Dependencies:** T-21
- **DoD:** `cd backend; go test ./...` ผ่าน; SQLite และ service บังคับลำดับ/final/terminal invariants แบบ atomic; response/error ตรง OpenAPI

### T-23 — สร้าง UI ปรับกระบวนการงานวิจัย
- **สิ่งที่ทำ:** อ่าน skill; สร้างหน้าค้นหา/เลือกงาน แสดงลำดับแปดขั้น, ขั้นปัจจุบัน/ถัดไป, confirmation และเรียก generated client
- **Dependencies:** T-22
- **DoD:** `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่าน; E2E: loading, empty search, not-found/invalid/final/terminal/API error, confirm/cancel, success refresh, keyboard/focus/contrast/responsive

## Hardening

### T-24 — รัน quality gates บนเครื่อง
- **สิ่งที่ทำ:** รัน OpenAPI validation, backend tests/lint/security, generated-client drift, frontend checks และ CRUD/status/process UI E2E
- **Dependencies:** T-23
- **DoD:** `npx @redocly/cli lint docs/openapi.yaml`; `cd backend; go test ./...`; `cd backend; golangci-lint run`; `cd backend; gosec ./...`; `cd backend; govulncheck ./...`; `cd frontend; npm run generate:api`; `cd frontend; npm run lint`; `cd frontend; npm run build` ผ่านทั้งหมด; UI ครบ loading/error/empty/success

### T-25 — เพิ่ม CI ด้วย GitHub Actions
- **สิ่งที่ทำ:** สร้าง workflow clean checkout ที่ pin Go/Node และรัน gates เดียวกับ T-24 รวม generated-client drift
- **Dependencies:** T-24
- **DoD:** GitHub Actions รัน OpenAPI lint, backend tests/lint/security, client generation/drift และ frontend lint/build ผ่าน; ยืนยันว่า CI ล้มเมื่อมี test แดง

### T-26 — Cross-agent review ด้วย session ใหม่
- **สิ่งที่ทำ:** ให้ agent ใน session ใหม่ตรวจ AGENTS, SPEC, PLAN, TASKS, OpenAPI, AC coverage, layer boundaries, database/file invariants, security และ CRUD/status/process UI พร้อมระดับ findings
- **Dependencies:** T-25
- **DoD:** findings ระดับสูงถูกแก้โดยมี failing regression test ก่อน fix; reviewer ยืนยัน endpoint/error code, AC tests, isolated SQLite/file storage, no direct fetch และ UI states; rerun T-24 และ CI ผ่าน
