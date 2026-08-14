# SPEC Tracking-Research-Status MVP
## Users
- ผู้ดูแลระบบ: เพิ่มงานวิจัย, ดูรายการงานวิจัยทั้งหมด, ลบงานวิจัย, แก้ไขข้อมูล
- ผู้ประสานงาน: เพิ่มงานวิจัย, ดูรายการงานวิจัยทั้งหมด, ลบงานวิจัย, แก้ไขข้อมูล
- นักวิจัย: ดูรายการงานวิจัยทั้งหมด

## Out of scope (MVP v1)
- Authentication / user accounts
- Deployment
- Notificaton การแจ้งเตือนให้กับนักวิจัยเมื่องานวิจัยของตนถูกเปลี่ยนสถานะ
- หน้าสำหรับดูสถานะและกระบวนการ
- การส่งออกข้อมูลในรูปแบบ Excel

## Acceptance Criteria

AC-1: ระบบสามารถเพิ่มงานวิจัยได้
- เมื่อเรียก POST /api/v1/researches ด้วย title, description และ continuationOfId ที่ถูกต้อง
  → ตอบ 201 พร้อม id, title, description, continuationOfId, status และ process
- Request body ต้องมีฟิลด์ title, description และ continuationOfId ครบทุกฟิลด์ และห้ามมีฟิลด์อื่น
- งานวิจัยต้นฉบับต้องส่ง continuationOfId เป็น null; ห้ามละฟิลด์นี้
- continuationOfId ต้องเป็น null หรือ integer บวกที่อ้างถึงงานวิจัยที่มีอยู่
- หาก continuationOfId อ้างถึงงานวิจัยที่ไม่มีอยู่
  → ตอบ 404 พร้อม code CONTINUATION_NOT_FOUND
- ห้าม client ส่ง id, status หรือ process
  → ตอบ 422 พร้อม code VALIDATION_ERROR
- ระบบสร้าง id เป็น integer บวกที่ไม่ซ้ำและแก้ไขไม่ได้
- id ต้องคงเดิมเมื่อแก้ไข title, description, status หรือ process
- ห้ามนำ id ของงานวิจัยที่ลบแล้วกลับมาใช้กับงานวิจัยใหม่
- ลำดับ id ไม่จำเป็นต้องต่อเนื่องและสามารถมีช่องว่างได้
- งานวิจัยใหม่ทุกงานมี status เริ่มต้นเป็น "กำลังดำเนินการ"
- งานวิจัยใหม่ทุกงานมี process เริ่มต้นเป็น "สัญญาโครงการ"
- "สัญญาโครงการ" ถูกนับเป็นกระบวนการแรกที่กำลังดำเนินการ ไม่ใช่ค่าก่อนเริ่มกระบวนการ
- Response ต้องตรงกับข้อมูลที่บันทึกในฐานข้อมูล
- การสร้าง id, การบันทึกข้อมูล, ความสัมพันธ์งานต่อเนื่อง, status และ process ต้องสำเร็จหรือล้มเหลวพร้อมกัน

AC-2: ระบบรองรับงานวิจัยต่อเนื่องและกฎ title ซ้ำ
- งานวิจัยต้นฉบับมี continuationOfId เป็น null
- งานวิจัยต่อเนื่องมี continuationOfId อ้างถึงงานวิจัยที่มีอยู่ก่อนแล้ว
- สามารถอ้างถึงงานวิจัยที่มี status เป็น "โครงการเสร็จสิ้น" หรือ "ยุติโครงการ" ได้
- continuationOfId แก้ไขไม่ได้หลังสร้าง เพื่อป้องกันการเปลี่ยนสายงานวิจัยและวงจรอ้างอิง
- งานวิจัยต้นฉบับห้ามใช้ title ที่ซ้ำกับงานวิจัยใดที่มีอยู่หลังตัด Unicode whitespace
  → ตอบ 409 พร้อม code TITLE_ALREADY_EXISTS
