import test from 'node:test'
import assert from 'node:assert/strict'
import { addAccount, approvalAccounts, decideAccount, demoAccounts, managedAccounts, managedDemoAccounts } from '../src/approvalDemo.ts'

test('user management contains approved and suspended accounts, never pending or rejected accounts', () => {
  const pending=demoAccounts[0]
  const rejected={...demoAccounts[1],status:'REJECTED'}
  const all=[pending,rejected,...managedDemoAccounts]
  assert.deepEqual(managedAccounts(all).map(account=>account.status),['APPROVED','SUSPENDED'])
  const approved=decideAccount(all,pending.email,'ผู้ดูแลระบบ','APPROVED')
  assert.ok(managedAccounts(approved).some(account=>account.id===pending.id))
})

test('approval module keeps first-login requests separate from manually managed accounts', () => {
  const all=[...demoAccounts,...managedDemoAccounts]
  assert.deepEqual(approvalAccounts(all),demoAccounts)
})

test('new managed accounts cannot be created in pending or rejected status', () => {
  const draft={email:'new@example.org',prefix:'นาย',firstName:'ทดสอบ',lastName:'ระบบ',title:'',unit:'สำนักงานบริหารการวิจัย',role:'นักวิจัย',status:'PENDING',notifyResearch:false}
  assert.throws(()=>addAccount(managedDemoAccounts,'new',draft,'ผู้ดูแลระบบ'),/สถานะ/)
  assert.throws(()=>addAccount(managedDemoAccounts,'new',{...draft,status:'REJECTED'},'ผู้ดูแลระบบ'),/สถานะ/)
})
