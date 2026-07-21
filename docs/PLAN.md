# PLAN Tracking-Research-Status MVP

## เป้าหมาย
- รองรับการเพิ่ม ดูรายการงานวิจัย
- ไม่รวม Authentication การเปลี่ยนสถานะ การแก้ไข การลบ และ Deployment
- ใช้ `docs/openapi.yaml` เป็น API contract และแหล่งความจริงเดียว

## Contract first
- จัดทำและตรวจสอบ `docs/openapi.yaml` ให้ตรงกับ endpoint, schema, status และ error code ใน `SPEC.md` ก่อนเริ่ม implementation
- กำหนด `id` เป็นข้อมูลชนิด string เป็นตัวเลข 6 ตัว และกำหนด error response แบบเดียวทั้งระบบ
- สร้าง typed API client ของ frontend จาก OpenAPI และห้ามแก้ generated code ด้วยมือ

## Backend
- `db`: เปิดการเชื่อมต่อ SQLite และสร้าง schema โดยใช้ `id` เป็น PRIMARY KEY พร้อม CHECK constraint สำหรับตัวอักษรที่เป็นตัวเลข 6 ตัว 
- `repo`: รู้เฉพาะ SQL สำหรับเพิ่ม เรียงรายการตาม `id` และค้นหาด้วย `id`
- `service`: ตรวจและ trim ข้อมูล สร้าง `id` ที่ไม่ซ้ำ จัดการ collision และคืน typed errors ตามกฎธุรกิจ
- `handler`: รู้เฉพาะ HTTP ตรวจ Content-Type/JSON เรียก service และแปลง typed errors เป็น HTTP status กับ error code ตาม SPEC
- `cmd/server`: ประกอบ dependency และเปิด Gin server โดยไม่ใส่กฎธุรกิจ

## Frontend
- React pages เป็นหน้าจอหลักและควบคุม flow การเพิ่ม ดูรายการ 
- Components รับข้อมูลและ event ผ่าน typed props และใช้งาน generated typed API client ตาม flow `pages → components → generated client`
- ห้าม pages หรือ components เรียก `fetch` โดยตรง
- แสดง validation และ API errors จาก error code โดยไม่ผูก logic กับข้อความ error

## Testing
- เขียน test ก่อนหรือพร้อม implementation ด้วย `go test` และ `net/http/httptest`
- ทุก test ที่ใช้ SQLite ต้องสร้างฐานข้อมูลใหม่ใน `t.TempDir()` และห้ามใช้ฐานข้อมูลร่วมกันระหว่าง test
- ทดสอบ schema constraints, repo queries, การเรียงรายการ, การสร้างและ collision ของ `id`, validation และ typed errors
- ทดสอบ handler ครบทุก success/error response ใน SPEC รวมถึง Content-Type, JSON ผิดรูปแบบ และ `id` ไม่ถูกต้อง
- ตรวจ frontend ด้วย ESLint, TypeScript build และการใช้งาน generated client

## ลำดับงาน
1. จัดทำและยืนยัน `docs/openapi.yaml`
2. สร้าง schema และ repo พร้อม test
3. สร้าง service พร้อม test กฎธุรกิจ
4. สร้าง Gin handler พร้อม `httptest`
5. สร้าง generated typed client, pages และ components
6. รัน backend tests/lint/security และ frontend lint/build