- งานวิจัยต่อเนื่องสามารถใช้ title ซ้ำกับงานวิจัยที่มีอยู่ได้
- งานวิจัยต่อเนื่องไม่จำเป็นต้องใช้ title เดียวกับงานที่อ้างถึง
- งานวิจัยต่อเนื่องหลายงานสามารถอ้างถึงงานวิจัยเดียวกันได้
- การตรวจ title ซ้ำเป็นแบบ case-sensitive หลังตัด Unicode whitespace
- ฐานข้อมูลต้องบังคับ foreign key และกฎ title ซ้ำให้ถูกต้องแม้มี concurrent requests

AC-3: ระบบตรวจสอบ title และ description
- ต้องมีฟิลด์ title และ description
- ทั้งสองค่าต้องเป็น string และห้ามเป็น null
- ระบบตัด Unicode whitespace ที่หัวและท้ายก่อน validation และบันทึก
- title หลังตัด whitespace ต้องมีความยาว 1–200 Unicode characters
- title ห้ามมี newline, tab, NUL, control characters และ /
- description หลังตัด whitespace ต้องมีความยาว 1–5,000 Unicode characters
- description อนุญาต newline และ tab
- description ห้ามมี NUL และ control characters อื่น
- หากไม่ผ่านเงื่อนไข
  → ตอบ 422 พร้อม code VALIDATION_ERROR

AC-4: ระบบตรวจสอบ request body
- POST, PUT และ PATCH ต้องใช้ Content-Type: application/json
- ยอมรับ parameter เช่น application/json; charset=utf-8
- การตรวจ media type ไม่สนใจตัวพิมพ์ใหญ่–เล็ก
- หากไม่มี Content-Type หรือเป็นชนิดอื่น
  → ตอบ 415 พร้อม code UNSUPPORTED_MEDIA_TYPE
- Body ว่างหรือมีเฉพาะ whitespace
  → ตอบ 400 พร้อม code INVALID_JSON
- JSON ไม่สมบูรณ์หรือ parse ไม่ได้
  → ตอบ 400 พร้อม code INVALID_JSON
- มีข้อมูลต่อท้าย JSON หรือมี JSON มากกว่าหนึ่งค่า
  → ตอบ 400 พร้อม code INVALID_JSON
- JSON ระดับบนสุดไม่ใช่ object
  → ตอบ 422 พร้อม code VALIDATION_ERROR
- JSON มี key ซ้ำหรือฟิลด์ที่ operation นั้นไม่รองรับ
  → ตอบ 422 พร้อม code VALIDATION_ERROR
- Request body มีขนาดเกิน 64 KiB
  → ตอบ 413 พร้อม code PAYLOAD_TOO_LARGE

AC-5: ระบบแสดงรายการงานวิจัยทั้งหมดได้
- เมื่อเรียก GET /api/v1/researches
  → ตอบ 200
- Response มี Content-Type: application/json
- Response เป็น JSON array
- แต่ละรายการมีเฉพาะ id, title, description, continuationOfId, status และ process
- ระบบเรียงรายการตาม title จากน้อยไปมาก และใช้ id จากน้อยไปมากเป็นลำดับรองเมื่อ title ซ้ำ
- หากไม่มีงานวิจัย
  → ตอบ 200 พร้อม []
- ผู้เรียกทุกคนเห็นรายการชุดเดียวกัน
- หากส่ง request body ที่ไม่ว่าง
  → ตอบ 400 พร้อม code INVALID_REQUEST_BODY
- หากส่ง query parameter ใด
  → ตอบ 422 พร้อม code VALIDATION_ERROR

AC-6: ระบบสามารถแก้ไข title และ description ของงานวิจัยได้
- เมื่อเรียก PUT /api/v1/researches/{id} โดย {id} เป็น integer บวก
- Request body ต้องมีเฉพาะ title และ description
- ห้ามเปลี่ยน id, continuationOfId, status หรือ process ผ่าน endpoint นี้
- การแก้ไขเป็นการแทน title และ description เดิมทั้งหมด
- เมื่อพบงานวิจัยและข้อมูลใหม่ถูกต้อง
  → ตอบ 200 พร้อมข้อมูลล่าสุดทุกฟิลด์
- หากไม่พบ id
  → ตอบ 404 พร้อม code RESEARCH_NOT_FOUND
