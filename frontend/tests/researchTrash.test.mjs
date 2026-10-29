import test from 'node:test'
import assert from 'node:assert/strict'
import { trashResearch, restoreResearch } from '../src/researchTrash.ts'
const project={id:101,title:'งานวิจัย',contract:'A-01',kind:'โครงการหลัก',status:'ยุติโครงการ',process:4,sdgs:[6],keywords:['น้ำ'],leader:{email:'demo@example.com'},contractFile:{name:'grant.pdf'}}
test('managers move projects into trash and restore all data with the original identity',()=>{
  for(const role of ['ผู้ดูแลระบบ','ผู้ประสานงาน']){
    const original={items:[project],trash:[]}
    const deleted=trashResearch(original,101,role,'2026-10-16T00:00:00Z')
    assert.equal(deleted.items.length,0)
    assert.equal(deleted.trash[0].research,project)
    assert.equal(deleted.trash[0].deletedBy,role)
    const restored=restoreResearch(deleted,101,role)
    assert.equal(restored.items[0],project)
    assert.equal(restored.trash.length,0)
    assert.equal(original.items.length,1)
    assert.throws(()=>restoreResearch(restored,101,role),/ไม่พบ/)
  }
})
test('researchers cannot delete or restore and unknown IDs are rejected',()=>{
  const original={items:[project],trash:[]}
  const deleted=trashResearch(original,101,'ผู้ดูแลระบบ','2026-10-16T00:00:00Z')
  assert.throws(()=>trashResearch(original,101,'นักวิจัย',''),/สิทธิ์/)
  assert.throws(()=>restoreResearch(deleted,101,'นักวิจัย'),/สิทธิ์/)
  assert.throws(()=>trashResearch(original,999,'ผู้ดูแลระบบ',''),/ไม่พบ/)
})
test('restore collisions preserve the bin record and continuation titles follow existing rules',()=>{
  const deleted=trashResearch({items:[project],trash:[]},101,'ผู้ดูแลระบบ','2026-10-16T00:00:00Z')
  for(const other of [{...project},{...project,id:102,title:'อื่น'},{...project,id:102,contract:'B'}])assert.throws(()=>restoreResearch({...deleted,items:[other]},101,'ผู้ดูแลระบบ'),/ซ้ำ|ID/)
  assert.equal(deleted.trash.length,1)
  const continuation={...deleted,items:[{...project,id:102,contract:'B'}],trash:[{...deleted.trash[0],research:{...project,kind:'โครงการต่อเนื่อง'}}]}
  assert.equal(restoreResearch(continuation,101,'ผู้ดูแลระบบ').items.length,2)
})
