# SPEC Tracking-Research-Status MVP
## Users
- ผู้ดูแลระบบ: เพิ่มงานวิจัย, ดูรายการงานวิจัยทั้งหมด, ลบงานวิจัย, แก้ไขข้อมูล
- ผู้ประสานงาน: เพิ่มงานวิจัย, ดูรายการงานวิจัยทั้งหมด, ลบงานวิจัย, แก้ไขข้อมูล
- นักวิจัย: ดูรายการงานวิจัยทั้งหมด
- ทั้งสามบทบาทเห็นรายการวิจัยเดียวกัน ยังไม่มีมุมมองแยกตามบทบาท

## Out of scope (MVP v1)
- Authentication / user accounts
- การปรับสถานะของงานวิจัย
- มุมมองแยกตามบทบาท 
- Deployment

## Acceptance Criteria

AC-1: ระบบสามารถเพิ่มงานวิจัยได้
- เมื่อเรียก POST /api/v1/researches ด้วย title และ description ที่ถูกต้อง (บันทึกงานวิจัยสำเร็จ)
  → ตอบ 201
- Request body ต้องเป็น JSON object ที่มีเฉพาะ title และ description
- Response มีเฉพาะ title และ description
- Response ต้องตรงกับข้อมูลที่บันทึกในฐานข้อมูล
- ระบบตัด Unicode whitespace ที่หัวและท้ายก่อนตรวจสอบและบันทึก
- หาก request มี id หรือฟิลด์อื่น
  → ตอบ 422 พร้อม code VALIDATION_ERROR
- หากมีงานวิจัยที่ใช้ title เดียวกันหลังตัด whitespace
  → ตอบ 409 พร้อม code TITLE_ALREADY_EXISTS
- การตรวจ title ซ้ำเป็นแบบ case-sensitive
- ฐานข้อมูลต้องมี UNIQUE constraint สำหรับ title เพื่อป้องกันข้อมูลซ้ำจาก concurrent requests

AC-2: ระบบตรวจสอบ title
- ต้องมีฟิลด์ title
- ค่าต้องเป็น string และห้ามเป็น null
- ระบบตัด Unicode whitespace ที่หัวและท้ายก่อน validation
- หลังตัด whitespace ต้องมีความยาว 1–200 Unicode characters
- ห้ามมี newline, tab, NUL, control characters และ /
- หากไม่ผ่านเงื่อนไข
  → ตอบ 422 พร้อม code VALIDATION_ERROR

AC-3: ระบบตรวจสอบ description
- ต้องมีฟิลด์ description
- ค่าต้องเป็น string และห้ามเป็น null
- ระบบตัด Unicode whitespace ที่หัวและท้ายก่อน validation
- หลังตัด whitespace ต้องมีความยาว 1–5,000 Unicode characters
- อนุญาต newline และ tab
- ห้ามมี NUL และ control characters อื่น
- หากไม่ผ่านเงื่อนไข
  → ตอบ 422 พร้อม code VALIDATION_ERROR

AC-4: ระบบตรวจสอบ request body
- POST และ PUT ต้องใช้ Content-Type: application/json
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
- JSON มี key ซ้ำหรือฟิลด์ที่ระบบไม่รองรับ
  → ตอบ 422 พร้อม code VALIDATION_ERROR
- Request body มีขนาดเกิน 64 KiB
  → ตอบ 413 พร้อม code PAYLOAD_TOO_LARGE

AC-5: ระบบแสดงรายการงานวิจัยทั้งหมดได้
- เมื่อเรียก GET /api/v1/researches
  → ตอบ 200
- Response มี Content-Type: application/json
- Response เป็น JSON array
- แต่ละรายการมีเฉพาะ title และ description
- ต้องไม่มี id หรือฟิลด์อื่น
- ระบบเรียงรายการตาม title จากน้อยไปมาก
- หากไม่มีงานวิจัย
  → ตอบ 200 พร้อม []
- ผู้เรียกทุกคนเห็นรายการชุดเดียวกัน
- หากส่ง request body ที่ไม่ว่าง
  → ตอบ 400 พร้อม code INVALID_REQUEST_BODY
- หากส่ง query parameter ใด
  → ตอบ 422 พร้อม code VALIDATION_ERROR

AC-6: ระบบสามารถแก้ไขงานวิจัยได้
- {title} คือ title ปัจจุบันและต้องผ่าน URL encoding
- Request body ต้องมีทั้ง title และ description
- การแก้ไขเป็นการแทนข้อมูลเดิมทั้งหมด
- เมื่อพบงานวิจัยและข้อมูลใหม่ถูกต้อง
  → ตอบ 200 พร้อม title และ description หลังแก้ไข
