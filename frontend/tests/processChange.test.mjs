import test from 'node:test'
import assert from 'node:assert/strict'
import { changeResearch } from '../src/research.ts'

test('process picker supports moving to every earlier and later process without changing project identity or status', () => {
  for (let current=1;current<=8;current++) {
    const item={id:101,title:'โครงการ',contract:'A',lead:'หัวหน้า',unit:'คณะ',budget:100,status:'กำลังดำเนินการ',process:current,kind:'โครงการหลัก',endDate:'30 ก.ย. 2570'}
    for (let target=1;target<=8;target++) {
      if(target===current)continue
      const [updated]=changeResearch([item],101,'ผู้ประสานงาน',{type:'process',process:target})
      assert.deepEqual(updated,{...item,process:target})
      assert.equal(item.process,current)
    }
  }
})
