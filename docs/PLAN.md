# PLAN — Tracking Research Status MVP

## หลักการ

- `docs/SPEC.md` กำหนดขอบเขตและกฎธุรกิจ; `docs/openapi.yaml` เป็นสัญญา API ที่ implementation ต้องทำตาม
- Backend แยกเป็น Gin handler → service → repo → SQLite: handler จัดการ HTTP, service บังคับกฎธุรกิจและ typed errors, repo ทำเฉพาะ SQL
- Frontend แยก React pages → components → generated typed API client; pages/components ห้ามเรียก `fetch` โดยตรง
- กฎสำคัญเรื่องสิทธิ์, ID, uniqueness, relation และ status/process transition ต้องบังคับที่ service และ database constraint/trigger ตาม SPEC

## ลำดับดำเนินงาน

1. **Contract และ schema** — ตรวจความสอดคล้องของ SPEC/OpenAPI, สร้าง migration แบบไม่ล้างข้อมูล, เปิด foreign keys และกำหนด constraint/index/trigger สำหรับ ID, งานต่อเนื่อง, contract number, status/process และ transaction
2. **Backend foundation** — เชื่อม `library.db`, health check, router, request limits และ error envelope กลาง โดย map typed errors เป็น HTTP status/error code ตาม OpenAPI
3. **Research backend** — ทำ list/create/update/delete และ status/process ผ่าน handler/service/repo; จัดการ PDF ด้วย staging, validation, atomic commit, compensation และ startup recovery
4. **Generated client** — generate TypeScript client จาก `docs/openapi.yaml` และห้ามแก้ generated files ด้วยมือ
5. **Research frontend** — สร้างหน้า list/detail/form และการติดตาม status/process ด้วย components ที่ใช้ typed client; รองรับ loading, empty, error, success, validation, duplicate-submit prevention และ confirmation
6. **Authentication และสิทธิ์** — ก่อนทำ AC-01–05 และ endpoint ที่มีสิทธิ์ ต้องกำหนดระบบออก Access Token และปรับ OpenAPI ให้ตรง SPEC แล้วขออนุมัติ contract ก่อน implement
7. **Admin modules** — หลัง contract ครบ จึงทำ users, academic titles, about, organization และ audit logs ตามสิทธิ์ ADMIN และ AC-21–23
8. **Verification** — ตรวจ backend, frontend, accessibility และ security ก่อนส่งมอบ; deployment, notification และ Excel export ไม่อยู่ใน MVP v1

## การทดสอบ

- เขียน test ก่อนหรือพร้อม implementation; ห้ามแก้หรือลบ test เพื่อให้ผ่าน
- Backend ใช้ `go test`, `net/http/httptest` และ SQLite จาก `t.TempDir()` โดยหนึ่ง database ต่อหนึ่ง test
- ทดสอบ success, validation, permission, typed-error mapping, transaction rollback, concurrency, PDF cleanup/recovery และ route/method errors
- Frontend ตรวจ lint/build และ flow จริงของ loading, empty, error, success, keyboard, focus, contrast, responsive และ reduced motion

## Quality gates

- `cd backend; go test ./...`
- `cd backend; golangci-lint run; gosec ./...; govulncheck ./...`
- `cd frontend; npm run lint; npm run build`
- ห้าม implement endpoint ที่ยังไม่มีใน OpenAPI; หาก SPEC กับ OpenAPI ขัดกัน ให้แก้ contract และขออนุมัติก่อนเขียน feature
