import test from 'node:test'
import assert from 'node:assert/strict'
import { decideAccount, demoAccounts } from '../src/approvalDemo.ts'

test('demo accounts include a first login date for each pending account', () => {
  assert.ok(demoAccounts.every(account => account.firstLoginAt))
  assert.ok(demoAccounts.every(account => account.status === 'PENDING'))
})

test('approval and rejection change only the selected pending account', () => {
  for (const decision of ['APPROVED', 'REJECTED']) {
    const next = decideAccount(demoAccounts, demoAccounts[0].email, 'ผู้ดูแลระบบ', decision)
    assert.equal(next[0].status, decision)
    assert.equal(next[1].status, 'PENDING')
    assert.equal(demoAccounts[0].status, 'PENDING')
  }
})

test('only an admin can decide, and an account can be decided once', () => {
  const email = demoAccounts[0].email
  assert.throws(() => decideAccount(demoAccounts, email, 'นักวิจัย', 'APPROVED'), /ไม่มีสิทธิ์/)
  assert.throws(() => decideAccount(demoAccounts, 'missing@example.org', 'ผู้ดูแลระบบ', 'APPROVED'), /ไม่พบ/)
  const approved = decideAccount(demoAccounts, email, 'ผู้ดูแลระบบ', 'APPROVED')
  assert.throws(() => decideAccount(approved, email, 'ผู้ดูแลระบบ', 'REJECTED'), /รออนุมัติ/)
})
