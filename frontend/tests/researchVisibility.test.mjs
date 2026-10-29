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
test('researchers cannot edit abstracts even on their own projects',()=>{
 const items=[{id:1,leader:{email:'a@example.com'},title:'Original',thai:'old',english:'old',sdgs:[1]},{id:2,leader:{email:'b@example.com'}}]
 const original=structuredClone(items)
 for(const id of [1,2])assert.throws(()=>changeOwnAbstracts(items,id,'นักวิจัย','a@example.com',{thai:'ใหม่',english:'New'}),/เฉพาะ SDGs/)
 assert.deepEqual(items,original)
})