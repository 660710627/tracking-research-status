import test from 'node:test'
import assert from 'node:assert/strict'
import {demoSubprocesses,validateProcessDetails} from '../src/processDemo.ts'
import {changeResearch} from '../src/research.ts'
const item={id:1,title:'โครงการ',contract:'A',lead:'หัวหน้า',unit:'คณะ',budget:100,status:'กำลังดำเนินการ',process:1,kind:'โครงการหลัก',endDate:'2570'}
test('demo subprocesses cover eight processes and restrict attachments per option',()=>{
  assert.equal(Object.keys(demoSubprocesses).length,8)
  for(const [process,options] of Object.entries(demoSubprocesses)){
    assert.ok(options.some(option=>option.attachment))
    assert.ok(options.some(option=>!option.attachment))
    for(const option of options){
      const details={subprocess:option.name,note:'',file:new File(['demo'],'test.txt')}
      if(option.attachment)assert.doesNotThrow(()=>validateProcessDetails(Number(process),details))
      else assert.throws(()=>validateProcessDetails(Number(process),details),/ไม่รองรับ/)
    }
  }
})
test('process details are stored together and invalid changes preserve original data',()=>{
  const details={subprocess:demoSubprocesses[2][0].name,note:' ตรวจแล้ว ',file:new File(['demo'],'test.txt')}
  const [updated]=changeResearch([item],1,'ผู้ประสานงาน',{type:'process',process:2,details})
  assert.equal(updated.processDetails.note,'ตรวจแล้ว')
  assert.equal(updated.processDetails.file,details.file)
  assert.equal(item.process,1)
  assert.throws(()=>changeResearch([item],1,'ผู้ประสานงาน',{type:'process',process:2,details:{...details,subprocess:'ไม่ตรง'}}),/เลือกกระบวนการย่อย/)
  assert.throws(()=>validateProcessDetails(2,{...details,file:new File([],'empty')}),/ไฟล์ว่าง/)
  assert.throws(()=>changeResearch([item],1,'นักวิจัย',{type:'process',process:2,details}),/ไม่มีสิทธิ์/)
})

test('process history keeps earlier subprocesses and attachments when moving forward and back',()=>{
  const first={subprocess:demoSubprocesses[2][0].name,note:'ครั้งแรก',file:new File(['demo'],'agreement.txt')}
  const second={subprocess:demoSubprocesses[3][1].name,note:'ถัดไป',file:null}
  const third={subprocess:demoSubprocesses[2][1].name,note:'กลับมาตรวจสอบ',file:null}
  let items=changeResearch([item],1,'ผู้ประสานงาน',{type:'process',process:2,details:first})
  const previousHistory=items[0].processHistory
  items=changeResearch(items,1,'ผู้ดูแลระบบ',{type:'process',process:3,details:second})
  items=changeResearch(items,1,'ผู้ประสานงาน',{type:'process',process:2,details:third})
  assert.equal(items[0].processHistory.length,3)
  assert.deepEqual(items[0].processHistory.map(entry=>entry.process),[2,3,2])
  assert.equal(items[0].processHistory[0].file,first.file)
  assert.equal(items[0].processHistory[0].note,'ครั้งแรก')
  assert.equal(items[0].processDetails.subprocess,third.subprocess)
  assert.equal(previousHistory.length,1)
  assert.ok(items[0].processHistory.every(entry=>Number.isFinite(Date.parse(entry.recordedAt))))
})

test('process log records actor, role, old and new process without rewriting older entries',()=>{
 const details={subprocess:demoSubprocesses[2][0].name,note:'ทดสอบ log',file:null}
 const first=changeResearch([item],1,'ผู้ดูแลระบบ',{type:'process',process:2,details},'ผู้ดูแลตัวอย่าง')
 const entry=first[0].processHistory[0]
 assert.equal(entry.changedBy,'ผู้ดูแลตัวอย่าง')
 assert.equal(entry.changedRole,'ผู้ดูแลระบบ')
 assert.equal(entry.fromProcess,1)
 assert.equal(entry.process,2)
 const second=changeResearch(first,1,'ผู้ประสานงาน',{type:'process',process:3,details:{subprocess:demoSubprocesses[3][1].name,note:'',file:null}},'ผู้ประสานตัวอย่าง')
 assert.equal(second[0].processHistory[0],entry)
 assert.equal(second[0].processHistory[1].changedBy,'ผู้ประสานตัวอย่าง')
 assert.equal(second[0].processHistory[1].fromProcess,2)
})
