import test from 'node:test'
import assert from 'node:assert/strict'
import { addAccount, deleteAccount, demoAccounts, updateAccount } from '../src/approvalDemo.ts'

const draft = {
  email: ' New.User@example.org ',
  prefix: 'นางสาว',
  firstName: 'มาลี',
  lastName: 'ใจดี',
  title: 'ดร.',
  unit: 'คณะวิทยาศาสตร์',
  role: 'นักวิจัย',
  status: 'APPROVED',
  notifyResearch: true,
}

test('admin adds an account from complete details without changing existing accounts', () => {
  const added = addAccount(demoAccounts, 'new-id', draft, 'ผู้ดูแลระบบ')
  assert.equal(added.length, demoAccounts.length + 1)
  assert.deepEqual(added.slice(0, demoAccounts.length), demoAccounts)
  assert.equal(added.at(-1).email, 'new.user@example.org')
  assert.equal(added.at(-1).name, 'มาลี ใจดี')
  assert.equal(added.at(-1).status, 'APPROVED')
  assert.equal(added.at(-1).notifyResearch, true)
  assert.equal(added.at(-1).firstLoginAt, '—')
})

test('admin edits email, name, role and status while preserving account identity and login date', () => {
  const current = demoAccounts[0]
  const edited = updateAccount(demoAccounts, current.id, {...draft, status:'SUSPENDED'}, 'ผู้ดูแลระบบ')
  assert.equal(edited[0].id, current.id)
  assert.equal(edited[0].firstLoginAt, current.firstLoginAt)
  assert.equal(edited[0].name, 'มาลี ใจดี')
  assert.equal(edited[0].status, 'SUSPENDED')
  assert.deepEqual(edited[1], demoAccounts[1])
  assert.deepEqual(demoAccounts[0], current)
})

test('delete removes only the selected account', () => {
  const removed = deleteAccount(demoAccounts, demoAccounts[0].id, 'ผู้ดูแลระบบ')
  assert.deepEqual(removed, [demoAccounts[1]])
})

test('account actions enforce admin rights, target existence, email uniqueness and required details', () => {
  assert.throws(() => addAccount(demoAccounts, 'new', draft, 'นักวิจัย'), /ไม่มีสิทธิ์/)
  assert.throws(() => updateAccount(demoAccounts, demoAccounts[0].id, draft, 'ผู้ประสานงาน'), /ไม่มีสิทธิ์/)
  assert.throws(() => deleteAccount(demoAccounts, demoAccounts[0].id, 'นักวิจัย'), /ไม่มีสิทธิ์/)
  assert.throws(() => addAccount(demoAccounts, demoAccounts[0].id, draft, 'ผู้ดูแลระบบ'), /ซ้ำ/)
  assert.throws(() => addAccount(demoAccounts, 'new', {...draft,email:demoAccounts[0].email.toUpperCase()}, 'ผู้ดูแลระบบ'), /อีเมลนี้/)
  assert.throws(() => addAccount(demoAccounts, 'new', {...draft,email:'invalid'}, 'ผู้ดูแลระบบ'), /อีเมล/)
  assert.throws(() => addAccount(demoAccounts, 'new', {...draft,firstName:' '}, 'ผู้ดูแลระบบ'), /กรอก/)
  assert.throws(() => updateAccount(demoAccounts, 'missing', draft, 'ผู้ดูแลระบบ'), /ไม่พบ/)
  assert.throws(() => deleteAccount(demoAccounts, 'missing', 'ผู้ดูแลระบบ'), /ไม่พบ/)
})
