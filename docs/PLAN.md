# PLAN: Tracking Research Status MVP

## ขอบเขตและข้อกำหนดก่อนเริ่ม

- พัฒนาตาม SPEC: เพิ่ม/แสดง/ค้นหา/กรอง/แก้ไข/ลบงานวิจัย พร้อมบุคลากร แหล่งทุน PDF และหน้าปรับสถานะ/กระบวนการ รวมการนำทางและ UI states
- ไม่ใช้ฟิลด์รายละเอียดแบบเก่า; ใช้ข้อมูลโครงการใหม่ตาม AC-01–AC-14 และ schema ใน OpenAPI
- ใช้ SPEC กำหนดพฤติกรรม และ OpenAPI กำหนด HTTP contract; ปิดข้อขัดแย้งก่อน implementation ของส่วนที่เกี่ยวข้อง ไม่เพิ่ม endpoint เอง
- เรื่องที่ต้องให้ human ตัดสินใจ: สิทธิ์สามบทบาทกับ authentication ที่อยู่นอก scope; AC-02 ห้ามแสดง ID แต่ส่วน Feature/AC-10 ให้แสดงรหัส; จำนวน/สัดส่วนบุคลากร; validation ข้อความและวันที่; ขนาด multipart; การเก็บ/เปลี่ยน/ลบ PDF และการจัดการคำขอซ้ำ/แก้ไขพร้อมกัน
- แก้การอ้างเลข AC ใน SQL ให้ตรงหัวข้อปัจจุบันก่อนเขียน tests: validation คือ AC-04, สถานะคือ AC-11, กระบวนการคือ AC-12
- ไม่รวม deployment, notification และ Excel export ตาม SPEC; ยังไม่ออกแบบ authentication เพิ่มเพื่อแก้ข้อขัดแย้งเอง

## Backend

- ลำดับการทำงาน: Gin handler → service → repo → SQLite ผ่าน modernc.org/sqlite
- Handler รับผิดชอบ HTTP เท่านั้น: path, media type, body limits, multipart/JSON decoding, ส่ง typed input ให้ service และแปลงผลเป็น response; ไม่มี SQL หรือกฎ transition
- Service รับผิดชอบ validation และกฎธุรกิจ: ข้อมูลโครงการ/บุคลากร, ID และงานต้นทางที่แก้ไม่ได้, ชื่อซ้ำ, ข้อจำกัดการลบ, transition และ terminal lock
- Repo รับผิดชอบ SQL, transaction และการแปลงข้อผิดพลาดจาก constraint เป็น typed persistence errors; ไม่รู้จัก Gin, HTTP หรือไฟล์ PDF
- SQLite บังคับ identity/non-reused ID, foreign keys, ชื่อซ้ำตามประเภทงาน, enum และ transition invariants ผ่าน constraints/indexes/triggers; เปิด foreign keys ทุก connection และตรวจค่าล่าสุดพร้อม mutation แบบ atomic
- Schema เป้าหมายมี researches, research_members แบบหนึ่งต่อหลาย และ research_contracts แบบหนึ่งต่อหนึ่งตาม SPEC; จัดการ migration แยกจากการเปิด server และไม่ล้างข้อมูลอัตโนมัติ

## PDF และ API

- POST/PUT ใช้ multipart/form-data: ข้อมูลโครงการ, projectMembers เป็น JSON array ใน part เดียว และ PDF เป็น binary; PATCH สถานะ/กระบวนการใช้ JSON
- คงเส้นทาง health, list, create, update, delete, status และ process ตาม docs/openapi.yaml; responses ใช้ JSON และส่งเฉพาะ metadata ของ PDF ไม่ส่ง path ภายใน
- แยกตัวจัดเก็บไฟล์ผ่าน interface ให้ service ประสานงาน: ตรวจเนื้อหา/ขนาด PDF, เก็บชั่วคราว, บันทึกข้อมูลผ่าน repo และเผยแพร่ไฟล์ก่อนรายงานสำเร็จ
- SQLite transaction ไม่ครอบคลุม filesystem; ต้องกำหนดขั้นตอนชดเชยและกู้คืนเมื่อบันทึก/เผยแพร่ไฟล์ล้มเหลวหรือ process หยุด เพื่อไม่ให้ข้อมูลสำเร็จอ้างไฟล์ที่ไม่มีอยู่ และรักษาข้อมูล/ไฟล์เดิมเมื่อแก้ไขล้มเหลว
- ก่อน submit เก็บไฟล์ที่เลือกไว้ใน browser; การยกเลิกก่อนบันทึกไม่สร้างข้อมูลหรือไฟล์บน server ตาม AC-06

