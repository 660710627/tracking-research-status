import type { Research } from './research'
import { toCalendarDate } from './dateInput.ts'

export type ResearchSortKey='contract'|'title'|'fund'|'status'|'process'|'endDate'
export type ResearchSort={key:ResearchSortKey;direction:'asc'|'desc'}
export const researchColumns:{key:ResearchSortKey;label:string;description:string}[]=[
  {key:'contract',label:'เลขสัญญา',description:'เลขสัญญา'},
  {key:'title',label:'โครงการ / ผู้รับผิดชอบ',description:'ชื่อโครงการ ตามด้วยชื่อผู้รับผิดชอบ'},
  {key:'fund',label:'ทุนวิจัย',description:'ชื่อทุน ตามด้วยจำนวนเงิน'},
  {key:'status',label:'สถานะ',description:'ชื่อสถานะ'},
  {key:'process',label:'กระบวนการปัจจุบัน',description:'ลำดับกระบวนการ'},
  {key:'endDate',label:'สิ้นสุด',description:'วันที่สิ้นสุด'},
]
const collator=new Intl.Collator('th',{numeric:true,sensitivity:'base'})
export function sortResearches(items:Research[],sort:ResearchSort|null) {
  if(!sort) return items
  const {key,direction}=sort
  return [...items].sort((a,b)=>{
    let comparison:number
    if(key==='process') comparison=a.process-b.process
    else if(key==='endDate') comparison=collator.compare(toCalendarDate(a.endDate),toCalendarDate(b.endDate))
    else comparison=collator.compare(a[key]??'',b[key]??'')
    if(!comparison && key==='title') comparison=collator.compare(a.lead,b.lead)
    if(!comparison && key==='fund') comparison=a.budget-b.budget
    return (direction==='asc'?comparison:-comparison) || a.id-b.id
  })
}
