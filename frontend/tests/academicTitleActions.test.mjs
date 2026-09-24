import test from 'node:test'
import assert from 'node:assert/strict'
import { addAcademicTitle, updateAcademicTitle, deleteAcademicTitle } from '../src/academicTitlesDemo.ts'

const first = { id:'one', thaiAbbreviation:'ผศ.', thaiTitle:'ผู้ช่วยศาสตราจารย์', englishAbbreviation:'Asst. Prof.', englishTitle:'Assistant Professor' }
const second = { id:'two', thaiAbbreviation:'รศ.', thaiTitle:'รองศาสตราจารย์', englishAbbreviation:'Assoc. Prof.', englishTitle:'Associate Professor' }
const draft = { thaiAbbreviation:' ศ. ', thaiTitle:' ศาสตราจารย์ ', englishAbbreviation:' Prof. ', englishTitle:' Professor ' }

test('admin can add a complete title without mutating the original list', () => {
  const current = [first, second]
  const next = addAcademicTitle(current, 'three', draft, 'ผู้ดูแลระบบ')
  assert.equal(current.length, 2)
  assert.deepEqual(next[2], {id:'three', thaiAbbreviation:'ศ.', thaiTitle:'ศาสตราจารย์', englishAbbreviation:'Prof.', englishTitle:'Professor'})
})

test('edit and delete affect only the chosen title', () => {
  const current = [first, second]
  const edited = updateAcademicTitle(current, 'one', draft, 'ผู้ดูแลระบบ')
  assert.equal(edited[0].id, 'one')
  assert.equal(edited[0].thaiAbbreviation, 'ศ.')
  assert.deepEqual(edited[1], second)
  assert.deepEqual(current[0], first)
  assert.deepEqual(deleteAcademicTitle(current, 'one', 'ผู้ดูแลระบบ'), [second])
})

test('admin rights, target existence and required fields are checked', () => {
  assert.throws(() => addAcademicTitle([], 'three', draft, 'นักวิจัย'), /ไม่มีสิทธิ์/)
  assert.throws(() => updateAcademicTitle([first], 'one', draft, 'ผู้ประสานงาน'), /ไม่มีสิทธิ์/)
  assert.throws(() => deleteAcademicTitle([first], 'one', 'นักวิจัย'), /ไม่มีสิทธิ์/)
  assert.throws(() => addAcademicTitle([first], 'one', draft, 'ผู้ดูแลระบบ'), /ซ้ำ/)
  assert.throws(() => updateAcademicTitle([first], 'missing', draft, 'ผู้ดูแลระบบ'), /ไม่พบ/)
  assert.throws(() => deleteAcademicTitle([first], 'missing', 'ผู้ดูแลระบบ'), /ไม่พบ/)
  assert.throws(() => addAcademicTitle([], 'three', {...draft, englishTitle:' '}, 'ผู้ดูแลระบบ'), /ครบทุกช่อง/)
})
