# TASKS Tracking-Research-Status MVP

ทำตามลำดับทีละหนึ่ง task เท่านั้น ห้ามเริ่ม implementation ของ slice ก่อน test task ของ slice นั้นแดงตามที่กำหนด

## Contract

### T-01 — จัดทำ OpenAPI contract
- **สิ่งที่ทำ:** สร้าง `docs/openapi.yaml` ให้ตรงกับ API Contract และ error format ใน `SPEC.md` เท่านั้น ครบ `GET /health`, `GET/POST /api/v1/researches`, `PUT/DELETE /api/v1/researches/{title}`; ไม่มี ID, auth, status หรือ endpoint เพิ่มเติม
- **Dependencies:** ไม่มี
- **DoD:** OpenAPI ระบุ request/response schema, required fields, limits, status และ error codes ครบทุก AC; `npx @redocly/cli lint docs/openapi.yaml` ผ่าน; human ยืนยัน contract ก่อนเริ่ม T-02

## Walking skeleton

### T-02 — เขียน backend tests สำหรับ health และ routing
- **สิ่งที่ทำ:** เขียน tests ครบกรณี `200`, `503 SERVICE_UNAVAILABLE`, `500 INTERNAL_ERROR`, JSON content type/error format รวมถึง global `404 ROUTE_NOT_FOUND` และ `405 METHOD_NOT_ALLOWED`
- **Dependencies:** T-01
- **DoD:** ทุก test แยก SQLite ด้วย `t.TempDir()` และหนึ่งฐานข้อมูลต่อหนึ่ง test; `cd backend; go test ./...` แดงเพราะ walking skeleton ยังไม่ถูก implement เท่านั้น และ test ครบ contract ของ slice

### T-03 — Implement walking skeleton
- **สิ่งที่ทำ:** สร้างโครง `cmd/server`, `handler`, `service`, `repo`, `db`; ต่อ Gin ถึง SQLite และ implement เฉพาะ health/routing โดย handler รู้ HTTP, service รู้กฎการพร้อมใช้งาน และ repo รู้การตรวจฐานข้อมูล
- **Dependencies:** T-02
- **DoD:** `cd backend; go test ./...` ผ่าน และ server เริ่มได้ด้วย `go run ./cmd/server` โดย routes ตรงกับ OpenAPI

### T-04 — สร้าง generated typed API client
- **สิ่งที่ทำ:** ตั้งค่า generator และสร้าง frontend typed API client จาก `docs/openapi.yaml`; generated files แยกจาก handwritten wrapper และห้ามแก้ด้วยมือ
- **Dependencies:** T-03
- **DoD:** `cd frontend; npm run generate:api` และ `npm run build` ผ่าน; generate ซ้ำให้ผลเดิม; ไม่มี `fetch` โดยตรงใน `src/pages` หรือ `src/components`

## Feature: แสดงรายการงานวิจัย — AC-5

### T-05 — เขียน backend tests สำหรับ list researches
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ AC-5: `200`, JSON array, fields เฉพาะ title/description, ไม่มี ID, เรียง title, empty `[]`, shared list, body → `400`, query → `422`, database failure → `500`
- **Dependencies:** T-04
- **DoD:** SQLite ทุก test อยู่ใน `t.TempDir()` และไม่ใช้ database ร่วมกัน; `cd backend; go test ./...` แดงเพราะ list slice ยังไม่ถูก implement เท่านั้น และทุกเงื่อนไข AC-5 มี assertion

### T-06 — Implement list researches
- **สิ่งที่ทำ:** เพิ่ม SQL query ใน repo, orchestration ใน service และ HTTP handler สำหรับ `GET /api/v1/researches` เท่านั้น
- **Dependencies:** T-05
- **DoD:** `cd backend; go test ./...` ผ่าน; response และ errors ตรง OpenAPI; ไม่มี SQL ใน service/handler และไม่มี HTTP logic ใน service/repo

