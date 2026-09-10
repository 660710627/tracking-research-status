# TASKS: Tracking Research Status MVP

## สถานะและกติกา

รายการนี้เป็นแผนงาน ยังไม่มี task ใดถูกประกาศว่าผ่าน DoD ในรอบเริ่มใหม่ ทำตามลำดับ dependencies ทีละ task; ไม่ใช่การอนุญาตให้เริ่ม implementation ขณะที่ SPEC ยังขัดกัน

- T-01 เป็นงานแรกของรอบนี้และต้องผ่านก่อนแก้โค้ดทุกชนิด รวม tests, generated client และ test tooling
- ต้องให้ human ปิดข้อกำหนดใน PLAN ก่อน T-01 เสร็จ: สิทธิ์ผู้ใช้/authentication, แสดงหรือซ่อน ID, validation/จำนวนบุคลากร/สัดส่วน, วันและข้อความ, PDF/ขนาด multipart/ไฟล์เดิม, retries/concurrent edits, search matching และเกณฑ์ UI; แก้เลข AC ที่ SQL อ้างผิดด้วย ห้ามเดาคำตอบ
- ถ้าข้อสรุปเพิ่มฟีเจอร์หรือ endpoint นอก SPEC ปัจจุบัน ให้ปรับแผนนี้ก่อนเริ่มโค้ด ไม่แทรกฟีเจอร์นั้นใน task เดิม
- ทุก slice ต้องมีรายการ mapping ทุก bullet ของ AC ไปยัง backend tests หรือ E2E test cases ตามพฤติกรรม ไม่ใช่ระบุเพียงหมายเลข AC รวม; ใช้ mapping นี้เป็นเงื่อนไข DoD เพิ่มเติมของทุก task ทดสอบและ task ปิด slice
- AC ที่เป็นกฎร่วมต้องถูกทดสอบซ้ำในทุก operation ที่ได้รับผล เช่น validation ใน create/update, immutable ID ในทุก mutation, errors/concurrency/PDF failure ใน operation ที่เกี่ยวข้อง; ห้ามย้ายไป hardening
- Task tests อนุญาตเฉพาะ tests, fixture, test doubles และเอกสาร test cases ไม่เขียน production implementation หรือ stub ให้ผ่าน; หลังรันต้องมีผลแดงที่อธิบายสาเหตุได้ ไม่ใช่ dependency/network/tooling failure
- ก่อนเขียน tests ของ API ภายในแต่ละ slice ให้ระบุชื่อ typed inputs/interfaces ที่คาดหวังในบันทึกของ task; compile error เพราะ production symbol เป้าหมายยังไม่มีนับเป็น RED ได้ แต่ต้องไม่ใช่ test syntax ผิด และ task implementation ถัดไปต้องรองรับ contract นั้น
- Task implement ห้ามแก้ *_test.go; ไม่ลด assertions/skip tests เพื่อผ่าน ถ้า test กับ SPEC ขัดกันให้หยุดส่วนที่เกี่ยวข้องและรายงาน
- ทุก test ที่แตะ persistence ใช้ SQLite ใหม่ใน t.TempDir() หนึ่ง database ต่อหนึ่ง test; PDF และ fixture ใช้พื้นที่ชั่วคราว ไม่แตะ library.db หรือไฟล์อัปโหลดจริง
- ระหว่าง implementation ราย layer อาจยังแดงใน layer ที่ยังไม่ได้ทำ ต้องรันคำสั่งเฉพาะ layer ให้ผ่านและรายงานผล go test ./... ที่เหลือแดงตาม DoD; ก่อนปิด backend slice ต้องผ่านทั้งชุด
- UI test tasks ใช้ E2E checklist ที่ทดลองได้และบันทึก actual/expected พร้อมหลักฐาน RED ก่อนสร้าง UI; ไม่บังคับเพิ่ม framework ทดสอบใหม่ ถ้าใช้ API fixtures ต้องตรง contract และทวนกับ backend จริงก่อนปิด slice
- UI ทุก task ต้องอ่าน .agents/skills/frontend-design/SKILL.md และใช้ generated typed client; UX ไม่มีสิทธิ์เปลี่ยน AC/API/scope
- คำสั่ง backend/frontend รันใน directory ที่ระบุ; รายการคำสั่งหลัง cd ภายใน DoD เดียวกันใช้ directory นั้น เว้นแต่ระบุจาก root
- ขั้นตอนตรวจรับที่ต้องสร้างข้อมูลให้ใช้ฐานข้อมูล/ไฟล์ทดลองแยก และขอสิทธิ์ระบบตามความจำเป็น; ไม่ push/merge main โดยอัตโนมัติ

