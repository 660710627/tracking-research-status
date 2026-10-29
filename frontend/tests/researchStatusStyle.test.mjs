import test from 'node:test'
import assert from 'node:assert/strict'
import { researchStatusClass } from '../src/researchStatusStyle.ts'

test('research badges distinguish extension rounds, completion and termination', () => {
  const expected = [
    ['กำลังดำเนินการ', 'status'],
    ['กำลังดำเนินการ(ขยายเวลาครั้งที่ 1)', 'status warn'],
    ['กำลังดำเนินการ(ขยายเวลาครั้งที่ 2)', 'status research-extended-twice'],
    ['กำลังดำเนินการ(ขยายเวลามากกว่า 2 ครั้ง)', 'status research-extended-many'],
    ['โครงการเสร็จสิ้น', 'status research-completed'],
    ['ยุติโครงการ', 'status research-terminated'],
  ]
  for (const [status, className] of expected) assert.equal(researchStatusClass(status), className)
})