### T-07 — สร้าง UI รายการงานวิจัย
- **สิ่งที่ทำ:** สร้าง page และ components สำหรับแสดงรายการผ่าน generated client เท่านั้น
- **Dependencies:** T-06
- **DoD:** `cd frontend; npm run lint` และ `npm run build` ผ่าน; E2E checklist ผ่านครบ: loading แสดงระหว่างรอ, error แสดงเมื่อ API ล้มเหลว, empty แสดงเมื่อได้ `[]`, success แสดง title/description ตามลำดับจาก API; ไม่มี component/page เรียก `fetch` ตรง

## Feature: เพิ่มงานวิจัย — AC-1 ถึง AC-4

### T-08 — เขียน backend tests สำหรับ create research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบทุกเงื่อนไข AC-1–AC-4: success/trim/persistence/response fields, title uniqueness และ concurrency, title/description ทุก boundary และ forbidden character, Content-Type, empty/malformed/multiple/top-level JSON, duplicate/unknown keys และ body limit
- **Dependencies:** T-07
- **DoD:** SQLite ทุก test อยู่ใน `t.TempDir()` และหนึ่งฐานข้อมูลต่อหนึ่ง test; `cd backend; go test ./...` แดงเพราะ create slice ยังไม่ถูก implement เท่านั้น; ทุก status/error code และทุกเงื่อนไข AC-1–AC-4 มี assertion โดยไม่เลื่อนไป task อื่น

### T-09 — Implement create research
- **สิ่งที่ทำ:** เพิ่ม insert SQL/constraint mapping ใน repo, trim/validation/typed errors ใน service และ HTTP parsing/error mapping สำหรับ `POST /api/v1/researches` ใน handler
- **Dependencies:** T-08
- **DoD:** `cd backend; go test ./...` ผ่าน; `201/400/409/413/415/422/500` และ error codes ตรง OpenAPI; `UNIQUE(title)` บังคับที่ SQLite; layer boundaries ตรง PLAN

### T-10 — สร้าง UI เพิ่มงานวิจัย
- **สิ่งที่ทำ:** สร้าง form/component สำหรับ title และ description; submit ผ่าน generated client; refresh list เมื่อสำเร็จ; map errors ด้วย error code
- **Dependencies:** T-09
- **DoD:** `cd frontend; npm run lint` และ `npm run build` ผ่าน; E2E checklist ผ่านครบ: loading/submitting ป้องกัน submit ซ้ำ, error แสดง validation/conflict/API failure, empty ปฏิเสธฟิลด์ว่าง, success แสดงรายการที่สร้างและค่าที่ trim แล้ว; ไม่มี direct `fetch`

## Feature: แก้ไขงานวิจัย — AC-2 ถึง AC-4 และ AC-6

### T-11 — เขียน backend tests สำหรับ update research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ AC-6 และทวนทุกเงื่อนไข AC-2–AC-4 สำหรับ PUT: full replacement, เปลี่ยน/คง title, old title ใช้ไม่ได้, not found, conflict, invalid path/body/media/size, transaction rollback และ database failure
- **Dependencies:** T-10
- **DoD:** SQLite ทุก test อยู่ใน `t.TempDir()` และหนึ่งฐานข้อมูลต่อหนึ่ง test; `cd backend; go test ./...` แดงเพราะ update slice ยังไม่ถูก implement เท่านั้น; ทุกเงื่อนไข AC ที่ใช้กับ PUT มี assertion ครบ

### T-12 — Implement update research
- **สิ่งที่ทำ:** เพิ่ม update transaction ใน repo, validation/conflict/not-found typed errors ใน service และ handler สำหรับ `PUT /api/v1/researches/{title}` เท่านั้น
- **Dependencies:** T-11
- **DoD:** `cd backend; go test ./...` ผ่าน; `200/400/404/409/413/415/422/500` และ error codes ตรง OpenAPI; rollback ไม่ทิ้งข้อมูลแก้ไขบางส่วน

### T-13 — สร้าง UI แก้ไขงานวิจัย
- **สิ่งที่ทำ:** เพิ่ม edit flow ที่ใช้ title ปัจจุบันใน path และส่ง title/description ใหม่ผ่าน generated client
- **Dependencies:** T-12
- **DoD:** `cd frontend; npm run lint` และ `npm run build` ผ่าน; E2E checklist ผ่านครบ: loading ระหว่างบันทึก, error ครอบ validation/not-found/conflict/API failure, empty ปฏิเสธฟิลด์ว่าง, success แสดงค่าที่แก้และไม่แสดง title เดิม; ไม่มี direct `fetch`