## ขอบเขต endpoint (ห้ามเพิ่มหรือเปลี่ยนชื่อ)

| Method | Endpoint |
|---|---|
| GET | /health |
| GET | /api/v1/researches |
| POST | /api/v1/researches |
| PUT | /api/v1/researches/{id} |
| DELETE | /api/v1/researches/{id} |
| PATCH | /api/v1/researches/{id}/status |
| PATCH | /api/v1/researches/{id}/process |

## การปิด coverage ของแต่ละ slice

| Slice | AC และเงื่อนไขที่ต้องครบ | Tests → implementation |
|---|---|---|
| Health | AC-13 health/errors/global routing | T-02 → T-03 |
| เพิ่มงานวิจัยพร้อมบุคลากร/PDF | AC-01–06, AC-13 และ AC-14 ส่วนฟอร์ม รวม field-level errors, submit/cancel/back, ID UI ตามข้อสรุป | T-05–07 → T-08–10 และ T-15 |
| รายการ | AC-07, AC-02 ส่วนใช้ ID, AC-13 และ UI states AC-14 | T-11/T-13 → T-12/T-14 |
| การนำทาง | AC-14 sidebar/list/create; เมนู status/process ตรวจใน slice ของหน้านั้น | T-16 → T-17 |
| ค้นหา/กรอง | AC-08 และ states/accessibility AC-14 | T-18 → T-19 |
| แก้ไข | AC-09 และกฎร่วม AC-02–06/13/14 รวม feedback AC-07 | T-20–22 → T-23–26 |
| ลบ | AC-10 และกฎร่วม AC-02/03/06/13/14 รวม empty/refresh AC-07 | T-27–28 → T-29–30 |
| สถานะ | AC-11, identity AC-02, AC-13/14 | T-31–32 → T-33–35 |
| กระบวนการ | AC-12, identity AC-02, terminal AC-11, AC-13/14 | T-36–37 → T-38–40 |

การเพิ่มงานต้องกลับมาเห็นรายการ จึงมี list เป็น dependency ที่ปิดก่อน T-15; ห้ามนับ create slice เสร็จเพียง backend ผ่าน T-10 และห้ามย้าย success/cancel/ไฟล์ค้างไป hardening แต่ละ task ของ list ยังทำเฉพาะ list ไม่รวม create code

## รายการ task ตามลำดับ

### T-01 — ปรับ docs/openapi.yaml ให้ตรง SPEC ก่อนโค้ด

- **สิ่งที่ทำ:** ปิดข้อขัดแย้งในรายการก่อนเริ่มกับ human แล้วปรับ OpenAPI ของ 7 operations เดิม รวม schema ข้อมูลใหม่, multipart, PDF, responses และ error codes; บันทึกข้อสรุปใน SPEC/PLAN เฉพาะที่จำเป็น ไม่เปลี่ยน endpoint หรือสร้าง authentication เอง
- **Dependencies:** ไม่มี
- **DoD:** `npx @redocly/cli lint docs/openapi.yaml` และ `git diff --check -- docs/openapi.yaml` ผ่าน; ทุกประเด็นก่อนเริ่มมีข้อสรุปที่วัดผลได้; SPEC/SQL/OpenAPI ไม่ขัดกันและไม่มีฟิลด์ข้อมูล description; ไม่มี backend/frontend code ใหม่

### T-02 — Tests: health และ global routing

- **สิ่งที่ทำ:** เขียน httptest/integration tests สำหรับ health 200/503/500, JSON/error envelope, unknown route 404, wrong method 405, internal-detail suppression และ database isolation ตาม AC-13
- **Dependencies:** T-01
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests

### T-03 — Implement: health walking skeleton