- หาก path id ไม่ใช่ integer บวก
  → ตอบ 422 พร้อม code VALIDATION_ERROR
- งานวิจัยต้นฉบับที่ส่ง title เดิมสามารถแก้ไข description ได้ แม้มีงานวิจัยต่อเนื่องใช้ title เดียวกัน
- หากงานวิจัยต้นฉบับเปลี่ยนเป็น title อื่น ต้องไม่ซ้ำกับงานวิจัยใดที่มีอยู่
  → ตอบ 409 พร้อม code TITLE_ALREADY_EXISTS
- งานวิจัยต่อเนื่องสามารถเปลี่ยนไปใช้ title ที่ซ้ำได้
- การแก้ไข title และ description ต้องสำเร็จหรือล้มเหลวพร้อมกันใน transaction เดียว

AC-7: ระบบสามารถลบงานวิจัยได้
- เมื่อเรียก DELETE /api/v1/researches/{id} โดย {id} เป็น integer บวกและลบสำเร็จ
  → ตอบ 204 โดยไม่มี response body
- หลังลบ รายการต้องไม่ปรากฏใน GET /api/v1/researches
- หากไม่พบ id หรือเรียกลบ id เดิมซ้ำ
  → ตอบ 404 พร้อม code RESEARCH_NOT_FOUND
- หลังลบ id เดิมต้องไม่ถูกนำกลับมาใช้กับงานวิจัยใหม่
- หาก path id ไม่ใช่ integer บวก
  → ตอบ 422 พร้อม code VALIDATION_ERROR
- หากมีงานวิจัยอื่นอ้างถึงรายการนี้ผ่าน continuationOfId
  → ตอบ 409 พร้อม code RESEARCH_HAS_CONTINUATIONS
- ฐานข้อมูลต้องห้ามลบงานต้นทางที่ยังมีงานวิจัยต่อเนื่องอ้างถึง
- หากส่ง request body ที่ไม่ว่าง
  → ตอบ 400 พร้อม code INVALID_REQUEST_BODY
- หากส่ง query parameter ใด
  → ตอบ 422 พร้อม code VALIDATION_ERROR

AC-8: ระบบสามารถปรับสถานะของงานวิจัยได้
- เมื่อเรียก PATCH /api/v1/researches/{id}/status ต้องส่ง JSON object ที่มีเฉพาะ status
- สถานะที่รองรับตามลำดับปกติ ได้แก่
  1. กำลังดำเนินการ
  2. กำลังดำเนินการ (ขยายเวลาครั้งที่ 1)
  3. กำลังดำเนินการ (ขยายเวลาครั้งที่ 2)
  4. กำลังดำเนินการ (ขยายเวลามากกว่า 2 ครั้ง)
- สถานะปกติเปลี่ยนได้เฉพาะสถานะถัดไป ห้ามข้ามและห้ามย้อนกลับ
- จากสถานะปกติระดับใดก็ได้ สามารถข้ามไป "โครงการเสร็จสิ้น" หรือ "ยุติโครงการ" ได้
- "โครงการเสร็จสิ้น" และ "ยุติโครงการ" เป็น terminal status และหมายถึงโครงการจบแล้ว
- เมื่อเข้าสู่ terminal status ค่า process ปัจจุบันต้องคงเดิมและไม่ถูกนับว่าเสร็จโดยอัตโนมัติ
- หลังเข้าสู่ terminal status ห้ามเปลี่ยนไปสถานะอื่นและห้ามปรับ process
- การส่ง status เดียวกับค่าปัจจุบันตอบ 200 โดยไม่เปลี่ยนข้อมูล เพื่อรองรับ request ซ้ำ
- การส่ง terminal status เดิมซ้ำตอบ 200 โดยไม่เปลี่ยนข้อมูล
- หาก transition ข้ามหรือย้อนสถานะโดยไม่ได้ไป terminal status
  → ตอบ 409 พร้อม code INVALID_STATUS_TRANSITION
- หากโครงการจบแล้วและขอเปลี่ยนเป็นสถานะอื่น
  → ตอบ 409 พร้อม code PROJECT_ALREADY_ENDED
