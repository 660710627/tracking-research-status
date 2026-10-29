import { canManageResearch } from './research.ts'
import type { Research, Role } from './research'

// Demo only: keep the original record so restoring never allocates a new ID.
export type TrashedResearch = { research:Research; deletedAt:string; deletedBy:Role }
export type ResearchStore = { items:Research[]; trash:TrashedResearch[] }
export function trashResearch(store:ResearchStore,id:number,role:Role,deletedAt:string):ResearchStore {
  if(!canManageResearch(role))throw new Error('คุณไม่มีสิทธิ์จัดการถังขยะ')
  const research=store.items.find(item=>item.id===id)
  if(!research)throw new Error('ไม่พบงานวิจัยที่เลือก')
  if(store.trash.some(entry=>entry.research.id===id))throw new Error('งานวิจัยนี้อยู่ในถังขยะแล้ว')
  return {items:store.items.filter(item=>item.id!==id),trash:[{research,deletedAt,deletedBy:role},...store.trash]}
}
export function restoreResearch(store:ResearchStore,id:number,role:Role):ResearchStore {
  if(!canManageResearch(role))throw new Error('คุณไม่มีสิทธิ์จัดการถังขยะ')
  const entry=store.trash.find(item=>item.research.id===id)
  if(!entry)throw new Error('ไม่พบงานวิจัยในถังขยะ')
  const research=entry.research
  if(store.items.some(item=>item.id===id))throw new Error('มีงานวิจัย ID นี้อยู่ในรายการแล้ว')
  if(store.items.some(item=>item.contract.trim().toLowerCase()===research.contract.trim().toLowerCase()))throw new Error('กู้คืนไม่ได้: เลขที่สัญญาทุนซ้ำกับโครงการในรายการ กรุณาแก้ไขโครงการที่ซ้ำก่อน')
  if(research.kind!=='โครงการต่อเนื่อง'&&store.items.some(item=>item.title.trim()===research.title.trim()))throw new Error('กู้คืนไม่ได้: ชื่องานวิจัยซ้ำกับโครงการในรายการ กรุณาแก้ไขโครงการที่ซ้ำก่อน')
  return {items:[research,...store.items],trash:store.trash.filter(item=>item.research.id!==id)}
}
