import test from 'node:test'
import assert from 'node:assert/strict'
import {sortResearches,researchColumns} from '../src/researchSort.ts'
import {paginate} from '../src/pagination.ts'
const items=[
  {id:2,contract:'A-10',title:'B',lead:'B',fund:'B',budget:200,status:'B',process:8,endDate:'1 ม.ค. 2571'},
  {id:1,contract:'A-2',title:'A',lead:'A',fund:'A',budget:100,status:'A',process:3,endDate:'30 ธ.ค. 2570'},
]
test('every column sorts both directions without mutating input',()=>{
  for(const {key} of researchColumns){
    assert.deepEqual(sortResearches(items,{key,direction:'asc'}).map(x=>x.id),[1,2])
    assert.deepEqual(sortResearches(items,{key,direction:'desc'}).map(x=>x.id),[2,1])
  }
  assert.deepEqual(items.map(x=>x.id),[2,1])
  assert.deepEqual(sortResearches(items,null),items)
})
test('Thai dates sort chronologically across months and years',()=>{
  const dated=['1 ม.ค. 2571','30 ก.ย. 2570','2 ก.พ. 2570'].map((endDate,id)=>({...items[0],id,endDate}))
  assert.deepEqual(sortResearches(dated,{key:'endDate',direction:'asc'}).map(x=>x.id),[2,1,0])
})
test('supported page sizes apply after sorting the entire result set',()=>{
  const many=Array.from({length:105},(_,i)=>({...items[0],id:i,process:105-i}))
  const sorted=sortResearches(many,{key:'process',direction:'asc'})
  for(const size of [10,25,50,100]){
    const page=paginate(sorted,1,size)
    assert.equal(page.items.length,size)
    assert.equal(page.items[0].process,1)
    assert.equal(page.end,size)
    assert.equal(paginate(sorted,2,size).start,size+1)
  }
})