- **สิ่งที่ทำ:** สร้าง Gin handler → service → repo → SQLite สำหรับ health และ global routing เท่านั้น พร้อม composition root และเปิด foreign keys ทุก connection; ไม่มี CRUD/status/process stub endpoints
- **Dependencies:** T-02
- **DoD:** `cd backend; go test ./...` ผ่าน; `cd backend; go run ./cmd/server` เปิดได้; `curl.exe -i http://127.0.0.1:8080/health` ได้ 200 ตาม contract บนฐานข้อมูลทดลอง

### T-04 — ตั้งค่า generated typed API client

- **สิ่งที่ทำ:** ตั้งค่า generator จาก OpenAPI และ wrapper แยกจาก generated code ให้รองรับ multipart กับ JSON ตาม schema; ยังไม่สร้างหน้าฟีเจอร์
- **Dependencies:** T-03
- **DoD:** `cd frontend; npm run generate:api`, `npm run lint`, `npm run build` ผ่าน; generate สองครั้งได้เนื้อหาเดียวกัน; ไม่มี endpoint ที่แต่งเองหรือ direct fetch ใน pages/components

### T-05 — Tests: create persistence และ invariants

- **สิ่งที่ทำ:** เขียน repo tests AC-01/02/03/04/05/06/13 ฝั่ง persistence: ทุกฟิลด์/สมาชิก/สัญญา, ID บวกไม่ซ้ำไม่ใช้ซ้ำ/แก้ไม่ได้, foreign key, ชื่อซ้ำหลัง trim/case, ต้นทางที่จบแล้ว, child หลายงาน, constraints, rollback, concurrent create; ใช้ SQL เฉพาะใน test fixture เพื่อพิสูจน์ non-reused ID โดยไม่ต้อง implement DELETE
- **Dependencies:** T-04
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests

### T-06 — Tests: create validation และ PDF orchestration

- **สิ่งที่ทำ:** เขียน service/storage tests ครบข้อมูลจำเป็นและเงื่อนไขที่ T-01 ปิดแล้ว: Unicode/1000 boundary, whitespace/control/title, enum/boolean/null, วันจริง พ.ศ., งบประมาณ precision, สมาชิก/email/roles/สัดส่วน, PDF content/size, stage/publish/cleanup/rollback/recovery ตามนโยบาย; ไม่รายงานสำเร็จก่อนข้อมูลและไฟล์พร้อม; ไม่เปิดเผย path
- **Dependencies:** T-05
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests

### T-07 — Tests: create HTTP และ E2E เพิ่มงานวิจัย

- **สิ่งที่ทำ:** เขียน handler/integration tests POST multipart ครบ 201/404/409/413/415/422/500, boundary/part ซ้ำหรือขาด/field เกิน/JSON สมาชิก/ชนิด scalar, ห้าม client ส่ง immutable/server fields และ response metadata; เขียน E2E เพิ่มครบทุกข้อ AC-01/02/03/04/05/06/14 รวมกลับรายการและยืนยันก่อนทิ้งข้อมูล
- **Dependencies:** T-06
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests; เขียน E2E checklist เป็น test cases พร้อม expected result และทดลองบนหน้าเริ่มต้น/fixture ที่ควบคุมได้; บันทึกผลแดงเพราะพฤติกรรมยังไม่มี (ไม่ใช่ server เปิดไม่ได้); ไม่เขียนฟีเจอร์; checklist: loading/submitting ป้องกันส่งซ้ำ; error แจ้งฟิลด์และคงค่าที่กรอก; empty ฟอร์มตรวจ required/ไม่มีต้นทาง; success กลับรายการเห็นงานใหม่; cancel/back ไม่มี mutation หรือไฟล์ค้าง

### T-08 — Implement: create schema และ repo

- **สิ่งที่ทำ:** สร้าง schema เป้าหมายและ create transaction/typed persistence errors ตาม tests; บันทึกโครงการ สมาชิกและสัญญาหนึ่งต่อหนึ่ง พร้อม constraints/identity/continuation; ไม่เพิ่ม list/update/delete SQL
- **Dependencies:** T-07
- **DoD:** `cd backend; go test ./internal/repo ./internal/db` ผ่าน tests ที่มี; `cd backend; go test ./...` รันและเหลือแดงเฉพาะ service/HTTP create ที่รอ T-09/T-10; ไม่แก้ *_test.go