## Frontend

- React pages → components → generated typed API client จาก docs/openapi.yaml; pages/components ห้ามเรียก fetch ตรง และห้ามแก้ generated files ด้วยมือ
- Pages จัดการ navigation, loading และผลการบันทึก; components รับผิดชอบตาราง ฟอร์มบุคลากร/โครงการ ตัวเลือก PDF ตัวกรอง และ confirmation
- ใช้ ID เป็นตัวอ้างอิงภายในทุก action แม้ชื่อซ้ำ; การแสดง ID รอข้อสรุปความขัดแย้งใน SPEC
- ค้นหา/กรองบน frontend จากรายการที่โหลดมา; ยุบตัวกรองโดยคงค่า ล้างค่าได้ และแยกไม่มีข้อมูลออกจากไม่พบผลลัพธ์
- หลังสร้างสำเร็จกลับหน้ารายการและแสดงงานใหม่; หลังแก้ไข/ลบ/ปรับสถานะ/กระบวนการให้แสดงข้อมูลล่าสุด พร้อมป้องกันการส่งซ้ำและยืนยันก่อนทิ้งข้อมูล
- ก่อนทำ UI อ่าน .agents/skills/frontend-design/SKILL.md; ตรวจ loading/error/empty/success, responsive, keyboard, focus, contrast และ reduced motion โดยไม่ขยายขอบเขตฟีเจอร์

## Error handling

- Repo ส่ง typed persistence errors → service แปลงเป็น typed business errors → handler map เป็น HTTP status และ error code ตาม SPEC/OpenAPI
- ใช้ error envelope เดียวทั้งระบบ ไม่เปิดเผย SQL, stack trace หรือ path ภายใน; รูปแบบ error รายฟิลด์ต้องตกลงใน contract ก่อนทำ UI validation errors
- แยก validation, not-found, conflict, payload/media errors และ internal failure; กรณี terminal process lock ต้องตรวจตามลำดับที่ SPEC กำหนดก่อน idempotency

## Testing และการตรวจรับ

- ทำทีละ vertical slice และเขียน tests ครอบ AC ของ slice ก่อน implementation; ไม่ลบหรือแก้ tests เพียงเพื่อให้ผ่าน
- ใช้ go test และ net/http/httptest; ทุก test ที่ใช้ SQLite สร้างฐานข้อมูลของตนเองใน t.TempDir() หนึ่ง database ต่อหนึ่ง test ไม่ใช้ library.db จริงร่วมกัน
- Repo tests ตรวจ constraints, transactions, rollback และ concurrent writes; service tests ตรวจ validation และ transition matrix; handler/integration tests ตรวจ request/response และ status/error ทุกกรณีใน contract
- Tests ที่เกี่ยวกับ PDF ใช้พื้นที่ชั่วคราวแยกต่อ test ตรวจกรณีจัดเก็บล้มเหลว การคงไฟล์เดิม และการกำจัดไฟล์ชั่วคราวตามนโยบายที่ตกลง
- UI ใช้ E2E checklist ครอบ loading/error/empty/success, submit/cancel, refresh list, filters และ status/process actions; ไม่ถือว่า build ผ่านแทนการตรวจพฤติกรรม
- Quality gates: OpenAPI lint; backend go test ./..., golangci-lint, gosec, govulncheck; frontend generate:api พร้อมตรวจ drift, lint และ build; รันใน CI และ review ความครบ AC ก่อนส่งมอบ

## ลำดับดำเนินงาน

1. ปิดข้อกำหนดที่ขัดกันและปรับ OpenAPI ให้ตรง SPEC ก่อนเขียน implementation
2. Walking skeleton: health และ global error routing ผ่าน handler/service/repo/SQLite แล้วตั้งค่า generated client
3. เพิ่มงานวิจัยพร้อมข้อมูลใหม่และ PDF แล้วแสดงรายการให้ตรวจข้อมูลที่บันทึกได้
4. สร้าง UI เพิ่ม/รายการ/การนำทาง แล้วทำค้นหาและกรองเป็น slice แยก
5. แก้ไขและลบงานวิจัยทีละ slice พร้อมกฎไฟล์และ UI
6. ปรับสถานะและปรับกระบวนการทีละ slice รวม backend และหน้าจอของแต่ละฟีเจอร์
7. Hardening: regression, quality gates บนเครื่อง, CI และ review; แตกงานลง TASKS แยกต่างหากก่อนเริ่มทำจริง
