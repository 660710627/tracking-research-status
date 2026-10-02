import test from 'node:test'
import assert from 'node:assert/strict'
import {visibleResearches} from '../src/researchVisibility.ts'
test('researcher access uses membership email, never display name',()=>{
 const items=[{id:1,leader:{email:'owner@example.com'}},{id:2,coResearchers:[{email:'OWNER@example.com'}]},{id:3,lead:'same name'},{id:4,leader:{email:'other@example.com'}}]
 assert.deepEqual(visibleResearches(items,'นักวิจัย',' owner@example.com ').map(x=>x.id),[1,2])
 assert.deepEqual(visibleResearches(items,'นักวิจัย',''),[])
 assert.deepEqual(visibleResearches(items,'นักวิจัย','unknown@example.com'),[])
 assert.equal(visibleResearches(items,'ผู้ประสานงาน',''),items)
 assert.equal(visibleResearches(items,'ผู้ดูแลระบบ',''),items)
})

import {changeOwnAbstracts} from '../src/researchVisibility.ts'
test('abstract edits require ownership and reject every other field',()=>{
 const items=[{id:1,leader:{email:'a@example.com'},title:'Original',thai:'old',english:'old',sdgs:[1]},{id:2,leader:{email:'b@example.com'}}]
 const result=changeOwnAbstracts(items,1,'นักวิจัย','a@example.com',{thai:'ใหม่',english:'New'})
 assert.equal(result[0].thai,'ใหม่')
 assert.equal(result[0].english,'New')
 assert.equal(result[0].title,'Original')
 assert.deepEqual(result[0].sdgs,[1])
 assert.equal(items[0].thai,'old')
 assert.equal(result[1],items[1])
 assert.throws(()=>changeOwnAbstracts(items,2,'นักวิจัย','a@example.com',{thai:'x',english:'x'}))
 assert.throws(()=>changeOwnAbstracts(items,1,'นักวิจัย','a@example.com',{thai:'x',english:'x',title:'tampered'}))
})