### T-09 — Implement: create service และ storage adapter

- **สิ่งที่ทำ:** สร้าง typed inputs/errors และ business validation พร้อม file-storage interface/adapter; ประสาน transaction กับ stage/publish/compensation ตาม tests T-06; ไม่มี HTTP parsing ใน service หรือ filesystem ใน repo
- **Dependencies:** T-08
- **DoD:** `cd backend; go test ./internal/service ./internal/repo ./internal/db` ผ่าน; `cd backend; go test ./...` เหลือแดงเฉพาะ HTTP create ที่รอ T-10; ไม่แก้ *_test.go

### T-10 — Implement: POST research handler

- **สิ่งที่ทำ:** เชื่อม POST /api/v1/researches เข้ากับ service; decode multipart/limits/typed errors ตาม T-07 และ update composition root; ไม่เปิด endpoint อัปโหลดแยก
- **Dependencies:** T-09
- **DoD:** `cd backend; go test ./...` ผ่าน; POST integration ครบ contract และ AC create ฝั่ง backend; ไม่แก้ *_test.go

### T-11 — Tests: list researches backend

- **สิ่งที่ทำ:** เขียน repo/service/handler tests GET /api/v1/researches ตาม AC-07/13: ทุกฟิลด์ใหม่และ metadata/สมาชิกครบ, ไม่มี binary/path/description, sort title/id และชื่อซ้ำ, 200 array/empty, 400 body, 422 query, 500 database failure; ใช้ create ที่ผ่านแล้วเตรียมข้อมูลใน test DB
- **Dependencies:** T-10
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests

### T-12 — Implement: list researches backend

- **สิ่งที่ทำ:** สร้าง list SQL/service/GET handler ตาม T-11; ไม่มี query filtering หรือ endpoint GET รายตัว
- **Dependencies:** T-11
- **DoD:** `cd backend; go test ./...` ผ่าน; response และ sort ตรง contract; ไม่แก้ *_test.go

### T-13 — Tests: UI รายการงานวิจัย

- **สิ่งที่ทำ:** เขียนและรัน E2E checklist ตาม AC-07 และส่วน UI ของ AC-02/14; ระบุคอลัมน์/ID ตามข้อสรุป T-01, จำนวนรายการและการระบุรายการที่เพิ่งเปลี่ยน ใช้ fixture สำหรับผลหลัง mutation
- **Dependencies:** T-12
- **DoD:** เขียน E2E checklist เป็น test cases พร้อม expected result และทดลองบนหน้าเริ่มต้น/fixture ที่ควบคุมได้; บันทึกผลแดงเพราะพฤติกรรมยังไม่มี (ไม่ใช่ server เปิดไม่ได้); ไม่เขียนฟีเจอร์; loading แสดงระหว่างรอ API; error แสดงและ retry ได้; empty [] ไม่ปะปน network error; success ข้อมูลครบ sort/จำนวน/ชื่อซ้ำใช้ ID ภายในถูกต้อง; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-14 — Implement: UI รายการงานวิจัย

- **สิ่งที่ทำ:** สร้าง page/components อ่านผ่าน generated client ตาม T-13; รองรับ feedback/refresh หลัง mutation โดยยังไม่ implement mutation UI อื่น
- **Dependencies:** T-13
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E ที่เขียนไว้ผ่านทุกข้อ พร้อมหลักฐาน; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-15 — Implement: UI เพิ่มงานวิจัย

- **สิ่งที่ทำ:** สร้างฟอร์มโครงการ สมาชิกและ PDF ตาม E2E T-07; ส่ง multipart ผ่าน generated client กลับรายการหลังสำเร็จ รักษาค่าเมื่อ error และยกเลิก/back พร้อมยืนยันก่อนทิ้ง; รองรับ ID ตาม T-01
- **Dependencies:** T-14
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E ที่เขียนไว้ผ่านทุกข้อ พร้อมหลักฐาน; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components; รันทวน E2E create ทั้งหมดใน T-07 กับ backend จริงบนข้อมูลทดลอง; AC-01 ถึง AC-06 ที่เกี่ยวกับ create และ UI AC-14 ครบก่อนปิด create slice

### T-16 — Tests: การนำทางและ sidebar

