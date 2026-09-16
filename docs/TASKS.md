# TASKS — Tracking Research Status MVP

ทำตามลำดับเท่านั้น งาน test ต้องเสร็จก่อนงาน implement ของ slice เดียวกัน และห้ามแก้ test เพื่อให้ผ่าน

## Phase 0 — Contract และ walking skeleton

### T-01 — API contract
- **สิ่งที่ทำ:** ปรับ `docs/openapi.yaml` ให้ตรง API Contract ใน SPEC ทุก endpoint, schema, permission, status และ error code รวมทั้งตัดสินวิธีออก Access Token กับ human; ไม่เขียนโค้ด
- **Dependencies:** ไม่มี
- **DoD:** OpenAPI validate ผ่าน; endpoint ตรง SPEC โดยไม่เพิ่ม/เปลี่ยนชื่อ; human อนุมัติ auth contract และความต่างระหว่าง SPEC/OpenAPI

### T-02 — Tests: backend walking skeleton
- **สิ่งที่ทำ:** เขียน test สำหรับ SQLite connection/migration, `GET /health`, route not found, method not allowed และ error envelope ตาม AC-24
- **Dependencies:** T-01
- **DoD:** test ใช้ `httptest` และ SQLite จาก `t.TempDir()` หนึ่งฐานต่อ test; ครบ success/database unavailable/internal/404/405; `cd backend; go test ./...` แดงเพราะ implementation ยังไม่มี

### T-03 — Implement: backend walking skeleton
- **สิ่งที่ทำ:** สร้าง Gin router, service/repo boundary, SQLite migration และ typed-error mapper ขั้นต่ำให้ T-02 ผ่าน
- **Dependencies:** T-02
- **DoD:** `cd backend; go test ./...` ผ่าน; server เปิด port 8080 และ health ใช้ฐานข้อมูลจริง

### T-04 — Tests: frontend walking skeleton
- **สิ่งที่ทำ:** เขียน E2E checklist สำหรับ app shell และ typed health client ครบ loading/error/success, keyboard/focus/contrast/responsive/reduced motion ตาม AC-25
- **Dependencies:** T-03
- **DoD:** checklist มี expected result ครบและทดลองแล้วแดงเพราะ UI/client ยังไม่มี; ยืนยันว่า pages/components จะไม่เรียก `fetch` ตรง

### T-05 — Implement: frontend walking skeleton
- **สิ่งที่ทำ:** ตั้ง React/Vite/TypeScript, generate client จาก OpenAPI และสร้าง app shell/health state ให้ T-04 ผ่าน
- **Dependencies:** T-04
- **DoD:** `cd frontend; npm run lint; npm run build` ผ่าน; E2E T-04 ผ่าน; ค้นไม่พบ direct `fetch` ใน pages/components

## Phase 1 — Authentication และสิทธิ์

### T-06 — Tests: authentication API
- **สิ่งที่ทำ:** เขียน test `GET /api/v1/auth/me` และ `POST /api/v1/auth/logout` ครบ AC-01: identity, role, approved/pending/suspended, invalid/expired token และ sanitized errors
- **Dependencies:** T-05
- **DoD:** ครบทุกเงื่อนไข AC-01 และ status/error code ตาม contract; SQLite แยกต่อ test; `cd backend; go test ./...` แดง

### T-07 — Implement: authentication API
- **สิ่งที่ทำ:** ทำ auth handler/service/repo และ middleware ตาม contract ให้ T-06 ผ่าน
- **Dependencies:** T-06
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-08 — Tests: authorization policy
- **สิ่งที่ทำ:** เขียน test สิทธิ์ RESEARCHER/COORDINATOR/ADMIN และ pending/suspended ครบ AC-02–04 กับทุกกลุ่ม route ที่ contract กำหนด
- **Dependencies:** T-07
- **DoD:** ครบ read/write/admin-only/direct API denial และสิทธิ์หลังอนุมัติ; `cd backend; go test ./...` แดง

