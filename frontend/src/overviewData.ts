import { isTerminal, statuses } from './research.ts'
import type { Research } from './research'
export type OverviewGroup='status'|'kind'|'budget'
export function overviewData(items:Research[],group:OverviewGroup) {
  const labels=group==='status'?[...statuses]:[...new Set(items.map(item=>(group==='kind'?item.kind:item.unit)||'ไม่ระบุ'))]
  const segments=labels.map(label=>{const records=items.filter(item=>(group==='status'?item.status:group==='kind'?item.kind:item.unit||'ไม่ระบุ')===label);return {label,count:records.length,value:group==='budget'?records.reduce((sum,item)=>sum+item.budget,0):records.length}})
  if(group==='budget')segments.sort((a,b)=>b.value-a.value)
  return {total:items.length,budget:items.reduce((sum,item)=>sum+item.budget,0),active:items.filter(item=>!isTerminal(item.status)).length,completed:items.filter(item=>item.status==='โครงการเสร็จสิ้น').length,segments}
}
