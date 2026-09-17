// In-memory demo behavior only. Production authorization belongs to the API/database.
export type Role = 'นักวิจัย' | 'ผู้ประสานงาน' | 'ผู้ดูแลระบบ'
export const statuses = ['กำลังดำเนินการ', 'กำลังดำเนินการ(ขยายเวลาครั้งที่ 1)', 'กำลังดำเนินการ(ขยายเวลาครั้งที่ 2)', 'กำลังดำเนินการ(ขยายเวลามากกว่า 2 ครั้ง)', 'โครงการเสร็จสิ้น', 'ยุติโครงการ'] as const
export type Research = { id:number; title:string; contract:string; lead:string; unit:string; budget:number; status:string; process:number; kind:string; endDate:string }
export type ResearchFields = Pick<Research, 'title' | 'contract' | 'lead' | 'unit' | 'budget' | 'endDate'>
export type ResearchAction = {type:'status'; status:string} | {type:'process'; process:number} | {type:'edit'; fields:ResearchFields} | {type:'delete'}
export const canManageResearch = (role:Role) => role === 'ผู้ประสานงาน' || role === 'ผู้ดูแลระบบ'
export const isTerminal = (status:string) => status === statuses[4] || status === statuses[5]
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
  }
  return items.map(item => item.id === id ? updated : item)
}
