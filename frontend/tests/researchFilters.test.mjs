import test from 'node:test'
import assert from 'node:assert/strict'
import {filterResearches,emptyResearchFilters,responsibleSuggestions} from '../src/researchFilters.ts'

const items=[
  {id:1,title:'Water study',contract:'A-01',unit:'Science',lead:'Alice',status:'active'},
  {id:2,title:'Art study',contract:'B-02',unit:'Arts',lead:'Bob',status:'completed'},
]
test('responsible suggestions match partial Thai names, deduplicate, and handle empty and unknown text',()=>{
  const names=['รศ. ดร. กานดา วัฒนศิลป์','ดร. พิมพ์ชนก ศรีสุข','รศ. ดร. กานดา วัฒนศิลป์']
  assert.deepEqual(responsibleSuggestions(names,' กาน '),[names[0]])
  assert.deepEqual(responsibleSuggestions(names,'ศรีสุข'),[names[1]])
  assert.deepEqual(responsibleSuggestions(names,'ไม่มี'),[])
  assert.deepEqual(responsibleSuggestions(names,'  '),[])
  assert.deepEqual(responsibleSuggestions([names[1]],'กาน'),[])
})
test('responsible filter accepts partial names and whitespace, case insensitively',()=>{
  assert.deepEqual(filterResearches(items,{...emptyResearchFilters,lead:' ALI '}),[items[0]])
  assert.deepEqual(filterResearches(items,{...emptyResearchFilters,lead:'  '}),items)
})
test('empty filters preserve all supplied accessible records',()=>{
  assert.deepEqual(filterResearches(items,emptyResearchFilters),items)
  assert.deepEqual(filterResearches([items[0]],emptyResearchFilters),[items[0]])
})
test('search trims whitespace and matches title or contract ignoring case',()=>{
  for(const query of [' WATER ','a-01']) assert.deepEqual(filterResearches(items,{...emptyResearchFilters,query}),[items[0]])
})
test('unit, responsible researcher and status combine with search',()=>{
  const filters={query:'study',unit:'Arts',lead:'Bob',status:'completed'}
  assert.deepEqual(filterResearches(items,filters),[items[1]])
  assert.deepEqual(filterResearches(items,{...filters,lead:'Alice'}),[])
  assert.deepEqual(filterResearches(items,{...filters,unit:'Science'}),[])
  assert.deepEqual(filterResearches(items,{...filters,status:'active'}),[])
})