- หากไม่พบ id
  → ตอบ 404 พร้อม code RESEARCH_NOT_FOUND
- เมื่อปรับสำเร็จ
  → ตอบ 200 พร้อมข้อมูลล่าสุดทุกฟิลด์

AC-9: ระบบสามารถปรับกระบวนการของงานวิจัยได้
- เมื่อเรียก PATCH /api/v1/researches/{id}/process ต้องส่ง JSON object ที่มีเฉพาะ process
- กระบวนการที่รองรับตามลำดับ ได้แก่
  1. สัญญาโครงการ
  2. บันทึกข้อตกลง
  3. เปิดบัญชีธนาคาร
  4. การเบิกจ่ายเงิน
  5. การจัดสรรค่าธรรมเนียม
  6. การติดตามส่งรายงาน
  7. รายงานสรุปการใช้เงิน
  8. การปิดบัญชีธนาคาร
- ระบบมี current process เพียงหนึ่งค่า
- เปลี่ยนได้เฉพาะ process ถัดไป ห้ามข้ามและห้ามย้อนกลับ
- การเปลี่ยนไป process ถัดไปหมายถึง process ก่อนหน้าเสร็จแล้ว
- เมื่ออยู่ที่ "การปิดบัญชีธนาคาร" ถือว่าดำเนินกระบวนการครบทั้งหมดและห้ามเปลี่ยนต่อ
- การดำเนินกระบวนการครบไม่เปลี่ยน status เป็น "โครงการเสร็จสิ้น" โดยอัตโนมัติ
- การเปลี่ยน process ไม่เปลี่ยน status และการเปลี่ยน status ปกติไม่เปลี่ยน process
- การส่ง process เดียวกับค่าปัจจุบันตอบ 200 โดยไม่เปลี่ยนข้อมูล เพื่อรองรับ request ซ้ำ
- หากข้าม ย้อน หรือเปลี่ยนต่อจาก process สุดท้าย
  → ตอบ 409 พร้อม code INVALID_PROCESS_TRANSITION
- หาก status เป็น "โครงการเสร็จสิ้น" หรือ "ยุติโครงการ"
  → ตอบ 409 พร้อม code PROJECT_ALREADY_ENDED
- หากไม่พบ id
  → ตอบ 404 พร้อม code RESEARCH_NOT_FOUND
- เมื่อปรับสำเร็จ
  → ตอบ 200 พร้อมข้อมูลล่าสุดทุกฟิลด์

AC-10: ระบบรักษาความถูกต้องเมื่อมี concurrent requests และใช้ error format เดียวกัน
- Error response ทั้งระบบใช้รูปแบบ {"error":{"code":"...","message":"..."}}
- ระบบต้องไม่เปิดเผย stack trace หรือรายละเอียดภายในฐานข้อมูล
- การสร้าง id พร้อมกันต้องได้ id ที่ไม่ซ้ำ
- การสร้างงานวิจัยต้นฉบับ title เดียวกันพร้อมกันต้องสำเร็จเพียงหนึ่ง request
- การปรับ status หรือ process พร้อมกันต้องตรวจค่าปัจจุบันและบันทึก transition แบบ atomic
- มีเพียง transition ที่ถูกต้องจากค่าล่าสุดเท่านั้นที่สำเร็จ
- การเปลี่ยนเข้าสู่ terminal status และการล็อกไม่ให้ปรับ status/process ต่อ ต้องเกิดแบบ atomic
- Database failure
  → ตอบ 500 พร้อม code INTERNAL_ERROR
- Unknown route
  → ตอบ 404 พร้อม code ROUTE_NOT_FOUND
- Method ไม่ตรงกับ endpoint
  → ตอบ 405 พร้อม code METHOD_NOT_ALLOWED

## Error format (ทั้งระบบ)
{"error": {"code": "MACHINE_READABLE_CODE", "message": "human readable"}}

## API Contract
GET  /health
     → 200 {"status":"ok"}
     → 503 SERVICE_UNAVAILABLE
     → 500 INTERNAL_ERROR

