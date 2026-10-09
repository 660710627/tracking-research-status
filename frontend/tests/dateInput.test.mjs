import test from 'node:test'
import assert from 'node:assert/strict'
import {toCalendarDate,fromCalendarDate} from '../src/dateInput.ts'
test('calendar converts existing Buddhist dates without shifting day or year',()=>{
 assert.equal(toCalendarDate('30 ก.ย. 2571'),'2028-09-30')
 assert.equal(toCalendarDate('01/10/2569'),'2026-10-01')
 assert.equal(fromCalendarDate('2028-02-29'),'29 ก.พ. 2571')
 assert.equal(toCalendarDate(fromCalendarDate('2028-02-29')),'2028-02-29')
 assert.equal(toCalendarDate(''),'')
 assert.equal(fromCalendarDate(''),'')
})