- **สิ่งที่ทำ:** เขียน E2E navigation AC-14/Feature 7 สำหรับเมนูรายการ/เพิ่มงานและขยายยุบ; ไม่สร้างเมนูที่ยังไม่มีหน้าโดยแสร้งว่าทำงานแล้ว; เมนู status/process จะทดสอบพร้อม slice ของหน้านั้น
- **Dependencies:** T-15
- **DoD:** เขียน E2E checklist เป็น test cases พร้อม expected result และทดลองบนหน้าเริ่มต้น/fixture ที่ควบคุมได้; บันทึกผลแดงเพราะพฤติกรรมยังไม่มี (ไม่ใช่ server เปิดไม่ได้); ไม่เขียนฟีเจอร์; loading เปลี่ยนหน้าแล้วแสดงสถานะรอ; error หน้าปลายทางเสียยังใช้เมนูได้; empty หน้ารายการว่างยังเข้าเพิ่มได้; success ขยาย/ยุบและไปหน้าถูกต้อง; dirty form ยืนยันก่อนทิ้ง; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-17 — Implement: การนำทางและ sidebar

- **สิ่งที่ทำ:** สร้าง navigation/shared shell ตาม T-16; เชื่อมเฉพาะหน้า list/create ที่เสร็จแล้ว ไม่เปลี่ยน API
- **Dependencies:** T-16
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E ที่เขียนไว้ผ่านทุกข้อ พร้อมหลักฐาน; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-18 — Tests: ค้นหาและกรอง

- **สิ่งที่ทำ:** เขียน E2E ครบ AC-08 และ UI AC-14: matching ตาม T-01, รหัส/ชื่อ, status/process, AND หลายเงื่อนไข, จำนวนพบ/ทั้งหมด, collapse คงค่า, clear คืนค่า, ไม่ค้น description
- **Dependencies:** T-17
- **DoD:** เขียน E2E checklist เป็น test cases พร้อม expected result และทดลองบนหน้าเริ่มต้น/fixture ที่ควบคุมได้; บันทึกผลแดงเพราะพฤติกรรมยังไม่มี (ไม่ใช่ server เปิดไม่ได้); ไม่เขียนฟีเจอร์; loading ก่อนโหลดรายการ; error โหลดไม่ได้ไม่รายงานว่าไม่มีผล; empty แยก [] กับไม่พบตาม filter; success matching/count/clear ถูกต้อง; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-19 — Implement: ค้นหาและกรอง

- **สิ่งที่ทำ:** เพิ่มตัวกรองใน frontend บนรายการที่โหลดมาเท่านั้น ตาม T-18 ไม่ส่ง query parameter ไป GET และไม่เพิ่ม search API
- **Dependencies:** T-18
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E ที่เขียนไว้ผ่านทุกข้อ พร้อมหลักฐาน; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-20 — Tests: update persistence

- **สิ่งที่ทำ:** เขียน repo tests AC-09 และกฎร่วม AC-02/03/04/05/06/13: full replacement ข้อมูลที่แก้ได้/สมาชิก/metadata, immutable fields, ชื่อเดิมของ root ที่มี child ชื่อซ้ำ, เปลี่ยนชื่อ conflict, ไม่พบ ID, rollback และ concurrency ตามข้อสรุป T-01
- **Dependencies:** T-19
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests

### T-21 — Tests: update validation และ PDF

- **สิ่งที่ทำ:** เขียน service/storage tests ทุก validation ที่ใช้ตอน create ซ้ำสำหรับ update, การคงข้อมูล/ไฟล์เดิมเมื่อ failure, publish/cleanup/recovery และการจัดการ PDF เดิมตามนโยบาย T-01; ไม่ลด coverage ด้วยการอ้างว่า create เคยทดสอบแล้ว
- **Dependencies:** T-20
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests

### T-22 — Tests: update HTTP และ UI

