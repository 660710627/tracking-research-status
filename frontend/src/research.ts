// In-memory demo behavior only. Production authorization belongs to the API/database.
import { normalizeSdgs } from './sdgs.ts'
export type Role = 'นักวิจัย' | 'ผู้ประสานงาน' | 'ผู้ดูแลระบบ'
export const statuses = ['กำลังดำเนินการ', 'กำลังดำเนินการ(ขยายเวลาครั้งที่ 1)', 'กำลังดำเนินการ(ขยายเวลาครั้งที่ 2)', 'กำลังดำเนินการ(ขยายเวลามากกว่า 2 ครั้ง)', 'โครงการเสร็จสิ้น', 'ยุติโครงการ'] as const
export type ResearchDetails = { projectType:string; isSubsidized:boolean; budgetUnit:string; startDate:string; collaborators:string; fundingType:string; fund:string; thai:string; english:string; objectives:string[]; keywords:string[]; contractFile:File|null }
export type Research = { id:number; title:string; contract:string; lead:string; unit:string; budget:number; status:string; process:number; kind:string; endDate:string; sdgs?:number[] } & Partial<ResearchDetails>
export type ResearchFields = Pick<Research, 'title' | 'contract' | 'lead' | 'unit' | 'budget' | 'endDate'> & Partial<ResearchDetails>
export type ResearchDraft = ResearchFields & ResearchDetails
export function researchDraft(research:Research):ResearchDraft {
  return {title:research.title,contract:research.contract,lead:research.lead,unit:research.unit,budget:research.budget,endDate:research.endDate,
    projectType:research.projectType??'งานวิจัย',isSubsidized:research.isSubsidized??true,budgetUnit:research.budgetUnit??research.unit,startDate:research.startDate??'',collaborators:research.collaborators??'',fundingType:research.fundingType??'ภายใน',fund:research.fund??'',thai:research.thai??'',english:research.english??'',objectives:[...(research.objectives??[])],keywords:[...(research.keywords??[])],contractFile:research.contractFile??null}
}
export type ResearchAction = {type:'status'; status:string} | {type:'process'; process:number} | {type:'edit'; fields:ResearchFields} | {type:'delete'}
export const canManageResearch = (role:Role) => role === 'ผู้ประสานงาน' || role === 'ผู้ดูแลระบบ'
export const isTerminal = (status:string) => status === statuses[4] || status === statuses[5]
export function changeResearchSdgs(items:Research[],id:number,role:Role,ids:number[]):Research[] {
  if(!['นักวิจัย','ผู้ประสานงาน','ผู้ดูแลระบบ'].includes(role))throw new Error('คุณไม่มีสิทธิ์แก้ไข SDGs')
  const research=items.find(item=>item.id===id)
  if(!research)throw new Error('ไม่พบงานวิจัยที่เลือก')
  if(research.status==='ยุติโครงการ')throw new Error('งานวิจัยยุติแล้ว ไม่สามารถแก้ไข SDGs ได้')
  const sdgs=normalizeSdgs(ids)
  return items.map(item=>item.id===id?{...item,sdgs}:item)
}
export function nextStatuses(status:string) {
  const index = statuses.findIndex(value => value === status)
  return index < 0 || isTerminal(status) ? [] : statuses.slice(index + 1)
}
export function changeResearch(items:Research[], id:number, role:Role, action:ResearchAction):Research[] {
  if (!canManageResearch(role)) throw new Error('คุณไม่มีสิทธิ์จัดการงานวิจัย')
  const current = items.find(item => item.id === id)
  if (!current) throw new Error('ไม่พบงานวิจัยที่เลือก กรุณากลับไปยังรายการ')
  if (action.type === 'delete') return items.filter(item => item.id !== id)
  let updated = current
  if (action.type === 'status') {
    if (!nextStatuses(current.status).some(status => status === action.status)) throw new Error('ไม่สามารถย้อนสถานะหรือเปลี่ยนสถานะโครงการที่สิ้นสุดแล้ว')
    updated = {...current, status:action.status}
  } else if (action.type === 'process') {
    if (isTerminal(current.status) || action.process !== current.process + 1 || action.process > 8) throw new Error('ไม่สามารถปรับกระบวนการนี้ได้')
    updated = {...current, process:action.process}
  } else {
    const {title, contract, lead, unit, budget, endDate} = action.fields
    if (![title, contract, lead, unit, endDate].every(value => value.trim()) || !Number.isFinite(budget) || budget <= 0) throw new Error('กรอกข้อมูลให้ครบและระบุงบประมาณมากกว่า 0 บาท')
    if (items.some(item => item.id !== id && item.contract.trim().toLowerCase() === contract.trim().toLowerCase())) throw new Error('เลขที่สัญญาทุนนี้มีอยู่แล้ว')
    if (current.kind !== 'โครงการต่อเนื่อง' && items.some(item => item.id !== id && item.title.trim() === title.trim())) throw new Error('ชื่องานวิจัยนี้มีอยู่แล้ว')
    updated = {...current, title:title.trim(), contract:contract.trim(), lead:lead.trim(), unit:unit.trim(), budget, endDate:endDate.trim()}
    // Whitelist editable details: identity, continuation kind and progress never enter this patch.
    for (const key of ['budgetUnit','startDate','collaborators','fund','thai','english'] as const) {
      const value=action.fields[key]
      if(value!==undefined) updated[key]=value.trim()
    }
    for (const key of ['objectives','keywords'] as const) {
      const value=action.fields[key]
      if(value!==undefined) updated[key]=value.map(text=>text.trim()).filter(Boolean)
    }
    if(action.fields.projectType!==undefined){
      if(!['งานวิจัย','บริการวิชาการ'].includes(action.fields.projectType))throw new Error('เลือกประเภทโครงการให้ถูกต้อง')
      updated.projectType=action.fields.projectType
    }
    if(action.fields.fundingType!==undefined){
      if(!['ภายใน','ภายนอก'].includes(action.fields.fundingType))throw new Error('เลือกประเภทแหล่งทุนให้ถูกต้อง')
      updated.fundingType=action.fields.fundingType
    }
    if(action.fields.isSubsidized!==undefined)updated.isSubsidized=action.fields.isSubsidized
    if(action.fields.contractFile!==undefined){
      const file=action.fields.contractFile
      if(file && (!file.name.toLowerCase().endsWith('.pdf') || (file.type && file.type!=='application/pdf') || file.size>20*1024*1024 || file.size===0))throw new Error('เลือกไฟล์ PDF ขนาดไม่เกิน 20 MB ที่ไม่ใช่ไฟล์ว่าง')
      updated.contractFile=file
    }
  }
  return items.map(item => item.id === id ? updated : item)
}
