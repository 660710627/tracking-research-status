# My research Agent Context

## What this project is
ระบบติดตามสถานะงานวิจัย: เพิ่มงานวิจัย, ดูรายการ, ลบงานวิจัย, แก้ไขข้อมูล, การปรับสถานะและกระบวนการของงานวิจัย, มี user 3 บทบาท 1.นักวิจัย 2.ผู้ประสานงาน 3.ผู้ดูแลระบบ
กฎเหล็ก:
- งานวิจัยจะถูกเพิ่มเข้ามาในระบบโดยผู้ประสานงานหรือผู้ดูแลระบบ และนักวิจัยไม่ต้องสามารถเพิ่มงานวิจัย ลบงานวิจัย แก้ไขข้อมูลและการปรับสถานะและกระบวนการของงานวิจัยได้
- ชื่องานวิจัยสามารถซ้ำกันได้หากงานวิจัยดังกล่าวถูกระบุว่าเป็นงานวิจัยต่อเนื่อง (งานวิจัยต่อเนื่อง คืองานวิจัยที่ถูกนำมาต่อยอดจากงานวิจัยก่อนหน้า)
- งานวิจัยแต่ละงานต้องมี id แบบ immutable ซึ่งระบบเป็นผู้สร้างเป็น integer บวกที่ไม่ซ้ำ Client ห้ามกำหนดหรือแก้ไข id, id ต้องไม่เปลี่ยนเมื่อข้อมูลอื่นเปลี่ยน และห้ามนำ id ของงานวิจัยที่ลบแล้วกลับมาใช้ใหม่ ทั้งนี้ลำดับ id สามารถมีช่องว่างได้
- การปรับสถานะและกระบวนการของงานวิจัย : ไม่สามารถย้อนสถานะงานวิจัยได้หากงานวิจัยถูกเปลี่ยนสถานะไปข้างหน้าแล้ว (กระบวนการจะถูกปรับเปลี่ยนบ่อยที่สุด ดังนั้นควรออกแบบให้ปรับง่าย)
- ประเถทสถานะ มีดังนี้
  1.กำลังดำเนินการ
  2.กำลังดำเนินการ(ขยายเวลาครั้งที่ 1)
  3.กำลังดำเนินการ(ขยายเวลาครั้งที่ 2)
  4.กำลังดำเนินการ(ขยายเวลามากกว่า 2 ครั้ง)
  5.โครงการเสร็จสิ้น
  6.ยุติโครงการ
- ประเภทกระบวนการ มีดังนี้
  1.สัญญาโครงการ
  2.บันทึกข้อตกลง
  3.เปิดบัญชีธนาคาร
  4.การเบิกจ่ายเงิน
  5.การจัดสรรค่าธรรมเนียม
  6.การติดตามส่งรายงาน
  7.รายงานสรุปการใช้เงิน
  8.การปิดบัญชีธนาคาร
  (ทั้งนี้ในแต่ละกระบวนการยังมีกระบวนการย่อยๆอีก แต่จะถูกเพิ่มเข้ามาภายหลัง)

## Tech stack
- Backend: Go 1.22+, Gin, SQLite ผ่าน modernc.org/sqlite (ไฟล์ library.db)
- Backend tests: go testing + net/http/httptest
- Frontend: React + Vite + TypeScript (โฟลเดอร์ frontend/)
- API contract: docs/openapi.yaml คือแหล่งความจริงเดียว
- Lint/Security: golangci-lint, gosec, govulncheck / ESLint, tsc

## Project layout
- backend/cmd/server/       main.go
- backend/internal/handler/ HTTP layer (Gin) รู้จักแค่ HTTP
- backend/internal/service/ business rules
- backend/internal/repo/    SQL เท่านั้น
- backend/internal/db/      connection + schema
- frontend/src/             React app (api/, components/, pages/)
- docs/                     SPEC.md, PLAN.md, TASKS.md, openapi.yaml ← อ่านก่อนเริ่มงานทุกครั้ง
- docs/DESIGN_BRIEF.md      ทิศทาง UX/UI ของโปรเจกต์ ← อ่านก่อนเริ่ม task ออกแบบหรือ UI
- .agents/skills/frontend-design/SKILL.md  หลักการออกแบบ frontend ← ใช้ร่วมกับ Design Brief

## UX/UI design references
- ก่อนทำ task ที่เกี่ยวกับการออกแบบหรือ UI ต้องอ่าน `docs/DESIGN_BRIEF.md` และ `.agents/skills/frontend-design/SKILL.md` ให้ครบ
- ลำดับแหล่งความจริงสำหรับงาน UI คือ `docs/SPEC.md` → `docs/openapi.yaml` → `docs/TASKS.md` → `docs/DESIGN_BRIEF.md` → รูปที่ลงทะเบียนใน `docs/design-references/` → `.agents/skills/frontend-design/SKILL.md`
- Design Brief, รูปอ้างอิง และ skill ใช้กำหนด UX, visual direction, content, responsive behavior และ accessibility เท่านั้น ห้ามเปลี่ยนขอบเขต กฎธุรกิจ endpoint หรือ Out of scope
- ใช้รูปต้นแบบเฉพาะไฟล์ที่เก็บใน `docs/design-references/` และมีรายการกำกับใน `docs/DESIGN_BRIEF.md`; หากยังไม่มี ห้ามสมมติว่ามีต้นแบบที่ได้รับอนุมัติ
- หากแหล่งอ้างอิงขัดกัน ให้ยึดแหล่งที่อยู่ลำดับสูงกว่า และถาม human ก่อนเมื่อความขัดแย้งกระทบพฤติกรรมหรือขอบเขต

## Rules (must follow)
1. Plan ก่อน code ห้ามเขียนโค้ดก่อนเสนอแผนสั้น ๆ ให้ human เห็น
2. ทำทีละ 1 task จาก docs/TASKS.md เท่านั้น ห้ามทำเกินขอบเขต task
3. Test ต้องมีก่อนหรือพร้อมโค้ดเสมอ และห้ามลบ/แก้ test เพื่อให้ผ่าน
4. Business rule บังคับใช้ที่ database constraint ไม่ใช่แค่ใน application code
5. ทุก endpoint ต้องตรงกับ docs/openapi.yaml ถ้าต้องเปลี่ยน ให้แก้ contract ก่อนแล้วถาม human
6. Error response ใช้รูปแบบเดียวทั้งระบบ: {"error": {"code": "...", "message": "..."}}
7. มีคำถามหรือความกำกวม ให้ถาม human ก่อน ห้ามเดา
8. งานออกแบบและ UI ต้องยึด design references ตามหัวข้อ UX/UI design references และตรวจ loading, error, empty, success, responsive, keyboard, focus, contrast และ reduced motion

## Commands
- Run backend:  cd backend ; go run ./cmd/server   (port 8080)
- Backend test: cd backend ; go test ./...
- Backend lint: cd backend ; golangci-lint run ; gosec ./... ; govulncheck ./...
- Run frontend: cd frontend ; npm run dev          (port 5173)
- Frontend check: cd frontend ; npm run lint ; npm run build