### T-09 — Implement: authorization policy
- **สิ่งที่ทำ:** บังคับ role/account status ใน service/middleware ให้ T-08 ผ่าน
- **Dependencies:** T-08
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-10 — Tests: authentication UI
- **สิ่งที่ทำ:** เขียน E2E checklist สำหรับ loading, authenticated, unauthenticated, pending, suspended, logout และเมนูตามบทบาท ครบ AC-01–04/25
- **Dependencies:** T-09
- **DoD:** ทุก state, keyboard/focus และ direct-navigation denial มี expected resultและทดลองแล้วแดง

### T-11 — Implement: authentication UI
- **สิ่งที่ทำ:** สร้าง auth bootstrap, account-state pages, logout และ role-aware navigation ผ่าน generated client
- **Dependencies:** T-10
- **DoD:** frontend lint/build และ E2E T-10 ผ่าน; ไม่มี direct `fetch`

### T-12 — Tests: user management API
- **สิ่งที่ทำ:** เขียน test list/detail/update/status ของ `/api/v1/users` ครบ AC-02/05 รวม approval, role change, duplicate email, suspension, invalid transition, not found และ ADMIN-only
- **Dependencies:** T-11
- **DoD:** ครบทุกเงื่อนไข AC-02/05; SQLite แยกต่อ test; `cd backend; go test ./...` แดง

### T-13 — Implement: user management API
- **สิ่งที่ทำ:** ทำ handler/service/repo และ database constraints สำหรับบัญชีให้ T-12 ผ่าน
- **Dependencies:** T-12
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-14 — Tests: user management UI
- **สิ่งที่ทำ:** เขียน E2E checklist รายการอนุมัติ, รายการบัญชี, แก้ข้อมูล/บทบาท/สถานะ และ denied states ครบ AC-02/05/25
- **Dependencies:** T-13
- **DoD:** loading/error/empty/success, validation, confirmation, keyboard/focus ครบและทดลองแล้วแดง

### T-15 — Implement: user management UI
- **สิ่งที่ทำ:** สร้างหน้าอนุมัติและจัดการบัญชีผ่าน generated client
- **Dependencies:** T-14
- **DoD:** frontend lint/build และ E2E T-14 ผ่าน; ไม่มี direct `fetch`

## Phase 2 — Research

### T-16 — Tests: research list/search API
- **สิ่งที่ทำ:** เขียน test `GET /api/v1/researches` ครบ AC-03/04/13/14 รวม empty, duplicate titles, query `q`/`status`/`process`, combined filters, Unicode, no match, invalid query, permissions และ errors
- **Dependencies:** T-15
- **DoD:** ครบทุกเงื่อนไข AC-13/14 ฝั่งรายการและค้นหา; SQLite แยกต่อ test; `cd backend; go test ./...` แดง

### T-17 — Implement: research list/search API
- **สิ่งที่ทำ:** ทำ handler/service/repo สำหรับรายการ ค้นหา และกรอง โดยไม่ค้น field นอก contract
- **Dependencies:** T-16
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-18 — Tests: research list/search UI
- **สิ่งที่ทำ:** เขียน E2E checklist รายการ ค้นหา และกรองครบ AC-03/04/13/14/25 รวมชื่อซ้ำ งานต่อเนื่อง empty/no-result/error/retry/clear filters
- **Dependencies:** T-17
- **DoD:** loading/error/empty/success, responsive, keyboard/focus/contrast/reduced motion ครบและทดลองแล้วแดง

### T-19 — Implement: research list/search UI
- **สิ่งที่ทำ:** สร้างหน้ารายการ ค้นหา และกรองผ่าน generated client
- **Dependencies:** T-18
- **DoD:** frontend lint/build และ E2E T-18 ผ่าน; ไม่มี direct `fetch`

### T-20 — Tests: research detail API
- **สิ่งที่ทำ:** เขียน test `GET /api/v1/researches/{id}` ครบ AC-03/04/07–10/13 รวมทุก field, members, contract metadata, continuation link, current status/process, invalid ID, not found, permissions และ errors
- **Dependencies:** T-19
- **DoD:** ครบทุกเงื่อนไข AC ที่ระบุสำหรับรายละเอียด; SQLite แยกต่อ test; `cd backend; go test ./...` แดง

