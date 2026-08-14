# PLAN: Tracking Research Status MVP

## ขอบเขต

- รองรับ health check และ CRUD งานวิจัยด้วย `id` ที่ระบบสร้างและเปลี่ยนไม่ได้
- ข้อมูลงานวิจัยประกอบด้วย `id`, `title`, `description`, `continuationOfId`, `status` และ `process`
- รองรับงานวิจัยต่อเนื่อง ชื่อซ้ำตามกฎใน SPEC การปรับสถานะ และการเดินกระบวนการไปข้างหน้า
- ไม่รวม authentication, notification, deployment และหน้า UI สำหรับดูหรือปรับสถานะและกระบวนการ

## Contract first

- ปรับ API Contract และ SQL ใน SPEC ที่ยังไม่ตรงกับ AC ล่าสุดก่อนเริ่มโค้ด
- ทำ `docs/openapi.yaml` ให้ครอบคลุม endpoint, schema, status และ error code ตาม SPEC และใช้เป็นแหล่งความจริงเดียวของ API
- สร้าง typed API client จาก OpenAPI ใหม่เมื่อ contract เปลี่ยน และห้ามแก้ไฟล์ generated ด้วยมือ

## Backend

- **Gin handler:** รับผิดชอบเฉพาะ HTTP เช่น path/query, content type, จำกัดขนาด body, JSON decoding และแปลง typed error เป็น response
- **Service:** ตรวจ validation และกฎธุรกิจทั้งหมด เช่น immutable ID/continuation, ชื่อซ้ำ, การลบ parent, status transition, process transition และ terminal lock
- **Repo:** รับผิดชอบเฉพาะ SQL, transaction, atomic update และแปลงข้อผิดพลาดจากฐานข้อมูลเป็น typed persistence error
- **SQLite:** บังคับ unique/non-reused ID, foreign key, delete restriction, enum และ invariant สำคัญด้วย constraint/index/trigger โดยเปิด foreign keys ทุก connection
- การสร้าง แก้ไข ลบ และ transition ต้อง atomic และปลอดภัยเมื่อมี request พร้อมกัน

## Frontend

- โครงสร้างเป็น React pages → components → generated typed API client
- หน้า CRUD ใช้ `id` ระบุรายการ แม้ชื่อซ้ำกันได้ และแสดง loading, error, empty และ success state
- component ห้ามเรียก `fetch` โดยตรง และห้ามสร้าง endpoint นอก OpenAPI
- ยังไม่สร้างหน้าดูหรือปรับ status/process ตาม Out of scope ของ MVP

## Error handling

- Repo ส่ง typed persistence error ให้ service แปลงเป็น typed domain error
- Handler map domain error เป็น HTTP status และ error code ตาม SPEC ด้วย response format เดียวกัน
- ไม่ส่งรายละเอียด SQL หรือ internal error ให้ client

## Testing

- แต่ละ slice เขียน test ให้แดงและครอบ AC ของ slice ก่อนเขียน implementation
- ใช้ `go test`, `httptest` และ SQLite จริง โดยใช้ `t.TempDir()` และฐานข้อมูลใหม่หนึ่งไฟล์ต่อหนึ่ง test
- Repo tests ครอบคลุม constraint, foreign key, ID ไม่ถูกนำกลับมาใช้, transaction และ concurrency
- Service tests ครอบคลุม validation, continuation/title rules และ transition matrix ของ status/process
- Handler tests ครอบคลุม request validation, response body, HTTP status และ error code ทุกกรณีใน SPEC
- Frontend ตรวจ typed-client integration และ E2E checklist สำหรับ loading, error, empty และ success

## ลำดับดำเนินงาน

1. ทำ API Contract, SQL ใน SPEC และ `docs/openapi.yaml` ให้ตรงกับ AC ล่าสุด
2. ทำ walking skeleton และตั้งค่า generated typed API client
3. ทำ schema/repo โดยเขียน test ก่อน implementation
4. ทำ CRUD และ continuation ทีละ slice โดยเขียน test ก่อน implementation
5. ทำ status transition โดยเขียน test ก่อน implementation
6. ทำ process transition โดยเขียน test ก่อน implementation
7. ทำหน้า CRUD และ UI states ผ่าน generated client
8. รัน quality gates, CI และ cross-agent review