- **สิ่งที่ทำ:** เขียน PUT handler/integration tests ครบ 200/404/409/413/415/422/500, multipart request/error/immutable rejection และ JSON Research response; เขียนและรัน E2E AC-09/14
- **Dependencies:** T-21
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests; เขียน E2E checklist เป็น test cases พร้อม expected result และทดลองบนหน้าเริ่มต้น/fixture ที่ควบคุมได้; บันทึกผลแดงเพราะพฤติกรรมยังไม่มี (ไม่ใช่ server เปิดไม่ได้); ไม่เขียนฟีเจอร์; loading โหลดค่าปัจจุบันและ submitting; error validation/conflict/not-found/API failure คงค่ากรอก; empty ตรวจ required; success ข้อมูลล่าสุดตาม ID; cancel ไม่เปลี่ยนข้อมูล; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-23 — Implement: update repo

- **สิ่งที่ทำ:** เพิ่ม atomic update SQL และ constraint error mapping ตาม T-20; ข้อมูลโครงการ สมาชิกและสัญญาต้องถูกแทนที่เป็นชุด
- **Dependencies:** T-22
- **DoD:** `cd backend; go test ./internal/repo ./internal/db` ผ่าน; `cd backend; go test ./...` เหลือแดงเฉพาะ update service/HTTP; ไม่แก้ *_test.go

### T-24 — Implement: update service และ PDF

- **สิ่งที่ทำ:** เพิ่ม validation/rules และ storage orchestration ตาม T-21; รักษา immutable fields และไฟล์เดิมเมื่อ failure
- **Dependencies:** T-23
- **DoD:** `cd backend; go test ./internal/service ./internal/repo ./internal/db` ผ่าน; `cd backend; go test ./...` เหลือแดงเฉพาะ update HTTP; ไม่แก้ *_test.go

### T-25 — Implement: PUT research handler

- **สิ่งที่ทำ:** เพิ่ม PUT /api/v1/researches/{id} และ map HTTP errors ตาม T-22 เท่านั้น
- **Dependencies:** T-24
- **DoD:** `cd backend; go test ./...` ผ่านทั้งหมด; ไม่แก้ *_test.go

### T-26 — Implement: UI แก้ไขงานวิจัย

- **สิ่งที่ทำ:** สร้าง edit form เติมค่าปัจจุบันผ่าน list data/generated client, ส่งเฉพาะฟิลด์ที่ contract อนุญาตและจัดการ PDF ตาม T-01; refresh/feedback หลังสำเร็จ
- **Dependencies:** T-25
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E ที่เขียนไว้ผ่านทุกข้อ พร้อมหลักฐาน; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components; E2E T-22 ผ่านกับ backend จริง; ปิด AC-09 และกฎร่วมทุกข้อของ update ใน slice นี้

### T-27 — Tests: ลบงานวิจัย backend

- **สิ่งที่ทำ:** เขียน repo/service/handler/storage tests AC-10/13 และ AC-02/03 ที่เกี่ยวข้อง: 204 body ว่าง, ไม่พบ/ลบซ้ำ 404, invalid ID 422, body 400/query 422, parent 409, metadata/สมาชิกหายพร้อมกัน, PDF ตามนโยบาย, ไม่ใช้ ID ซ้ำ, rollback, concurrent delete/continuation และ 500
- **Dependencies:** T-26
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests

### T-28 — Tests: UI ลบงานวิจัย

- **สิ่งที่ทำ:** เขียนและรัน E2E AC-10/14; แสดงข้อมูลยืนยันตามข้อสรุปเรื่อง ID ใน T-01
- **Dependencies:** T-27
- **DoD:** เขียน E2E checklist เป็น test cases พร้อม expected result และทดลองบนหน้าเริ่มต้น/fixture ที่ควบคุมได้; บันทึกผลแดงเพราะพฤติกรรมยังไม่มี (ไม่ใช่ server เปิดไม่ได้); ไม่เขียนฟีเจอร์; loading/deleting กันส่งซ้ำ; error 404/409/500 แจ้งและไม่ซ่อนงานผิด; empty หลังลบรายการสุดท้าย; success ลบตรง ID และ refresh; cancel ไม่ส่ง DELETE; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-29 — Implement: ลบงานวิจัย backend

- **สิ่งที่ทำ:** เพิ่ม delete repo/service/handler และการจัดการไฟล์ตาม T-27; ห้ามลบ parent ที่มี child และห้าม mutation บางส่วน
- **Dependencies:** T-28
- **DoD:** `cd backend; go test ./...` ผ่านทั้งหมด; ไม่แก้ *_test.go