### T-21 — Implement: research detail API
- **สิ่งที่ทำ:** ทำ handler/service/repo สำหรับรายละเอียดให้ T-20 ผ่าน
- **Dependencies:** T-20
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-22 — Tests: research detail UI
- **สิ่งที่ทำ:** เขียน E2E checklist หน้ารายละเอียดครบ AC-03/04/07–10/13/25 รวมข้อมูลครบ ชื่อซ้ำ งานต่อเนื่อง status/process, denied, not found และ retry
- **Dependencies:** T-21
- **DoD:** loading/error/success, responsive, keyboard/focus/contrast/reduced motion ครบและทดลองแล้วแดง

### T-23 — Implement: research detail UI
- **สิ่งที่ทำ:** สร้างหน้ารายละเอียดผ่าน generated client
- **Dependencies:** T-22
- **DoD:** frontend lint/build และ E2E T-22 ผ่าน; ไม่มี direct `fetch`

### T-24 — Tests: research navigation
- **สิ่งที่ทำ:** เขียน E2E checklist การเลือกแถวจากรายการไปยังรายละเอียดและย้อนกลับ โดยชื่อซ้ำต้องเปิดรายการที่เลือกถูกต้องตาม AC-13/25
- **Dependencies:** T-23
- **DoD:** loading/error/success, keyboard/focus และ browser history ครบและทดลองแล้วแดง

### T-25 — Implement: research navigation
- **สิ่งที่ทำ:** เชื่อม route ระหว่างรายการกับรายละเอียดผ่าน internal ID โดยไม่แสดงหรือค้นหาด้วย ID
- **Dependencies:** T-24
- **DoD:** frontend lint/build และ E2E T-24 ผ่าน; ไม่มี direct `fetch`

### T-26 — Tests: create research API
- **สิ่งที่ทำ:** เขียน multipart/DB/filesystem tests ของ `POST /api/v1/researches` ครบ AC-06–12/24: ทุก field, ID immutable/non-reuse/concurrency, root/continuation/title rules, members, dates/amounts, PDF validation/limits, uniqueness, defaults, atomic failure และ permissions
- **Dependencies:** T-25
- **DoD:** ทุกเงื่อนไข AC-06–12 ฝั่ง API/DB ครบ; SQLite หนึ่งฐานต่อ testและ storage แยก; `cd backend; go test ./...` แดง

### T-27 — Implement: create research API
- **สิ่งที่ทำ:** ทำ handler/service/repo/schema/file store สำหรับ create ให้ T-26 ผ่าน
- **Dependencies:** T-26
- **DoD:** `cd backend; go test ./...` ผ่าน; ไม่มี partial row/file เมื่อ failure

### T-28 — Tests: create research UI
- **สิ่งที่ทำ:** เขียน E2E checklist ฟอร์ม create ครบ AC-06–12/25 รวม continuation picker, validation ทุก field, PDF, duplicate errors, submit once, retain-on-error, cancel และ success return
- **Dependencies:** T-27
- **DoD:** loading/error/empty/success และ accessibility states ครบ; ทดลองกับ backend/SQLite แยกแล้วแดง

### T-29 — Implement: create research UI
- **สิ่งที่ทำ:** สร้างฟอร์ม create multipart ผ่าน generated client ให้ T-28 ผ่าน
- **Dependencies:** T-28
- **DoD:** frontend lint/build และ E2E T-28 ผ่านกับ backend จริง; ไม่มี direct `fetch`

### T-30 — Tests: contract file API
- **สิ่งที่ทำ:** เขียน test `GET /api/v1/researches/{id}/contract-file` ครบ AC-09: bytes/media type/filename, permission, missing research/file, invalid ID และไม่เปิดเผย storage path
- **Dependencies:** T-29
- **DoD:** ครบ AC-09 ฝั่ง download; `cd backend; go test ./...` แดง

