# DESIGN BRIEF — Tracking Research Status MVP

สถานะ: Draft สำหรับใช้กำกับการออกแบบ UX/UI ของ MVP

## 1. ลำดับแหล่งความจริง

เอกสารนี้กำหนดทิศทาง UX/UI เท่านั้น และต้องไม่เปลี่ยนขอบเขตหรือพฤติกรรมของระบบ

1. `docs/SPEC.md` — ขอบเขตและพฤติกรรมของระบบ
2. `docs/openapi.yaml` — API contract และ error code
3. `docs/TASKS.md` — ขอบเขตของ task ที่กำลังทำ
4. `docs/DESIGN_BRIEF.md` — ทิศทาง UX/UI เฉพาะของโปรเจกต์
5. รูปที่ลงทะเบียนใน `docs/design-references/` — ต้นแบบเฉพาะส่วนที่ระบุไว้ในเอกสารนี้
6. `.agents/skills/frontend-design/SKILL.md` — หลักการออกแบบสำหรับสิ่งที่เอกสารนี้ไม่ได้กำหนด

หากเอกสารขัดแย้งกัน ให้ยึดเอกสารที่มีลำดับสูงกว่า หากยังตัดสินไม่ได้ให้ถาม human ก่อน

## 2. เป้าหมายผลิตภัณฑ์

- ช่วยให้ผู้ใช้เห็นรายการงานวิจัยร่วมกันได้อย่างรวดเร็ว
- รองรับการเพิ่ม แก้ไข และลบงานวิจัยโดยไม่ทำให้ผู้ใช้หลงจากหน้ารายการ
- รองรับงานวิจัยต้นฉบับและงานวิจัยต่อเนื่อง รวมถึงรายการที่มี title ซ้ำ โดยใช้ ID แยกรายการภายในระบบอย่างถูกต้อง
- ทำให้สถานะ loading, empty, error และ success เข้าใจได้ทันที
- ใช้งานได้ดีทั้ง desktop และ mobile

ข้อมูลจาก API ประกอบด้วย `id`, `title`, `description`, `continuationOfId`, `status` และ `process` โดย UI ใช้ `id` เป็น identity ภายในสำหรับ update/delete และแยกรายการชื่อซ้ำ แต่ไม่ใช้ ID เป็นข้อมูลนำของหน้า

MVP ยังไม่มี authentication, role selector หรือหน้ารายละเอียดแยก และยังไม่มีหน้าสำหรับดูหรือปรับ `status`/`process` แม้ API จะรองรับข้อมูลและ operation เหล่านี้แล้ว

## 3. กลุ่มผู้ใช้

- ผู้ดูแลระบบและผู้ประสานงาน: ดู เพิ่ม แก้ไข และลบงานวิจัย
- นักวิจัย: ดูรายการงานวิจัย
- ทั้งสามบทบาทเห็นรายการเดียวกัน

MVP ยังไม่มี authentication จึงห้ามสร้างหน้าล็อกอิน ตัวเลือกบทบาท หรือระบบจำลองสิทธิ์ขึ้นเอง หน้า MVP แสดงความสามารถ CRUD ตาม task ที่กำหนดโดยไม่อ้างว่าได้บังคับสิทธิ์ตามตัวตนแล้ว

## 4. ทิศทางการออกแบบ

### Visual thesis

“ทะเบียนงานวิจัยแบบสมุดบันทึกสถาบันร่วมสมัย” — สุขุม แม่นยำ และโปร่งสบาย โดยใช้โครงสร้างคล้ายสารบัญงานวิชาการ เส้นแบ่งบาง และสีเขียวอมฟ้าเป็นจุดนำสายตาเพียงสีหลักเดียว

งานต้องรู้สึกเป็นเครื่องมือปฏิบัติงาน ไม่ใช่ landing page เชิงการตลาด และไม่ควรมี hero ขนาดใหญ่ dashboard card mosaic, gradient ตกแต่ง, glassmorphism หรือภาพประกอบที่ไม่ช่วยการทำงาน

### Signature element

