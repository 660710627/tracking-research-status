import test from 'node:test'
import assert from 'node:assert/strict'
import { overviewData } from '../src/overviewData.ts'
const items=[{status:'กำลังดำเนินการ',kind:'โครงการหลัก',unit:'คณะ ก',budget:200},{status:'โครงการเสร็จสิ้น',kind:'โครงการต่อเนื่อง',unit:'คณะ ข',budget:100},{status:'ยุติโครงการ',kind:'โครงการหลัก',unit:'คณะ ก',budget:50}]
test('each dashboard view reconciles counts and budget while preserving source records',()=>{
  const before=JSON.stringify(items)
  for(const group of ['status','kind','budget']){
    const result=overviewData(items,group)
    assert.equal(result.total,3);assert.equal(result.active,1);assert.equal(result.completed,1);assert.equal(result.budget,350)
    assert.equal(result.segments.reduce((sum,item)=>sum+item.count,0),3)
    assert.equal(result.segments.reduce((sum,item)=>sum+item.value,0),group==='budget'?350:3)
  }
  assert.equal(JSON.stringify(items),before)
  assert.deepEqual(overviewData(items,'budget').segments.map(item=>item.value),[250,100])
})
test('dashboard supports empty data and missing units',()=>{
  assert.equal(overviewData([],'status').segments.length,6)
  assert.equal(overviewData([],'kind').total,0)
  assert.deepEqual(overviewData([],'budget').segments,[])
  assert.equal(overviewData([{...items[0],unit:''}],'budget').segments[0].label,'ไม่ระบุ')
})