### T-31 — Implement: contract file API
- **สิ่งที่ทำ:** ทำ file download handler/service/repo lookup ให้ T-30 ผ่าน
- **Dependencies:** T-30
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-32 — Tests: contract file UI
- **สิ่งที่ทำ:** เขียน E2E checklist เปิด/ดาวน์โหลด PDF จาก detail ครบ loading/error/success และสิทธิ์ตาม AC-09/25
- **Dependencies:** T-31
- **DoD:** expected result ครบและทดลองแล้วแดง

### T-33 — Implement: contract file UI
- **สิ่งที่ทำ:** เชื่อมการเปิดไฟล์สัญญาผ่าน generated client
- **Dependencies:** T-32
- **DoD:** frontend lint/build และ E2E T-32 ผ่าน; ไม่มี direct `fetch`

### T-34 — Tests: update research API
- **สิ่งที่ทำ:** เขียน test `PUT /api/v1/researches/{id}` ครบ AC-07–12/15/24 รวม immutable fields, full replacement, optional PDF, conflicts, validation, atomic rollback, old/new file cleanup, concurrency และ permissions
- **Dependencies:** T-33
- **DoD:** ทุกเงื่อนไข AC-15 และกฎข้อมูลที่เกี่ยวข้องครบ; `cd backend; go test ./...` แดง

### T-35 — Implement: update research API
- **สิ่งที่ทำ:** ทำ update handler/service/repo/file compensation ให้ T-34 ผ่าน
- **Dependencies:** T-34
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-36 — Tests: update research UI
- **สิ่งที่ทำ:** เขียน E2E checklist edit form ครบ AC-15/25: preload, immutable fields, keep/replace PDF, validation, cancel, retain-on-error, reload success และ permissions
- **Dependencies:** T-35
- **DoD:** loading/error/success และ accessibility states ครบ; ทดลองแล้วแดง

### T-37 — Implement: update research UI
- **สิ่งที่ทำ:** สร้างหน้าแก้ไขผ่าน generated client ให้ T-36 ผ่าน
- **Dependencies:** T-36
- **DoD:** frontend lint/build และ E2E T-36 ผ่าน; ไม่มี direct `fetch`

### T-38 — Tests: delete research API
- **สิ่งที่ทำ:** เขียน test `DELETE /api/v1/researches/{id}` ครบ AC-16/24: success, not found, continuation restriction, body/query validation, permissions, transaction/file rollback, cleanup และ ID non-reuse
- **Dependencies:** T-37
- **DoD:** ทุกเงื่อนไข AC-16 ฝั่ง API/DB/filesystem ครบ; `cd backend; go test ./...` แดง

### T-39 — Implement: delete research API
- **สิ่งที่ทำ:** ทำ atomic delete และ staged-file recovery ให้ T-38 ผ่าน
- **Dependencies:** T-38
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-40 — Tests: delete research UI
- **สิ่งที่ทำ:** เขียน E2E checklist delete confirmation/cancel/success/conflict/error/permissions ครบ AC-16/25
- **Dependencies:** T-39
- **DoD:** loading/error/empty-after-delete/success และ focus return ครบ; ทดลองแล้วแดง

### T-41 — Implement: delete research UI
- **สิ่งที่ทำ:** เชื่อม delete flow ผ่าน generated client ให้ T-40 ผ่าน
- **Dependencies:** T-40
- **DoD:** frontend lint/build และ E2E T-40 ผ่าน; ไม่มี direct `fetch`

### T-42 — Tests: status API
- **สิ่งที่ทำ:** เขียน test `PATCH /api/v1/researches/{id}/status` ครบ AC-17/18/24: six values, next-step/direct terminal, no skip/reverse, idempotency, terminal lock, unchanged process/ID, malformed/media/size/errors/concurrency/permissions
- **Dependencies:** T-41
- **DoD:** ทุก transition และทุกเงื่อนไข AC-17/18 ครบทั้ง service/DB; `cd backend; go test ./...` แดง