ใช้ “catalog spine” เป็นเส้นแนวตั้งสี accent ที่ขอบซ้ายของหัวหน้ารายการ พร้อมข้อความจำนวนรายการ เช่น “งานวิจัย 12 รายการ” เพื่อช่วยการวางตำแหน่งและสร้างเอกลักษณ์ ห้ามใช้เลขลำดับกับแต่ละงานวิจัย เพราะลำดับไม่ใช่ข้อมูลถาวรของรายการ

### Content plan

1. Header ขนาดกะทัดรัด: ชื่อระบบ คำอธิบายหนึ่งบรรทัด และปุ่ม “เพิ่มงานวิจัย”
2. List workspace: จำนวนรายการ สถานะการโหลด และรายการ title/description/actions
3. Dialog เพิ่ม: ฟอร์ม `title`, `description` และการเลือกงานวิจัยต้นทางที่ map เป็น `continuationOfId`
4. Dialog แก้ไข: ฟอร์ม `title` และ `description` เท่านั้น เพราะ `continuationOfId` แก้ไขไม่ได้
5. Confirmation dialog: ยืนยันก่อนลบโดยแสดง title ที่กำลังจะลบ
6. Feedback region: error, retry และ success message ที่ screen reader รับรู้ได้

### Interaction thesis

- Dialog เปิดและปิดด้วย transition สั้น 120–160 ms เพื่ออธิบายการเปลี่ยนบริบท
- หลังเพิ่มหรือแก้ไขสำเร็จ ให้แถวที่เกี่ยวข้องมี highlight ชั่วคราวแบบสุภาพเพื่อช่วยค้นหารายการ
- การลบต้องไม่มี animation ยืดเยื้อ รายการหายหลัง API สำเร็จและแสดงข้อความยืนยัน
- ต้องเคารพ `prefers-reduced-motion`; เมื่อเปิดใช้ให้ลด transition และยกเลิก highlight animation

## 5. ภาษาและน้ำเสียง

- ภาษาหลักของ UI: ภาษาไทย
- น้ำเสียง: สุภาพ กระชับ ตรงไปตรงมา และไม่ใช้ศัพท์เทคนิคโดยไม่จำเป็น
- ใช้คำให้สม่ำเสมอ: “งานวิจัย”, “เพิ่มงานวิจัย”, “แก้ไข”, “บันทึกการแก้ไข”, “ลบงานวิจัย”, “ยกเลิก” และ “ลองอีกครั้ง”
- ปุ่มต้องบอกผลลัพธ์ของการกระทำ เช่น “เพิ่มงานวิจัย” แทน “ส่งข้อมูล”
- Error message ต้องอธิบายปัญหาและสิ่งที่ผู้ใช้ทำต่อได้ โดยตัดสินประเภทข้อความจาก error code ไม่ผูก logic กับข้อความจาก server

## 6. สีและ typography

### Color tokens

- `canvas` — `#F4F7F8`: พื้นหลังของแอป
- `surface` — `#FFFFFF`: พื้นที่รายการและ dialog
- `ink-strong` — `#17242D`: heading และข้อความสำคัญ
- `ink` — `#40515D`: body text
- `ink-muted` — `#687984`: supporting text
- `line` — `#D7E0E4`: เส้นแบ่งและขอบ
- `accent` — `#0B6B63`: primary action, focus และ catalog spine
- `accent-strong` — `#07534D`: hover/pressed
- `danger` — `#B42318`: destructive action และ destructive error
- `success` — `#147D64`: success feedback

ห้ามใช้สีเพียงอย่างเดียวเพื่อสื่อสถานะ ทุกสถานะต้องมีข้อความหรือสัญลักษณ์ที่มี accessible name และต้องตรวจ contrast ตาม WCAG AA ก่อน implementation เสร็จ

### Typography

- Display/heading: `IBM Plex Sans Thai`, น้ำหนัก 600–700
- Body/UI: `Noto Sans Thai`, น้ำหนัก 400–600
- Fallback: `Leelawadee UI`, `Tahoma`, sans-serif
- ใช้ไม่เกินสองตระกูลตัวอักษร และห้ามเพิ่ม font package โดยไม่จำเป็น หากยังไม่มีไฟล์ font ให้ใช้ fallback ก่อน
- Body เริ่มที่ 16 px, line-height 1.55; description ที่ยาวต้องอ่านง่ายและไม่ชิดเกินไป

