import test from 'node:test'
import assert from 'node:assert/strict'
import { academicTitlesDemo } from '../src/academicTitlesDemo.ts'

test('academic title demo has the four requested columns for all seven rows', () => {
  assert.equal(academicTitlesDemo.length, 7)
  for (const row of academicTitlesDemo) {
    assert.deepEqual(Object.keys(row), ['thaiAbbreviation', 'englishAbbreviation', 'thaiTitle', 'englishTitle'])
    assert.ok(Object.values(row).every(value => typeof value === 'string' && value.trim().length > 0))
  }
})

test('doctor and non-doctor examples retain their distinct abbreviations', () => {
  assert.equal(academicTitlesDemo[1].thaiAbbreviation, 'ผศ.ดร.')
  assert.equal(academicTitlesDemo[4].thaiAbbreviation, 'ผศ.')
  assert.equal(academicTitlesDemo[1].englishAbbreviation, 'Asst. Prof. Dr.')
  assert.equal(academicTitlesDemo[4].englishAbbreviation, 'Asst. Prof.')
})