### T-43 — Implement: status API
- **สิ่งที่ทำ:** ทำ status handler/service/repo และ DB trigger ให้ T-42 ผ่าน
- **Dependencies:** T-42
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-44 — Tests: status UI
- **สิ่งที่ทำ:** เขียน E2E checklist แสดง/เปลี่ยน status ครบ AC-17/18/25 รวม old→new confirmation, allowed choices, conflict/error, terminal และ permissions
- **Dependencies:** T-43
- **DoD:** loading/error/success และ accessibility states ครบ; ทดลองแล้วแดง

### T-45 — Implement: status UI
- **สิ่งที่ทำ:** สร้าง status tracking/update UI ผ่าน generated client
- **Dependencies:** T-44
- **DoD:** frontend lint/build และ E2E T-44 ผ่าน; ไม่มี direct `fetch`

### T-46 — Tests: process API
- **สิ่งที่ทำ:** เขียน test `PATCH /api/v1/researches/{id}/process` ครบ AC-19/20/24: eight values, next-only, no skip/reverse/beyond-final, idempotency, terminal lock, unchanged status/ID, malformed/media/size/errors/concurrency/permissions
- **Dependencies:** T-45
- **DoD:** ทุก transition และทุกเงื่อนไข AC-19/20 ครบทั้ง service/DB; `cd backend; go test ./...` แดง

### T-47 — Implement: process API
- **สิ่งที่ทำ:** ทำ process handler/service/repo และ DB trigger ให้ T-46 ผ่าน
- **Dependencies:** T-46
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-48 — Tests: process UI
- **สิ่งที่ทำ:** เขียน E2E checklist แสดง/เปลี่ยน process ครบ AC-19/20/25 รวม current/next, conflict/error, final/terminal และ permissions
- **Dependencies:** T-47
- **DoD:** loading/error/success และ accessibility states ครบ; ทดลองแล้วแดง

### T-49 — Implement: process UI
- **สิ่งที่ทำ:** สร้าง process tracking/update UI ผ่าน generated client
- **Dependencies:** T-48
- **DoD:** frontend lint/build และ E2E T-48 ผ่าน; ไม่มี direct `fetch`

## Phase 3 — Admin modules

### T-50 — Tests: academic titles API
- **สิ่งที่ทำ:** เขียน test list/create/update/delete ครบ AC-21 รวม duplicate, in-use, not found, validation, permissions และข้อมูลผู้ใช้เดิมไม่เสีย
- **Dependencies:** T-49
- **DoD:** ครบทุกเงื่อนไข AC-21; `cd backend; go test ./...` แดง

### T-51 — Implement: academic titles API
- **สิ่งที่ทำ:** ทำ handler/service/repo/schema ให้ T-50 ผ่าน
- **Dependencies:** T-50
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-52 — Tests: academic titles UI
- **สิ่งที่ทำ:** เขียน E2E checklist list/create/edit/delete/use-as-option ครบ AC-21/25
- **Dependencies:** T-51
- **DoD:** loading/error/empty/success, confirmation และ accessibility ครบ; ทดลองแล้วแดง

### T-53 — Implement: academic titles UI
- **สิ่งที่ทำ:** สร้างหน้าตำแหน่งทางวิชาการผ่าน generated client
- **Dependencies:** T-52
- **DoD:** frontend lint/build และ E2E T-52 ผ่าน

### T-54 — Tests: about API
- **สิ่งที่ทำ:** เขียน test `GET/PUT /api/v1/about` ครบส่วน About ของ AC-22 รวม latest value, validation, atomic failure และ ADMIN-only update
- **Dependencies:** T-53
- **DoD:** ครบทุกเงื่อนไข About; `cd backend; go test ./...` แดง

### T-55 — Implement: about API
- **สิ่งที่ทำ:** ทำ handler/service/repo ให้ T-54 ผ่าน
- **Dependencies:** T-54
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-56 — Tests: about UI
- **สิ่งที่ทำ:** เขียน E2E checklist public display/admin edit ครบส่วน About ของ AC-22/25
- **Dependencies:** T-55
- **DoD:** loading/error/empty/success และ accessibility ครบ; ทดลองแล้วแดง