## 7. Layout และ responsive behavior

- Content width สูงสุดประมาณ 1120 px และมีระยะขอบที่อ่านสบาย
- Header ไม่ต้องติดหน้าจอ และต้องไม่กินพื้นที่แนวตั้งมากเกินไป
- Primary action มีเพียง “เพิ่มงานวิจัย” และอยู่ขวาของ header บน desktop
- Desktop ใช้ list/table ที่มีคอลัมน์ title, description และ actions โดยให้ description ใช้พื้นที่หลัก
- ไม่แสดง ID เป็นคอลัมน์หลัก แต่ต้องใช้ ID เป็น key และเป็นตัวระบุใน update/delete; แสดง ID เป็นข้อความรองได้เฉพาะจุดที่จำเป็นต่อการแยกรายการ title ซ้ำ เช่นตัวเลือกงานวิจัยต้นทาง
- ห้ามเพิ่มคอลัมน์ status, process, owner, created date หรือข้อมูลที่อยู่นอกขอบเขต UI ของ MVP
- รายการเรียงตามลำดับที่ API ส่งกลับ ห้าม sort เพิ่มใน UI
- ที่ความกว้างไม่เกินประมาณ 720 px ให้เปลี่ยนแต่ละแถวเป็น stacked record: title → description → actions
- Mobile ต้องไม่มี horizontal scroll และ action target ต้องมีขนาดอย่างน้อย 44 × 44 px
- Description ใน list แสดงแบบจำกัดจำนวนบรรทัดได้ แต่ข้อมูลเต็มต้องปรากฏใน edit form และไม่สร้าง route รายละเอียดใหม่

## 8. หน้ารายการงานวิจัย

### Header

- ชื่อหน้าหลัก: “ระบบติดตามสถานะงานวิจัย”
- คำอธิบาย: “จัดการและดูรายการงานวิจัยที่ใช้งานร่วมกัน”
- ปุ่มหลัก: “เพิ่มงานวิจัย”

### แต่ละรายการ

- Title เป็นข้อมูลนำและต้องเด่นกว่า description
- Description รองรับ newline และข้อความยาวโดยไม่ทำให้ layout แตก
- รายการ title ซ้ำต้องยังเลือกแก้ไขหรือลบได้ถูก record โดยผูก action กับ `id` ไม่ผูกกับ title
- เมื่อ title ซ้ำและ description ยังแยกไม่ชัด ให้แสดง ID เป็น metadata รองเฉพาะรายการที่จำเป็น ห้ามนำ ID มาแทน title หรือใช้เป็นเลขลำดับตกแต่ง
- ไม่แสดงหรือเพิ่ม action สำหรับ status/process ในหน้ารายการของ MVP
- Actions ใช้ข้อความ “แก้ไข” และ “ลบ”; ไม่ใช้ icon เพียงอย่างเดียว
- “ลบ” เป็น secondary destructive action ไม่แข่งขันทางสายตากับ primary action ของหน้า
- ห้ามเพิ่ม action “ดูรายละเอียด” ใน MVP

## 9. ฟอร์มเพิ่มและแก้ไข

