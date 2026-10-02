import test from 'node:test'
import assert from 'node:assert/strict'
import { normalizeSdgs } from '../src/sdgs.ts'
import { changeResearch, changeResearchSdgs } from '../src/research.ts'

test('SDG selection is deduplicated and sorted without mutating its draft',()=>{
  const draft=[13,6,13,9]
  assert.deepEqual(normalizeSdgs(draft),[6,9,13])
  assert.deepEqual(draft,[13,6,13,9])
})
test('SDG selection rejects invalid goal identifiers',()=>{
  for(const invalid of [[0],[18],[-1],[1.5],['6'],[NaN],null])assert.throws(()=>normalizeSdgs(invalid))
})

const research={id:101,title:'งานวิจัย',contract:'A-01',lead:'หัวหน้า',unit:'คณะ',budget:100,status:'กำลังดำเนินการ',process:3,kind:'โครงการหลัก',endDate:'30 ก.ย. 2570',sdgs:[6]}
test('at least one SDG is required for creation and SDG updates',()=>{
  assert.throws(()=>normalizeSdgs([]),/อย่างน้อย 1/)
  assert.throws(()=>changeResearchSdgs([research],101,'นักวิจัย',[]),/อย่างน้อย 1/)
  assert.deepEqual(research.sdgs,[6])
})
test('all three roles can change only SDGs, including completed projects',()=>{
  for(const role of ['นักวิจัย','ผู้ประสานงาน','ผู้ดูแลระบบ'])for(const status of ['กำลังดำเนินการ','โครงการเสร็จสิ้น']){
    const current={...research,status}
    const other={...research,id:102}
    const [saved,unchanged]=changeResearchSdgs([current,other],101,role,[13,6,13])
    assert.deepEqual(saved,{...current,sdgs:[6,13]})
    assert.equal(unchanged,other)
    assert.deepEqual(current.sdgs,[6])
  }
  assert.throws(()=>changeResearch([research],101,'นักวิจัย',{type:'edit',fields:{...research,title:'เปลี่ยนชื่อ'}}),/ไม่มีสิทธิ์/)
})
test('terminated projects reject SDG updates for every role',()=>{
  for(const role of ['นักวิจัย','ผู้ประสานงาน','ผู้ดูแลระบบ'])assert.throws(()=>changeResearchSdgs([{...research,status:'ยุติโครงการ'}],101,role,[9]),/ยุติแล้ว/)
  assert.throws(()=>changeResearchSdgs([research],999,'ผู้ดูแลระบบ',[9]),/ไม่พบ/)
  assert.throws(()=>changeResearchSdgs([research],101,'unknown',[9]),/ไม่มีสิทธิ์/)
})
test('general project edits cannot overwrite SDGs',()=>{
  const [saved]=changeResearch([research],101,'ผู้ดูแลระบบ',{type:'edit',fields:{...research,title:'ชื่อใหม่',sdgs:[]}})
  assert.deepEqual(saved.sdgs,[6])
})
