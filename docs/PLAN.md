# PLAN Tracking-Research-Status MVP

## ขอบเขต
- รองรับ health check และการเพิ่ม ดูรายการ แก้ไข และลบงานวิจัยตาม `SPEC.md`
- งานวิจัยมีเฉพาะ `title` และ `description`; `title` ใช้ระบุรายการ
- ไม่รวม ID, authentication, การแยกสิทธิ์ตามบทบาท, การปรับสถานะ และ deployment

## Contract first
- จัดทำ `docs/openapi.yaml` ให้ตรงกับ endpoint, schema, status และ error code ใน `SPEC.md` ก่อน implementation
- ใช้ docs/openapi.yaml เป็นแหล่งความจริงเดียวและสร้าง typed API client สำหรับ frontend; ห้ามแก้ generated code ด้วยมือ
- หาก contract หรือ SPEC กำกวม ให้หยุดและถามก่อนเปลี่ยนขอบเขต

## Backend
- `db`: เปิด SQLite `library.db` และสร้าง schema พร้อม `NOT NULL`, validation constraints และ `UNIQUE(title)`
- `repo`: มีเฉพาะ SQL สำหรับเพิ่ม เรียงรายการตาม title ค้นหา แก้ไขแบบ transaction และลบ
- `service`: trim/validate ข้อมูล บังคับกฎ title ไม่ซ้ำ และคืน typed errors ตามกฎธุรกิจ
- `handler`: จัดการเฉพาะ HTTP เช่น Content-Type, body limit, JSON, path/query และแปลง typed errors เป็น status กับ error code ตาม SPEC
- `cmd/server`: ประกอบ dependencies, routes และเริ่ม Gin โดยไม่ใส่ business rules
- Error response ทั้งระบบใช้ `{"error":{"code":"...","message":"..."}}`

## Frontend
- ใช้ flow `pages → components → generated typed API client`
- Pages ควบคุม flow การโหลดรายการ เพิ่ม แก้ไข และลบ; components รับข้อมูลและ events ผ่าน typed props
- แสดง loading, empty, success และ error state โดยอ้างอิง error code จาก contract
- ห้าม pages หรือ components เรียก `fetch` โดยตรง

## Testing
- เขียน test ก่อนหรือพร้อม implementation และครอบคลุม success/error ทุกกรณีใน SPEC
- Database/repo test ใช้ SQLite ใหม่ใน `t.TempDir()` และหนึ่งฐานข้อมูลต่อหนึ่ง test; ทดสอบ constraints, sorting, transaction และ concurrent title conflict
- Service test ครอบคลุม trim, validation, duplicate title, update และ delete rules
- Handler test ใช้ `go test` กับ `net/http/httptest` ครบทุก endpoint, status, headers และ error format
- Frontend ตรวจ typed client integration, ESLint และ TypeScript/Vite build

## ลำดับงาน
1. ยืนยัน `docs/openapi.yaml` และ generated client contract
2. สร้าง schema และ repo พร้อม test
3. สร้าง service และ typed errors พร้อม test
4. สร้าง Gin handlers และ routes พร้อม `httptest`
5. ประกอบ server และตรวจ backend tests/lint/security
6. สร้าง frontend pages/components ผ่าน generated client
7. ตรวจ frontend lint/build และทดสอบ flow ตาม acceptance criteria

ดำเนินงานทีละหนึ่ง task ตาม `docs/TASKS.md` และไม่ทำเกินขอบเขต task
