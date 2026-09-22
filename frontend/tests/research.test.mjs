import test from 'node:test'
import assert from 'node:assert/strict'
import { changeResearch, nextStatuses, researchDraft, statuses } from '../src/research.ts'

const item = {id:101,title:'โครงการเดิม',contract:'A-01',lead:'หัวหน้า',unit:'คณะ',budget:100,status:statuses[0],process:3,kind:'โครงการหลัก',endDate:'30 ก.ย. 2570'}
test('researchers cannot edit, delete, change status or process', () => {
  for (const action of [{type:'edit',fields:item},{type:'delete'},{type:'status',status:statuses[1]},{type:'process',process:4}]) {
    assert.throws(() => changeResearch([item],101,'นักวิจัย',action), /ไม่มีสิทธิ์/)
  }
})
test('every status transition respects forward order and terminal locks', () => {
  for (let from=0; from<6; from++) for (let to=0; to<6; to++) {
    const current={...item,status:statuses[from]}
    if (from<4 && to>from) {
      const [changed]=changeResearch([current],101,'ผู้ประสานงาน',{type:'status',status:statuses[to]})
      assert.equal(changed.status,statuses[to])
      assert.equal(changed.process,3)
      assert.equal(changed.id,101)
    } else assert.throws(() => changeResearch([current],101,'ผู้ดูแลระบบ',{type:'status',status:statuses[to]}))
  }
  assert.deepEqual(nextStatuses('unknown'),[])
})
test('editing preserves identity, status and process, and rejects invalid or duplicate data', () => {
  const other={...item,id:102,title:'อีกโครงการ',contract:'A-02'}
  const fields={...item,id:999,status:statuses[5],process:8,title:'ชื่อใหม่',budget:200}
  const [changed]=changeResearch([item,other],101,'ผู้ดูแลระบบ',{type:'edit',fields})
  assert.deepEqual(changed,{...item,title:'ชื่อใหม่',budget:200})
  for (const patch of [{title:' '},{budget:-1},{budget:NaN},{contract:' a-02 '},{title:'อีกโครงการ'}]) {
    assert.throws(() => changeResearch([item,other],101,'ผู้ดูแลระบบ',{type:'edit',fields:{...item,...patch}}))
  }
  assert.equal(item.title,'โครงการเดิม')
})
test('delete removes only the selected project, including the final item', () => {
  assert.deepEqual(changeResearch([item],101,'ผู้ดูแลระบบ',{type:'delete'}),[])
  assert.throws(() => changeResearch([item],999,'ผู้ดูแลระบบ',{type:'delete'}),/ไม่พบ/)
})
test('process changes preserve status and stop at terminal projects or the final step', () => {
  assert.equal(changeResearch([item],101,'ผู้ประสานงาน',{type:'process',process:4})[0].status,statuses[0])
  for (const process of [2,3,5,9]) assert.throws(() => changeResearch([item],101,'ผู้ดูแลระบบ',{type:'process',process}))
  for (const status of statuses.slice(4)) assert.throws(() => changeResearch([{...item,status}],101,'ผู้ดูแลระบบ',{type:'process',process:4}))
})

test('inline edit draft is independent and saves every displayed project detail', () => {
  const original={...item,objectives:['เดิม'],keywords:['เดิม'],contractFile:new File(['%PDF-demo'],'old.pdf',{type:'application/pdf'})}
  const draft=researchDraft(original)
  draft.objectives.push('ร่างที่ยังไม่บันทึก')
  draft.keywords[0]='คำค้นใหม่'
  assert.deepEqual(original.objectives,['เดิม'])
  assert.deepEqual(original.keywords,['เดิม'])
  const details={projectType:'บริการวิชาการ',isSubsidized:false,budgetUnit:'หน่วยงบประมาณใหม่',startDate:'1 ม.ค. 2569',collaborators:'ผู้ร่วมใหม่',fundingType:'ภายนอก',fund:'ทุนใหม่',thai:'บทคัดย่อใหม่',english:'New abstract',objectives:['เป้าหมายใหม่'],keywords:['คำค้นใหม่'],contractFile:new File(['%PDF-demo-new'],'new.pdf',{type:'application/pdf'})}
  const [saved]=changeResearch([original],101,'ผู้ประสานงาน',{type:'edit',fields:{...draft,...details,kind:'โครงการต่อเนื่อง',id:999,status:statuses[5],process:8}})
  for(const [key,value] of Object.entries(details))assert.deepEqual(saved[key],value)
  assert.equal(saved.kind,item.kind)
  assert.equal(saved.id,item.id)
  assert.equal(saved.status,item.status)
  assert.equal(saved.process,item.process)
  assert.equal(original.contractFile.name,'old.pdf')
  assert.throws(()=>changeResearch([original],101,'นักวิจัย',{type:'edit',fields:draft}),/ไม่มีสิทธิ์/)
})

test('invalid inline selections and attachments fail without modifying the project', () => {
  for(const patch of [{projectType:'invalid'},{fundingType:'invalid'},{contractFile:new File(['x'],'file.txt',{type:'text/plain'})},{contractFile:new File([],'empty.pdf',{type:'application/pdf'})}])assert.throws(()=>changeResearch([item],101,'ผู้ดูแลระบบ',{type:'edit',fields:{...item,...patch}}))
  assert.equal(item.title,'โครงการเดิม')
})