### T-30 — Implement: UI ลบงานวิจัย

- **สิ่งที่ทำ:** เพิ่ม confirmation/action ผ่าน generated client และ refresh list ตาม T-28
- **Dependencies:** T-29
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E ที่เขียนไว้ผ่านทุกข้อ พร้อมหลักฐาน; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components; AC-10 และกฎร่วมของ delete ครบใน slice นี้

### T-31 — Tests: status transition backend

- **สิ่งที่ทำ:** เขียน repo/service/handler tests AC-11/13 และ identity: matrix 6 สถานะครบทุกคู่, next/skip/back, terminal jump/lock, same-state รวม terminal, process/ข้อมูลอื่นคงเดิม, latest-state atomic concurrency, 200/400/404/409/413/415/422/500 และ request rules ของ PATCH
- **Dependencies:** T-30
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests

### T-32 — Tests: UI ปรับสถานะ

- **สิ่งที่ทำ:** เขียน E2E AC-11/14 รวมเมนูเข้าหน้าสถานะ, ค้นหางาน, summary, allowed options, confirmation และ refresh
- **Dependencies:** T-31
- **DoD:** เขียน E2E checklist เป็น test cases พร้อม expected result และทดลองบนหน้าเริ่มต้น/fixture ที่ควบคุมได้; บันทึกผลแดงเพราะพฤติกรรมยังไม่มี (ไม่ใช่ server เปิดไม่ได้); ไม่เขียนฟีเจอร์; loading ผลค้นหา/submitting; error not-found/transition/terminal/API failure; empty ไม่พบงาน; success เปลี่ยนสถานะถูก ID ข้อมูลอื่นคงเดิม; cancel ไม่ส่ง PATCH; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-33 — Implement: status repo/service

- **สิ่งที่ทำ:** เพิ่ม atomic status SQL/trigger และ service rules/typed errors ตาม T-31; ไม่เปลี่ยน process หรือเพิ่ม process endpoint
- **Dependencies:** T-32
- **DoD:** `cd backend; go test ./internal/repo ./internal/service ./internal/db` ผ่าน; `cd backend; go test ./...` เหลือแดงเฉพาะ status HTTP; ไม่แก้ *_test.go

### T-34 — Implement: status PATCH handler

- **สิ่งที่ทำ:** เพิ่ม PATCH /api/v1/researches/{id}/status ตาม T-31 พร้อม request validation/error mapping
- **Dependencies:** T-33
- **DoD:** `cd backend; go test ./...` ผ่านทั้งหมด; ไม่แก้ *_test.go

### T-35 — Implement: UI ปรับสถานะ

- **สิ่งที่ทำ:** สร้างหน้าและเมนู status ตาม T-32 ผ่าน generated client; backend ยังตรวจ transition ซ้ำเสมอ
- **Dependencies:** T-34
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E ที่เขียนไว้ผ่านทุกข้อ พร้อมหลักฐาน; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components; AC-11 ทั้ง backend/UI ครบก่อนเริ่ม process slice

### T-36 — Tests: process transition backend

- **สิ่งที่ทำ:** เขียน repo/service/handler tests AC-12/13 และ identity: matrix 8 ขั้นครบทุกคู่, next/skip/back/last, same-state เมื่อยังไม่จบ, terminal lock รวม same-process, status/ข้อมูลอื่นคงเดิม/ไม่ auto-complete, atomic concurrency รวมแข่งกับ terminal status; 200/400/404/409/413/415/422/500 และ PATCH rules
- **Dependencies:** T-35
- **DoD:** รัน `cd backend; go test ./...` แล้วแดงจากความสามารถเป้าหมายที่ยังไม่มี พร้อมบันทึกคำสั่งและสาเหตุ; mapping AC ของขอบเขตนี้ครบ ไม่มี production code หรือ skip tests

### T-37 — Tests: UI ปรับกระบวนการ