### T-57 — Implement: about UI
- **สิ่งที่ทำ:** สร้างหน้า About และ admin editor ผ่าน generated client
- **Dependencies:** T-56
- **DoD:** frontend lint/build และ E2E T-56 ผ่าน

### T-58 — Tests: organization API
- **สิ่งที่ทำ:** เขียน test `GET/PUT /api/v1/organization` ครบส่วน Organization ของ AC-22 รวมทุก field, latest value, validation, atomic failure และ ADMIN-only update
- **Dependencies:** T-57
- **DoD:** ครบทุกเงื่อนไข Organization; `cd backend; go test ./...` แดง

### T-59 — Implement: organization API
- **สิ่งที่ทำ:** ทำ handler/service/repo ให้ T-58 ผ่าน
- **Dependencies:** T-58
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-60 — Tests: organization UI
- **สิ่งที่ทำ:** เขียน E2E checklist published display/admin edit ครบส่วน Organization ของ AC-22/25
- **Dependencies:** T-59
- **DoD:** loading/error/empty/success และ accessibility ครบ; ทดลองแล้วแดง

### T-61 — Implement: organization UI
- **สิ่งที่ทำ:** สร้างหน้า organization และ admin editor ผ่าน generated client
- **Dependencies:** T-60
- **DoD:** frontend lint/build และ E2E T-60 ผ่าน

### T-62 — Tests: audit log API
- **สิ่งที่ทำ:** เขียน test การสร้างและ `GET /api/v1/audit-logs` ครบ AC-23: who/what/target/time, filters, ordering, immutable/no write route, ADMIN-only และ validation
- **Dependencies:** T-61
- **DoD:** ครบทุกเหตุการณ์ mutation ที่ SPEC กำหนดให้ตรวจย้อนหลัง; `cd backend; go test ./...` แดง

### T-63 — Implement: audit log API
- **สิ่งที่ทำ:** บันทึก audit แบบ atomic กับ mutation และทำ read-only handler/service/repo ให้ T-62 ผ่าน
- **Dependencies:** T-62
- **DoD:** `cd backend; go test ./...` ผ่าน

### T-64 — Tests: audit log UI
- **สิ่งที่ทำ:** เขียน E2E checklist รายงาน/filter/detail/permission ครบ AC-23/25
- **Dependencies:** T-63
- **DoD:** loading/error/empty/success และ accessibility ครบ; ทดลองแล้วแดง

### T-65 — Implement: audit log UI
- **สิ่งที่ทำ:** สร้างหน้ารายงาน Log แบบ read-only ผ่าน generated client
- **Dependencies:** T-64
- **DoD:** frontend lint/build และ E2E T-64 ผ่าน

## Phase 4 — Hardening

### T-66 — Hardening: quality gates บนเครื่อง
- **สิ่งที่ทำ:** รันและแก้ผลตรวจ backend/frontend/security พร้อม regression ทุก AC โดยไม่เปลี่ยน contract หรือขอบเขต
- **Dependencies:** T-65
- **DoD:** `cd backend; go test ./...; golangci-lint run; gosec ./...; govulncheck ./...` ผ่าน; `cd frontend; npm run lint; npm run build` ผ่าน; E2E ทั้งหมดผ่าน

### T-67 — Hardening: CI
- **สิ่งที่ทำ:** เพิ่ม GitHub Actions สำหรับ backend tests/lint/security, frontend lint/build และ contract validation
- **Dependencies:** T-66
- **DoD:** workflow รันบน clean checkout และผ่านทุก job; failure ใดทำให้ workflow แดง; ไม่มี secret ฝังใน repository

### T-68 — Hardening: cross-agent review
- **สิ่งที่ทำ:** เปิด session ใหม่ให้ agent ที่ไม่มีบริบทเดิม review SPEC/OpenAPI/schema/backend/frontend/tests และ trace AC-01–26
- **Dependencies:** T-67
- **DoD:** มีรายงาน findings พร้อม severity/evidence; แก้ finding ที่กระทบ acceptance แล้วรัน T-66/T-67 ซ้ำผ่าน; ไม่มี AC หรือเงื่อนไขที่ไม่มี test รองรับ
