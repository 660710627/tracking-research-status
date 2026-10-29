import type { Research, Role } from './research'

export const demoResearcher = {name:'กานดา วัฒนศิลป์',email:'kanda@example.com'}
export function visibleResearches(items:Research[],role:Role,email:string):Research[] {
  if(role!=='นักวิจัย')return items
  const identity=email.trim().toLowerCase()
  if(!identity)return []
  return items.filter(item=>[item.leader,...(item.coResearchers??[])].some(member=>member?.email.trim().toLowerCase()===identity))
}

// Retained to reject legacy demo callers; researchers may edit SDGs only.
export function changeOwnAbstracts(..._request:unknown[]):never {
  void _request
  throw new Error('ไม่มีสิทธิ์แก้ไขบทคัดย่อ นักวิจัยแก้ไขได้เฉพาะ SDGs')
}