- สามารถเปลี่ยน title ได้
- หลังเปลี่ยน title แล้ว title เดิมต้องไม่สามารถใช้อ้างอิงรายการได้
- หากไม่พบ title ปัจจุบัน
  → ตอบ 404 พร้อม code RESEARCH_NOT_FOUND
- หาก title ใหม่เหมือน title เดิม สามารถแก้ไข description ได้
- หาก path title มีรูปแบบไม่ถูกต้อง
  → ตอบ 422 พร้อม code VALIDATION_ERROR
- การแก้ไข title และ description ต้องสำเร็จหรือล้มเหลวพร้อมกันใน transaction เดียว

AC-7: ระบบสามารถลบงานวิจัยได้
- {title} คือ title ปัจจุบันและต้องผ่าน URL encoding
- เมื่อพบและลบสำเร็จ
  → ตอบ 204 โดยไม่มี response body
- หลังลบ รายการต้องไม่ปรากฏใน GET /api/v1/researches
- หากไม่พบ title
  → ตอบ 404 พร้อม code RESEARCH_NOT_FOUND
- หากเรียกลบ title เดิมซ้ำ
  → ตอบ 404 พร้อม code RESEARCH_NOT_FOUND
- หากส่ง request body ที่ไม่ว่าง
  → ตอบ 400 พร้อม code INVALID_REQUEST_BODY
- หากส่ง query parameter ใด
  → ตอบ 422 พร้อม code VALIDATION_ERROR

## Error format (ทั้งระบบ)
{"error": {"code": "MACHINE_READABLE_CODE", "message": "human readable"}}

## API Contract
GET  /health
     → 200 {"status":"ok"}
     → 503 SERVICE_UNAVAILABLE
     → 500 INTERNAL_ERROR

GET  /api/v1/researches
     → 200 [{title,description}]
     → 400 INVALID_REQUEST_BODY
     → 422 VALIDATION_ERROR
     → 500 INTERNAL_ERROR

POST /api/v1/researches {title,description}
     → 201 {title,description}
     → 400 INVALID_JSON
     → 409 TITLE_ALREADY_EXISTS
     → 413 PAYLOAD_TOO_LARGE
     → 415 UNSUPPORTED_MEDIA_TYPE
     → 422 VALIDATION_ERROR
     → 500 INTERNAL_ERROR

PUT  /api/v1/researches/{title} {title,description}
     → 200 {title,description}
     → 400 INVALID_JSON
     → 404 RESEARCH_NOT_FOUND
     → 413 PAYLOAD_TOO_LARGE
     → 415 UNSUPPORTED_MEDIA_TYPE
     → 422 VALIDATION_ERROR
     → 500 INTERNAL_ERROR

DELETE /api/v1/researches/{title}
       → 204
       → 400 INVALID_REQUEST_BODY
       → 404 RESEARCH_NOT_FOUND
       → 422 VALIDATION_ERROR
       → 500 INTERNAL_ERROR

ทุก Endpoint
     → 404 ROUTE_NOT_FOUND
     → 405 METHOD_NOT_ALLOWED

## SQL
     CREATE TABLE IF NOT EXISTS researches (
  title TEXT NOT NULL COLLATE BINARY,
  description TEXT NOT NULL,

  -- ใช้ title เป็นตัวระบุรายการและห้ามซ้ำ
  CONSTRAINT researches_pk
    PRIMARY KEY (title),

  -- ต้องบันทึกค่าที่ตัดช่องว่างหัวท้ายแล้ว
  CONSTRAINT title_must_be_trimmed
    CHECK (title = trim(title)),

  CONSTRAINT description_must_be_trimmed
    CHECK (description = trim(description)),

  -- ตรวจความยาวตาม SPEC
  CONSTRAINT title_length
    CHECK (length(title) BETWEEN 1 AND 200),

  CONSTRAINT description_length
    CHECK (length(description) BETWEEN 1 AND 5000),

  -- title ห้ามมี /, newline, tab และ NUL
  CONSTRAINT title_forbidden_characters
    CHECK (
      instr(title, '/') = 0
      AND instr(title, char(0)) = 0
      AND instr(title, char(9)) = 0
      AND instr(title, char(10)) = 0
      AND instr(title, char(13)) = 0
    ),

  -- description อนุญาต tab/newline แต่ห้าม NUL
  CONSTRAINT description_forbidden_characters
    CHECK (
      instr(description, char(0)) = 0
    )
) WITHOUT ROWID;
