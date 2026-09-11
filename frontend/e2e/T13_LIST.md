# T-13: E2E หน้ารายการงานวิจัย

## ขอบเขตและแหล่งอ้างอิง
AGENTS.md, docs/SPEC.md AC-07 และ UI ของ AC-02/14, docs/PLAN.md, docs/TASKS.md T-13, docs/openapi.yaml GET /api/v1/researches และ .agents/skills/design-taste-frontend/SKILL.md
Skill ระบุว่าตารางข้อมูลไม่ใช่ขอบเขตหลัก จึงใช้ SPEC สำหรับคอลัมน์และพฤติกรรม ใช้หลัก clarity, focus, contrast, responsive และ reduced motion ที่เกี่ยวข้องเท่านั้น ไม่เพิ่ม UI framework หรือฟีเจอร์
คอลัมน์ตาม T-01: เลขสัญญาทุน, ประเภทโครงการหลัก/ต่อเนื่อง, ชื่อ, สถานะ, กระบวนการ ไม่แสดง ID หรือ description
ไม่ทดสอบค้นหา/กรองหรือ implement mutation ใน slice นี้

## วิธีรันและผลจริง (2026-09-11)
1. ติดตั้ง dependencies ตาม lockfile: `& 'C:\Program Files\nodejs\npm.cmd' ci --no-audit --no-fund` ภายใน frontend ผล exit 0, added 209 packages
2. ภายใน frontend: `node node_modules/vite/bin/vite.js --host=127.0.0.1 --port=5173 --strictPort` ผล Vite v8.1.5 ready in 794 ms
3. เปิด http://127.0.0.1:5173/ ใน in-app browser สำเร็จ ตรวจ accessibility tree และ screenshot ที่ 375×667 และ 1280×720 กด Tab บน desktop แล้ว reset viewport
4. R0: ทั้งสองขนาดมี heading “ระบบติดตามสถานะงานวิจัย” และ “อยู่ระหว่างเตรียมระบบใหม่” เท่านั้น ไม่มีตาราง จำนวนรายการ ปุ่ม retry หรือ controls ของรายการ กด Tab ไม่พบ control ของรายการ
5. ภาพหน้าเริ่มต้นข้อความไม่ถูกตัดทั้งสองขนาด แต่ยังไม่ยืนยัน responsive ของตารางที่ยังไม่มี
6. ตรวจ source: App.tsx เป็น static placeholder; index.css ไม่มี animation/transition ของรายการ ไม่มี controls ให้ตรวจ focus/contrast และยังไม่ได้ emulate reduced motion
7. `rg -n 'fetch\s*\(' frontend/src` พบเฉพาะ _fetch ใน generated client/core ไม่มี direct fetch ใน App/pages/components ปัจจุบัน ยังต้องตรวจซ้ำหลัง T-14

การเปิดเว็บที่ล้มเหลวก่อนติดตั้ง dependencies ไม่นับเป็น RED ผล RED อ้างหน้า React ที่เปิดสำเร็จแล้วเท่านั้น
DoD เป็น manual E2E checklist ไม่มีคำสั่ง automated test กำหนดไว้ ไม่ใช้ lint/build แทน E2E
ขั้นปลายทางที่ต้องมี list UI ยังไม่ได้ execute: ไม่อ้างว่าจำลอง API delay/error/retry หรือ mutation แล้ว ไม่ถือว่า accessibility ของฟีเจอร์ผ่าน

## Fixture และการเตรียมรอบ GREEN
ไฟล์ fixtures/T13_LIST.json เป็น test data เท่านั้น ไม่ import เข้า production
- success: สามรายการ ID ภายใน 101,108,120 เรียงชื่อ ก,ก,ข; 108 เป็น continuation ของ 101 ชื่อเหมือนกัน แต่เลขสัญญา CN-101/CN-108 ต่างกัน
- ทุก record มี fields ใหม่ สมาชิกสองคน PDF metadata วันที่ พ.ศ. ครบ ไม่มี bytes/path/description
- empty: []; error: response 500 ตาม OpenAPI; network failure ใช้ abort request แยกจาก HTTP error
- afterCreate: จำนวนสี่รายการ เพิ่ม ID 125; afterUpdate: จำนวนสามรายการ แก้เฉพาะ ID 108 โดยคง ID/ชื่อเดิม
- changedIdAfterCreate/changedIdAfterUpdate เป็น metadata ของ test harness ห้ามส่งปะปนใน GET response
- รอบ GREEN ใช้ network interception ของ test harness คืนเฉพาะ array ที่เลือกให้ GET /api/v1/researches ก่อนเปิดหน้า คืน JSON Content-Type; ไม่เพิ่ม endpoint หรือ mock code ลง production
- loading: hold GET ไว้ 2 วินาที แล้วปล่อย success; retry: 500 ครั้งแรกแล้ว success ครั้งถัดไป
- หลัง mutation: inject ผลสำเร็จและ changed ID ผ่าน test harness ตาม interface ของ UI ที่ T-14 สร้าง แล้วคืน afterCreate/afterUpdate ใน GET ถัดไป ไม่ต้องสร้างแบบฟอร์มหรือยิง mutation จริงใน T-13
- ทุก case reset fixture และหน้า ทดสอบ viewport ทั้งสองขนาดเมื่อเกี่ยวกับ layout

