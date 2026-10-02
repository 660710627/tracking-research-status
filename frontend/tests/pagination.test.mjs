import test from 'node:test'
import assert from 'node:assert/strict'
import {paginate} from '../src/pagination.ts'

test('pagination covers first, middle and final pages without skipping records',()=>{
  const items=Array.from({length:24},(_,i)=>i+1)
  const pages=[1,2,3].map(page=>paginate(items,page))
  assert.deepEqual(pages.flatMap(page=>page.items),items)
  assert.deepEqual(pages.map(({start,end})=>[start,end]),[[1,10],[11,20],[21,24]])
  assert.equal(pages[2].pageCount,3)
})
test('pagination clamps after filtering or deletion and handles empty results',()=>{
  assert.deepEqual(paginate([1,2,3,4],3),{items:[1,2,3,4],page:1,pageCount:1,start:1,end:4,total:4})
  assert.deepEqual(paginate([],1),{items:[],page:1,pageCount:1,start:0,end:0,total:0})
  assert.equal(paginate([1],0).page,1)
})