GET  /api/v1/researches
     → 200 [{id,title,description,continuationOfId,status,process}]
     → 400 INVALID_REQUEST_BODY
     → 422 VALIDATION_ERROR
     → 500 INTERNAL_ERROR

POST /api/v1/researches {title,description,continuationOfId}
     → 201 {id,title,description,continuationOfId,status,process}
     → 400 INVALID_JSON
     → 404 CONTINUATION_NOT_FOUND
     → 409 TITLE_ALREADY_EXISTS
     → 413 PAYLOAD_TOO_LARGE
     → 415 UNSUPPORTED_MEDIA_TYPE
     → 422 VALIDATION_ERROR
     → 500 INTERNAL_ERROR

PUT  /api/v1/researches/{id} {title,description}
     → 200 {id,title,description,continuationOfId,status,process}
     → 400 INVALID_JSON
     → 404 RESEARCH_NOT_FOUND
     → 409 TITLE_ALREADY_EXISTS
     → 413 PAYLOAD_TOO_LARGE
     → 415 UNSUPPORTED_MEDIA_TYPE
     → 422 VALIDATION_ERROR
     → 500 INTERNAL_ERROR

DELETE /api/v1/researches/{id}
       → 204
       → 400 INVALID_REQUEST_BODY
       → 404 RESEARCH_NOT_FOUND
       → 409 RESEARCH_HAS_CONTINUATIONS
       → 422 VALIDATION_ERROR
       → 500 INTERNAL_ERROR

PATCH /api/v1/researches/{id}/status {status}
      → 200 {id,title,description,continuationOfId,status,process}
      → 400 INVALID_JSON
      → 404 RESEARCH_NOT_FOUND
      → 409 INVALID_STATUS_TRANSITION | PROJECT_ALREADY_ENDED
      → 413 PAYLOAD_TOO_LARGE
      → 415 UNSUPPORTED_MEDIA_TYPE
      → 422 VALIDATION_ERROR
      → 500 INTERNAL_ERROR

PATCH /api/v1/researches/{id}/process {process}
      → 200 {id,title,description,continuationOfId,status,process}
      → 400 INVALID_JSON
      → 404 RESEARCH_NOT_FOUND
      → 409 INVALID_PROCESS_TRANSITION | PROJECT_ALREADY_ENDED
      → 413 PAYLOAD_TOO_LARGE
      → 415 UNSUPPORTED_MEDIA_TYPE
      → 422 VALIDATION_ERROR
      → 500 INTERNAL_ERROR

ทุก Endpoint
     → 404 ROUTE_NOT_FOUND
     → 405 METHOD_NOT_ALLOWED

## SQL
- ตาราง `researches` ต้องมี `id`, `title`, `description`, `continuation_of_id`, `status` และ `process`
- `id` เป็น integer บวกที่ SQLite สร้าง เป็น primary key แบบไม่ใช้ค่าซ้ำหลังลบ และห้ามแก้ไข
- `continuation_of_id` เป็น nullable self-reference foreign key; ห้ามแก้ไข และใช้ `ON UPDATE RESTRICT` กับ `ON DELETE RESTRICT`
- `title` และ `description` เป็น `NOT NULL` และฐานข้อมูลต้องตรวจค่าที่ตัด Unicode whitespace แล้ว ความยาว และอักขระต้องห้ามตาม AC-3
- `status` เป็น `NOT NULL` ค่าเริ่มต้น "กำลังดำเนินการ" และรับเฉพาะหกค่าที่ระบุใน AC-8
- `process` เป็น `NOT NULL` ค่าเริ่มต้น "สัญญาโครงการ" และรับเฉพาะแปดค่าที่ระบุใน AC-9
- constraint/index/trigger ต้องบังคับกฎ title ของงานต้นฉบับและงานต่อเนื่องให้ถูกต้องภายใต้ concurrent writes
- trigger ต้องห้ามแก้ `id` และ `continuation_of_id` หลังสร้าง
- trigger ต้องบังคับ status transition, process transition, terminal lock และการคง process เดิมเมื่อเข้าสู่ terminal status ตาม AC-8 และ AC-9
- ทุก connection ต้องเปิด foreign keys และทุก mutation ต้องทำใน transaction แบบ atomic