- ใช้ dialog หรือ side panel บนหน้ารายการ ไม่สร้าง route ใหม่
- แสดง label ที่มองเห็นได้สำหรับทุก field; placeholder ใช้เป็นตัวอย่างเท่านั้น
- `title` เป็น single-line input และแสดงตัวนับสูงสุด 200 ตัวอักษร
- `description` เป็น textarea และแสดงตัวนับสูงสุด 5,000 ตัวอักษร
- Create ต้องมีตัวเลือก “งานวิจัยต้นฉบับ” ซึ่งส่ง `continuationOfId: null` และตัวเลือกงานวิจัยต้นทางซึ่งส่ง positive ID ของงานที่มีอยู่
- รายการงานวิจัยต้นทางต้องรวมโครงการที่เสร็จสิ้นหรือยุติแล้ว และแสดง title ร่วมกับ ID แบบข้อความรองเพื่อแยกรายการ title ซ้ำ
- Create ห้ามส่ง `id`, `status` หรือ `process`; ค่า status/process เริ่มต้นเป็นหน้าที่ของระบบ
- Edit ส่งเฉพาะ `title` และ `description`; ห้ามให้ผู้ใช้เปลี่ยน `id`, `continuationOfId`, `status` หรือ `process` ผ่านฟอร์มนี้
- ระบุ field ที่จำเป็นด้วยข้อความ ไม่พึ่งเครื่องหมายดอกจันเพียงอย่างเดียว
- ขณะ submit ให้ disable ปุ่ม action ที่เกี่ยวข้องและแสดงข้อความ “กำลังบันทึก…” เพื่อป้องกันการส่งซ้ำ
- Validation error แสดงใกล้ field ที่เกี่ยวข้องและ focus ไปยัง field แรกที่ผิด
- Create สำเร็จ: ปิด dialog, refresh list, แสดงรายการใหม่และ success feedback
- Edit สำเร็จ: ปิด dialog, refresh list และแสดง title ใหม่แทน title เดิม
- ยกเลิกต้องไม่เปลี่ยนข้อมูล หากมีข้อมูลที่แก้แต่ยังไม่บันทึกและกำลังจะปิด dialog ให้เตือนก่อนทิ้งเมื่อเหมาะสม

## 10. การยืนยันการลบ

- ต้องมี confirmation dialog ก่อนลบทุกครั้ง
- แสดง title ของงานวิจัยที่กำลังจะลบอย่างชัดเจน
- หาก title ซ้ำ ให้แสดง ID และข้อความย่อของ description เป็นบริบทรองเพื่อยืนยันว่าเป็น record ที่ถูกต้อง
- แจ้งว่า “การลบนี้ไม่สามารถย้อนกลับได้”
- ปุ่มอันตราย: “ลบงานวิจัย”; ปุ่มรอง: “ยกเลิก”
- Focus เริ่มที่ปุ่ม “ยกเลิก” ไม่ใช่ปุ่มลบ
- ขณะลบให้ disable actions เพื่อป้องกัน request ซ้ำ
- ลบสำเร็จ: ปิด dialog, นำรายการออก และแสดง success feedback

## 11. UI states

### Loading

- โหลดรายการครั้งแรกด้วย skeleton rows ที่มีโครงใกล้เคียงข้อมูลจริงและไม่ทำให้ layout กระโดด
- การเพิ่ม แก้ไข หรือลบ แสดง loading เฉพาะ action ที่กำลังทำ
- ป้องกันการส่ง request เดิมซ้ำขณะ loading

### Empty

- Heading: “ยังไม่มีงานวิจัย”
- Supporting text อธิบายสั้น ๆ ว่าสามารถเริ่มเพิ่มรายการแรกได้
- แสดงปุ่ม “เพิ่มงานวิจัย” โดย empty state ต้องไม่ดูเหมือน error

### Error

- List API ล้มเหลว: แสดง error region พร้อมปุ่ม “ลองอีกครั้ง” โดยยังคง header ของหน้าไว้
- `VALIDATION_ERROR`: แสดง inline ใต้ field เมื่อระบุ field ได้
- `TITLE_ALREADY_EXISTS`: focus ช่อง title และแจ้งว่า “มีชื่องานวิจัยนี้แล้ว”
- `CONTINUATION_NOT_FOUND`: แจ้งว่างานวิจัยต้นทางไม่มีอยู่แล้ว ให้ refresh ตัวเลือกและเลือกใหม่
- `RESEARCH_NOT_FOUND`: แจ้งว่ารายการไม่มีอยู่แล้ว ปิด dialog ที่เกี่ยวข้องและ refresh list
- `RESEARCH_HAS_CONTINUATIONS`: แจ้งว่ายังลบไม่ได้เพราะมีงานวิจัยต่อเนื่องอ้างถึง และคงรายการไว้
- `PAYLOAD_TOO_LARGE`: แจ้งให้ลดความยาวข้อมูล
- `INTERNAL_ERROR` หรือ network failure: แจ้งว่าไม่สามารถดำเนินการได้และให้ลองใหม่
- ห้ามแสดง stack trace หรือรายละเอียดภายในระบบ

