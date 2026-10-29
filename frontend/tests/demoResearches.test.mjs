import test from 'node:test'
import assert from 'node:assert/strict'
import { seed } from '../src/demoResearches.ts'
test('demo contains eleven projects with distinct positive IDs and contracts',()=>{
  assert.equal(seed.length,11)
  assert.equal(new Set(seed.map(item=>item.id)).size,11)
  assert.equal(new Set(seed.map(item=>item.contract)).size,11)
  assert.ok(seed.every(item=>Number.isInteger(item.id)&&item.id>0&&item.title&&item.lead))
})
