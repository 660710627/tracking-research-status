# SPEC Tracking-Research-Status MVP
## Users
- Admin: เพิ่มงานวิจัย, ดูรายการงานวิจัยทั้งหมด
- Member: ดูรายการงานวิจัยทั้งหมด
- ทั้งสองบทบาทเห็นรายการวิจัยเดียวกัน ไม่มีมุมมองแยกตามบทบาท

## Out of scope (MVP v1)
- Authentication / user accounts
- การปรับสถานะของงานวิจัย
- มุมมองแยกตามบทบาท 
- Deployment

## Acceptance Criteria

AC-1: ระบบสามารถเพิ่มงานวิจัยได้
- ระบบได้รับคำขอ POST /api/v1/researches ที่มี Content-Type เป็น application/json พร้อม title และ description ที่ถูกต้อง
  → 201 พร้อมข้อมูล id, title และ description
- ระบบตัดช่องว่างหัวท้ายของ title และ description ก่อนบันทึก
- คำขอไม่มี title หรือ description
  → 422 พร้อม code "VALIDATION_ERROR"
- title หรือ description มีค่าเป็น null
  → 422 พร้อม code "VALIDATION_ERROR"
- title หรือ description ไม่ใช่ข้อมูลชนิด string
  → 422 พร้อม code "VALIDATION_ERROR"
- title หรือ description เป็นข้อความว่างหรือมีเฉพาะช่องว่าง
  → 422 พร้อม code "VALIDATION_ERROR"
- request body ไม่ใช่ JSON ที่ถูกต้อง
  → 400 พร้อม code "INVALID_JSON"
- Content-Type ไม่ใช่ application/json
  → 415 พร้อม code "UNSUPPORTED_MEDIA_TYPE"
- request body มี id หรือฟิลด์อื่นนอกเหนือจาก title และ description
  → 422 พร้อม code "VALIDATION_ERROR"

AC-2: ระบบสร้างรหัสประจำตัวของงานวิจัยโดยอัตโนมัติ
- ระบบสร้าง id ให้อัตโนมัติเมื่อเพิ่มงานวิจัย โดยผู้ใช้ไม่ต้องและไม่สามารถกำหนด id เอง
- id เป็นข้อมูลชนิด integer ที่มีค่าระหว่าง 100000 ถึง 999999
- id ไม่มีเลขศูนย์นำหน้า
- งานวิจัยแต่ละรายการมี id ที่ไม่ซ้ำกัน
- ฐานข้อมูลกำหนด id เป็น PRIMARY KEY
- ฐานข้อมูลกำหนด CHECK constraint ให้ id อยู่ระหว่าง 100000 ถึง 999999
- เมื่อระบบสร้าง id ที่ซ้ำกับรายการเดิม ระบบจะสร้าง id ใหม่ก่อนบันทึก
- เมื่อระบบไม่สามารถสร้าง id ที่ไม่ซ้ำได้
  → 503 พร้อม code "ID_GENERATION_FAILED"
- เมื่อเพิ่มงานวิจัยสำเร็จ
  → 201 พร้อม id ที่ระบบสร้างขึ้น

AC-3: ระบบสามารถแสดงรายการงานวิจัยทั้งหมดได้
- ระบบได้รับคำขอ GET /api/v1/researches
  → 200 พร้อม JSON array ที่มีรายการงานวิจัยทั้งหมด
- งานวิจัยแต่ละรายการประกอบด้วย id, title และ description
- id ใน response เป็น integer จำนวน 6 หลัก
- ระบบเรียงรายการตาม id จากน้อยไปมาก
- Admin และ Member เห็นรายการงานวิจัยชุดเดียวกัน
- ระบบยังไม่มีงานวิจัย
  → 200 พร้อมรายการว่าง []
- Endpoint นี้ไม่ต้องรับ request body

AC-4: ระบบสามารถแสดงรายละเอียดงานวิจัยได้
- ระบบได้รับคำขอ GET /api/v1/researches/{id} ด้วย id ที่เป็นตัวเลข 6 หลักและมีอยู่
  → 200 พร้อมข้อมูล id, title และ description
- ระบบได้รับ id ที่มีรูปแบบถูกต้องแต่ไม่พบงานวิจัย
  → 404 พร้อม code "RESEARCH_NOT_FOUND"
- id มีตัวอักษรหรืออักขระที่ไม่ใช่ตัวเลข
  → 422 พร้อม code "VALIDATION_ERROR"
- id มีจำนวนน้อยกว่าหรือมากกว่า 6 หลัก
  → 422 พร้อม code "VALIDATION_ERROR"
- id มีค่าต่ำกว่า 100000 หรือสูงกว่า 999999
  → 422 พร้อม code "VALIDATION_ERROR"
- Endpoint นี้ไม่ต้องรับ request body

## Error format (ทั้งระบบ)
{"error": {"code": "MACHINE_READABLE_CODE", "message": "human readable"}}

## API Contract
GET  /health                              → 200 {"status":"ok"}
GET  /api/v1/researches                   → 200 [{id,title,description}]
GET  /api/v1/researches/{id}              → 200 | 404 RESEARCH_NOT_FOUND | 422 VALIDATION_ERROR
POST /api/v1/researches {title,description}
                                           → 201 | 400 INVALID_JSON | 415 UNSUPPORTED_MEDIA_TYPE
                                                 | 422 VALIDATION_ERROR | 503 ID_GENERATION_FAILED