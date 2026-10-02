import type { Research, Role } from './research'

export const demoResearcher = {name:'กานดา วัฒนศิลป์',email:'kanda@example.com'}
export function visibleResearches(items:Research[],role:Role,email:string):Research[] {
  if(role!=='นักวิจัย')return items
  const identity=email.trim().toLowerCase()
  if(!identity)return []
  return items.filter(item=>[item.leader,...(item.coResearchers??[])].some(member=>member?.email.trim().toLowerCase()===identity))
}

export function changeOwnAbstracts(items:Research[],id:number,role:Role,email:string,patch:{thai:string;english:string}):Research[] {
  if(role!=='นักวิจัย'||!visibleResearches(items,role,email).some(item=>item.id===id))throw new Error('ไม่มีสิทธิ์แก้ไขบทคัดย่อของงานวิจัยนี้')
  if(Object.keys(patch).some(key=>key!=='thai'&&key!=='english')||typeof patch.thai!=='string'||typeof patch.english!=='string')throw new Error('แก้ไขได้เฉพาะบทคัดย่อไทยและอังกฤษ')
  return items.map(item=>item.id===id?{...item,thai:patch.thai.trim(),english:patch.english.trim()}:item)
}