## Feature: ลบงานวิจัย — AC-7

### T-14 — เขียน backend tests สำหรับ delete research
- **สิ่งที่ทำ:** เขียน repo/service/handler tests ครบ AC-7: `204` body ว่าง, หายจาก list, not found, delete ซ้ำ, request body, query parameters และ database failure
- **Dependencies:** T-13
- **DoD:** SQLite ทุก test อยู่ใน `t.TempDir()` และหนึ่งฐานข้อมูลต่อหนึ่ง test; `cd backend; go test ./...` แดงเพราะ delete slice ยังไม่ถูก implement เท่านั้น; ทุกเงื่อนไข AC-7 มี assertion

### T-15 — Implement delete research
- **สิ่งที่ทำ:** เพิ่ม delete SQL ใน repo, not-found typed error ใน service และ handler สำหรับ `DELETE /api/v1/researches/{title}` เท่านั้น
- **Dependencies:** T-14
- **DoD:** `cd backend; go test ./...` ผ่าน; `204/400/404/422/500` และ error codes ตรง OpenAPI; success response ไม่มี body

### T-16 — สร้าง UI ลบงานวิจัย
- **สิ่งที่ทำ:** เพิ่ม delete action พร้อม confirmation และ refresh list ผ่าน generated client
- **Dependencies:** T-15
- **DoD:** `cd frontend; npm run lint` และ `npm run build` ผ่าน; E2E checklist ผ่านครบ: loading/deleting ป้องกันกดซ้ำ, error ครอบ not-found/API failure, empty แสดงเมื่อรายการสุดท้ายถูกลบ, success นำรายการที่ลบออก; ยกเลิก confirmation แล้วข้อมูลไม่เปลี่ยน; ไม่มี direct `fetch`

## Hardening

### T-17 — รัน quality gates บนเครื่อง
- **สิ่งที่ทำ:** รัน test, lint, security, generated-client drift check และ E2E checklist เต็มระบบ; แก้เฉพาะข้อบกพร่องโดยเพิ่ม failing regression test ก่อนแก้
- **Dependencies:** T-16
- **DoD:** `cd backend; go test ./...`, `golangci-lint run`, `gosec ./...`, `govulncheck ./...`, `cd ../frontend; npm run generate:api`, `npm run lint`, `npm run build` ผ่านทั้งหมด; generated client ไม่มี drift; E2E loading/error/empty/success ของ list/create/update/delete ผ่าน

### T-18 — เพิ่ม CI ด้วย GitHub Actions
- **สิ่งที่ทำ:** สร้าง workflow ที่ติดตั้ง Go/Node แบบ pin version, validate OpenAPI, ตรวจ generated-client drift และรัน quality gates เดียวกับ T-17
- **Dependencies:** T-17
- **DoD:** GitHub Actions รัน `go test ./...`, `golangci-lint run`, `gosec ./...`, `govulncheck ./...`, `npm run generate:api`, `npm run lint`, `npm run build` และ OpenAPI lint; workflow ผ่านบน clean checkout และ fail เมื่อ test ตัวอย่างถูกทำให้แดง

### T-19 — Cross-agent review ด้วย session ใหม่
- **สิ่งที่ทำ:** เปิด session ใหม่ให้ agent ที่ไม่มีบริบทการ implement ตรวจ `AGENTS.md`, SPEC, PLAN, TASKS, OpenAPI, layer boundaries, tests, security และ UI flows; บันทึก findings ตามระดับความรุนแรง
- **Dependencies:** T-18
- **DoD:** reviewer ยืนยันว่า endpoints/error codes ตรง SPEC, ทุก AC มี test, ไม่มี direct `fetch`, test DB แยกต่อ test และไม่มี business rule หลุด layer; findings ระดับสูงถูกแก้ด้วย failing regression test ก่อน fix; rerun คำสั่ง T-17 และ CI ผ่านทั้งหมด