## Test cases
R0 = RED ที่ prerequisite ไม่มีหน้ารายการ ขั้นต่อจากนั้นยังไม่ได้รัน ต้องทดลองซ้ำหลัง T-14

| Case | AC / ขั้นตอน | Expected | Actual |
|---|---|---|---|
| L01 | AC-07 เปิดหน้า success | โหลด GET ผ่าน generated client โดยไม่มี query/body แสดงรายการ | R0 |
| L02 | AC-07 hold GET แล้วปล่อย | เห็น loading ระหว่างรอ ไม่แสดง empty/error ก่อนผล; จบแล้ว success | R0; delay ยังไม่ได้ฉีด |
| L03 | AC-07 คืน 500 | error ที่อ่านได้พร้อมลองใหม่ ไม่มี SQL/path ไม่ตีความเป็นไม่มีข้อมูล | R0 |
| L04 | AC-07 abort network | error แยกจาก empty ไม่ค้าง loading | R0 |
| L05 | AC-07 error แล้ว retry คืน success | ใช้ปุ่มลองใหม่ได้ โหลดใหม่และล้าง error แสดงสามรายการ | R0 |
| L06 | AC-07 คืน empty | empty state และจำนวน 0 ไม่แสดง network error | R0 |
| L07 | AC-07 success | คอลัมน์เลขสัญญา ประเภทหลัก/ต่อเนื่อง ชื่อ สถานะ กระบวนการครบ จำนวน 3; ไม่มี description | R0 |
| L08 | AC-07 success ชื่อซ้ำ | ลำดับ CN-101,CN-108,CN-120 ไม่มีรวม/หาย/สลับรายการที่ชื่อซ้ำ | R0 |
| L09 | AC-02/07 ตรวจ DOM และ source identity | ใช้ ID เป็น key/ตัวระบุภายใน ไม่มี ID column/input/search และไม่ใช้ title เป็น identity | R0 |
| L10 | AC-07 ตรวจ response mapping | fields โครงการใหม่ สมาชิก metadata วันที่ งบไม่สูญหาย/ปะปนระหว่าง rows; ตารางแสดงตามคอลัมน์กำหนด ไม่บังคับยัดทุก field เป็นคอลัมน์ | R0 |
| L11 | AC-07 inject afterCreate + changed ID | จำนวน 4 แสดง CN-125 พร้อม feedback/การระบุรายการใหม่หลัง success | R0 |
| L12 | AC-07 inject afterUpdate + changed ID | จำนวน 3 ข้อมูลล่าสุดเป็นของ 108 เน้น CN-108 เท่านั้น แม้ชื่อเหมือน 101; ID ไม่เปลี่ยน | R0 |
| L13 | AC-07 refresh หลัง fixture เปลี่ยน | แสดงค่าล่าสุด ไม่มี stale rows หรือรายการซ้ำเพิ่ม | R0 |
| L14 | AC-14 เข้า/กลับรายการผ่าน navigation ที่อยู่ใน slice | เข้าถึงรายการได้และระบุหน้าปัจจุบัน; sidebar ขยาย/ยุบไม่บังเนื้อหา เมื่อมี sidebar | R0 |
| L15 | AC-14 375×667 / 1280×720 | ตาราง/ข้อความ/ปุ่มไม่ตัด ไม่มี scroll แนวนอนทั้งหน้า ตารางเลื่อนเฉพาะ container ได้ | R0 ทั้งสองขนาด; placeholder มองเห็นครบ |
| L16 | AC-14 Tab/Shift+Tab Enter/Space | controls มีชื่อและ focus ชัด ลำดับตรงภาพ retry ใช้ keyboard ได้ ไม่มี trap; Escape/focus return หากมี popup | R0; ทดลอง Tab แล้วไม่มี controls |
| L17 | AC-14 contrast success/error/loading/changed row | text >=4.5:1, large text/UI >=3:1 วัด foreground/background จริงรวม focus ไม่อาศัยสีอย่างเดียว | R0; ยังไม่มีองค์ประกอบรายการให้วัด |
| L18 | AC-14 prefers-reduced-motion reduce | loading/feedback ยังสื่อความหมาย ปิด animation ไม่จำเป็น transition จำเป็นทันที | R0; ไม่มี motion ของรายการ ยังไม่ emulate |
| L19 | AC-07/14 long title/long contract และ 30 records จากสำเนา fixture ที่ ID/contract ไม่ซ้ำ | อ่านข้อความครบ layout ไม่พัง จำนวนถูกต้อง ไม่แสดง internal ID | R0 |
| L20 | PLAN ตรวจ pages/components และ client imports | ไม่มี direct fetch ใช้ generated client ไม่แก้ generated code ด้วยมือ | PASS source ปัจจุบันเฉพาะ no-direct-fetch; การเชื่อม list ยัง R0 |

## สรุปผล
RED ยืนยันจากเว็บจริงที่ยังไม่มี list UI ไม่ใช่ server failure
Checklist และ fixtures พร้อมสำหรับ T-14; cases ที่ติด prerequisite บันทึกตามจริง ไม่มีการเขียน feature code หรือเปลี่ยน tests เดิม

