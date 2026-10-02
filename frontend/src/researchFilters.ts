import type { Research } from './research'

export type ResearchFilters = { query:string; unit:string; lead:string; status:string }
export const emptyResearchFilters:ResearchFilters = {query:'',unit:'',lead:'',status:''}
export function matchesResponsibleName(name:string, query:string) {
  const normalize=(value:string)=>value.normalize('NFC').toLocaleLowerCase().replace(/\s+/g,' ').trim()
  return normalize(name).includes(normalize(query))
}
export function responsibleSuggestions(names:string[], query:string) {
  if(!query.trim()) return []
  return [...new Set(names)].filter(name=>name.trim() && matchesResponsibleName(name,query)).sort((a,b)=>a.localeCompare(b,'th'))
}
export function filterResearches(items:Research[], filters:ResearchFilters) {
  const query=filters.query.trim().toLocaleLowerCase()
  return items.filter(item =>
    (item.title.toLocaleLowerCase().includes(query) || item.contract.toLocaleLowerCase().includes(query)) &&
    (!filters.unit || item.unit===filters.unit) &&
    matchesResponsibleName(item.lead,filters.lead) &&
    (!filters.status || item.status===filters.status))
}