- **สิ่งที่ทำ:** เขียน E2E AC-12/14 รวมเมนูเข้าหน้า, ค้นหา, current/next ของ 8 ขั้น, confirmation, refresh และ terminal lock
- **Dependencies:** T-36
- **DoD:** เขียน E2E checklist เป็น test cases พร้อม expected result และทดลองบนหน้าเริ่มต้น/fixture ที่ควบคุมได้; บันทึกผลแดงเพราะพฤติกรรมยังไม่มี (ไม่ใช่ server เปิดไม่ได้); ไม่เขียนฟีเจอร์; loading ค้นหา/submitting; error not-found/transition/terminal/API failure; empty ไม่พบงาน; success ไปขั้นถัดไปโดย status ไม่เปลี่ยน; cancel ไม่ส่ง PATCH; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components

### T-38 — Implement: process repo/service

- **สิ่งที่ทำ:** เพิ่ม atomic process SQL/trigger และ service rules/typed errors ตาม T-36; terminal lock มาก่อน same-state; ไม่ auto-complete status
- **Dependencies:** T-37
- **DoD:** `cd backend; go test ./internal/repo ./internal/service ./internal/db` ผ่าน; `cd backend; go test ./...` เหลือแดงเฉพาะ process HTTP; ไม่แก้ *_test.go

### T-39 — Implement: process PATCH handler

- **สิ่งที่ทำ:** เพิ่ม PATCH /api/v1/researches/{id}/process ตาม T-36 และ request/error mapping
- **Dependencies:** T-38
- **DoD:** `cd backend; go test ./...` ผ่านทั้งหมด; ไม่แก้ *_test.go

### T-40 — Implement: UI ปรับกระบวนการ

- **สิ่งที่ทำ:** สร้างหน้าและเมนู process ตาม T-37 ผ่าน generated client พร้อม current/next/terminal behavior
- **Dependencies:** T-39
- **DoD:** `cd frontend; npm run lint` และ `cd frontend; npm run build` ผ่าน; E2E ที่เขียนไว้ผ่านทุกข้อ พร้อมหลักฐาน; ตรวจ responsive, keyboard, focus, contrast, reduced motion และไม่มี direct fetch ใน pages/components; AC-12 ทั้ง backend/UI ครบใน slice นี้

### T-41 — Hardening: quality gates บนเครื่อง

- **สิ่งที่ทำ:** ตรวจ AC coverage ทุก bullet ของ SPEC และ E2E ที่ปิดครบแล้ว; รัน lint/security/tests/client drift โดยไม่ย้าย AC ที่ขาดจาก slice มาเพิ่งทำที่นี่; defect กลับไป failing regression test ก่อน fix
- **Dependencies:** T-40
- **DoD:** `npx @redocly/cli lint docs/openapi.yaml`; `cd backend; go test ./...`, `golangci-lint run`, `gosec ./...`, `govulncheck ./...`; `cd frontend; npm run generate:api`, `npm run lint`, `npm run build`; `git diff --exit-code -- frontend/src/api/generated` จาก root ผ่าน; E2E ทุก slice ผ่าน loading/error/empty/success

### T-42 — Hardening: CI บน GitHub Actions

- **สิ่งที่ทำ:** เพิ่ม workflow clean checkout pin Go/Node และ tool versions; รัน gates เดียวกับ T-41 รวม generated drift; ใช้ test DB/files ชั่วคราว ไม่ใช้ข้อมูลจริงหรือ secrets ใน fixture
- **Dependencies:** T-41
- **DoD:** workflow ของ branch นี้บน GitHub Actions ผ่าน OpenAPI lint, go test ./..., golangci-lint, gosec, govulncheck, generate/drift, frontend lint/build; ทดลอง failing test ใน branch ทดลองแยกแล้ว CI ต้องแดงโดยไม่แก้ tests หลัก; ไม่มีการ merge main อัตโนมัติ

### T-43 — Hardening: cross-agent review ด้วย session ใหม่

- **สิ่งที่ทำ:** ให้ reviewer ใน session ใหม่อ่าน AGENTS/SPEC/PLAN/TASKS/OpenAPI และตรวจ AC coverage, boundaries, SQL/file invariants, errors, security และ UI; รายงาน findings พร้อมตำแหน่งและความรุนแรง
- **Dependencies:** T-42
- **DoD:** reviewer ยืนยัน coverage ครบและไม่มี high/critical findings ค้าง; findings ที่ต้องแก้เริ่มด้วย failing regression test; รันคำสั่งทั้งหมดใน T-41 และ GitHub Actions ตาม T-42 ผ่านหลังแก้