### Success

- แสดง toast หรือ inline status หลังเพิ่ม แก้ไข หรือลบสำเร็จ
- Feedback ใช้ `aria-live="polite"`, ไม่บัง primary action และปิดเองได้
- ข้อความใช้ชื่อ action เดิม เช่น “เพิ่มงานวิจัยแล้ว”, “บันทึกการแก้ไขแล้ว” และ “ลบงานวิจัยแล้ว”

## 12. Accessibility

- ใช้ semantic HTML, heading order ที่ถูกต้อง และ landmark ที่จำเป็น
- ทุก control มี accessible name และทุก field เชื่อม label/error/help text อย่างถูกต้อง
- รองรับ keyboard navigation, Escape เพื่อปิด dialog เมื่อปลอดภัย และ focus trap ภายใน modal
- เมื่อปิด dialog ให้คืน focus ไปยัง control ที่เปิด dialog
- Focus indicator ต้องเห็นชัดอย่างน้อย 2 px และไม่ถูกตัด
- Touch target อย่างน้อย 44 × 44 px
- ไม่พึ่งสีอย่างเดียว และทดสอบ contrast, zoom 200%, reduced motion และ screen reader announcement

## 13. รูปต้นแบบและสินทรัพย์อ้างอิง

ขณะร่างเอกสารนี้ยังไม่มีรูปต้นแบบที่ได้รับการยืนยัน ห้าม AI สร้างรูปอ้างอิงขึ้นเองแล้วถือเป็นข้อกำหนด

เมื่อมีรูปต้นแบบ ให้เก็บไว้ใน `docs/design-references/` และเพิ่มรายการในส่วนนี้ โดยระบุให้ชัดเจนสำหรับแต่ละรูปว่า:

- ใช้กับหน้าจอหรือ breakpoint ใด
- ส่วนใดต้องยึดตาม เช่น hierarchy, spacing หรือ interaction
- ส่วนใดเป็นเพียงแรงบันดาลใจ
- ส่วนใดห้ามคัดลอก เช่น brand, logo, ข้อความ หรือข้อมูลเฉพาะของต้นฉบับ

รูปแบบชื่อไฟล์ที่แนะนำ:

- `research-list-desktop.png`
- `research-list-mobile.png`
- `research-form-dialog.png`
- `research-delete-confirmation.png`

หากข้อความใน Design Brief ขัดกับรูปต้นแบบ ให้ยึดข้อความใน Design Brief และถาม human เมื่อความขัดแย้งกระทบพฤติกรรมสำคัญ

## 14. Definition of Done ด้าน UX/UI

- หน้ารายการมี loading, error, empty และ success ครบ
- Create, edit และ delete มี submitting/error/success และป้องกันการทำซ้ำ
- ทุก record และ action ใช้ `id` เป็น identity ภายใน แม้ title ซ้ำ; ไม่ใช้ title เป็น path identifier
- Create ส่ง `continuationOfId` ทุกครั้ง โดยงานต้นฉบับส่ง `null`; Edit ไม่เปิดให้แก้ field นี้
- ไม่มีคอลัมน์หรือหน้าสำหรับ status/process, authentication, role selector หรือ route ที่นอก MVP
- ไม่มี page/component เรียก `fetch` โดยตรง
- ใช้ generated typed API client และ map error ด้วย error code
- ใช้งานด้วย keyboard ได้ครบและ focus ไม่สูญหายหลัง dialog ปิด
- ไม่มี horizontal overflow ที่ mobile และข้อความยาวไม่ทำให้ layout แตก
- ตรวจ narrow mobile, common desktop, wide desktop, reduced motion และ zoom 200%
- `npm run lint` และ `npm run build` ผ่านก่อนถือว่างาน UI เสร็จ
