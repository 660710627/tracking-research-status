import type { Role } from './research'

export type AccountDecision = 'APPROVED' | 'REJECTED'
export type AccountStatus = 'PENDING' | AccountDecision | 'SUSPENDED'
export type AccountDraft = {
  email: string
  prefix: string
  firstName: string
  lastName: string
  title: string
  unit: string
  role: Role
  status: AccountStatus
  notifyResearch: boolean
}
export type DemoAccount = {
  id: string
  name: string
  firstLoginAt: string
} & AccountDraft

export const demoAccounts: DemoAccount[] = [
  { id:'demo-researcher', name: 'อภิชาติ ตั้งมั่น', prefix:'นาย', firstName:'อภิชาติ', lastName:'ตั้งมั่น', title: 'ดร.', unit: 'คณะวิทยาศาสตร์', email: 'researcher@example.org', role: 'นักวิจัย', firstLoginAt: '4 ต.ค. 2568 09:42', status: 'PENDING', notifyResearch:false },
  { id:'demo-coordinator', name: 'ณัฐชา พูนผล', prefix:'นางสาว', firstName:'ณัฐชา', lastName:'พูนผล', title: 'เจ้าหน้าที่', unit: 'สำนักงานบริหารการวิจัย', email: 'coordinator@example.org', role: 'ผู้ประสานงาน', firstLoginAt: '4 ต.ค. 2568 10:15', status: 'PENDING', notifyResearch:false },
]

export const managedDemoAccounts: DemoAccount[] = [
  { id:'demo-admin', name:'ศศิธร แก้วใส', prefix:'นางสาว', firstName:'ศศิธร', lastName:'แก้วใส', title:'เจ้าหน้าที่', unit:'สำนักงานบริหารการวิจัย', email:'admin@example.org', role:'ผู้ดูแลระบบ', firstLoginAt:'—', status:'APPROVED', notifyResearch:false },
  { id:'demo-suspended', name:'ปราณี ใจดี', prefix:'นางสาว', firstName:'ปราณี', lastName:'ใจดี', title:'ดร.', unit:'คณะวิทยาศาสตร์', email:'pranee@example.org', role:'นักวิจัย', firstLoginAt:'—', status:'SUSPENDED', notifyResearch:false },
]

export function managedAccounts(accounts:DemoAccount[]):DemoAccount[] {
  return accounts.filter(account=>account.status==='APPROVED'||account.status==='SUSPENDED')
}

export function approvalAccounts(accounts:DemoAccount[]):DemoAccount[] {
  return accounts.filter(account=>account.firstLoginAt!=='—')
}

function requireAdmin(role:Role) {
  if (role !== 'ผู้ดูแลระบบ') throw new Error('ไม่มีสิทธิ์จัดการบัญชีผู้ใช้งาน')
}

function normalizeAccountDraft(draft:AccountDraft):AccountDraft {
  const normalized={...draft,email:draft.email.trim().toLowerCase(),prefix:draft.prefix.trim(),firstName:draft.firstName.trim(),lastName:draft.lastName.trim(),title:draft.title.trim(),unit:draft.unit.trim()}
  if (!normalized.prefix || !normalized.firstName || !normalized.lastName || !normalized.unit) throw new Error('กรุณากรอกข้อมูลบัญชีให้ครบ')
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(normalized.email) || normalized.email.length>254) throw new Error('กรุณากรอกอีเมลให้ถูกต้อง')
  if (!(['นักวิจัย','ผู้ประสานงาน','ผู้ดูแลระบบ'] as Role[]).includes(normalized.role)) throw new Error('กรุณาเลือกสิทธิ์การใช้งาน')
  if (!(['PENDING','APPROVED','SUSPENDED','REJECTED'] as AccountStatus[]).includes(normalized.status)) throw new Error('กรุณาเลือกสถานะบัญชี')
  return normalized
}

function requireUniqueEmail(accounts:DemoAccount[], email:string, exceptId?:string) {
  if(accounts.some(account=>account.id!==exceptId&&account.email.toLowerCase()===email)) throw new Error('อีเมลนี้ถูกใช้แล้ว')
}

export function addAccount(accounts:DemoAccount[], id:string, draft:AccountDraft, actorRole:Role):DemoAccount[] {
  requireAdmin(actorRole)
  if(accounts.some(account=>account.id===id)) throw new Error('รหัสบัญชีซ้ำ')
  const normalized=normalizeAccountDraft(draft)
  if(normalized.status!=='APPROVED'&&normalized.status!=='SUSPENDED') throw new Error('เลือกสถานะใช้งานปกติหรือระงับการใช้งานสำหรับบัญชีใหม่')
  requireUniqueEmail(accounts,normalized.email)
  return [...accounts,{id,...normalized,name:`${normalized.firstName} ${normalized.lastName}`,firstLoginAt:'—'}]
}

export function updateAccount(accounts:DemoAccount[], id:string, draft:AccountDraft, actorRole:Role):DemoAccount[] {
  requireAdmin(actorRole)
  if(!accounts.some(account=>account.id===id)) throw new Error('ไม่พบบัญชีผู้ใช้งาน')
  const normalized=normalizeAccountDraft(draft)
  requireUniqueEmail(accounts,normalized.email,id)
  return accounts.map(account=>account.id===id?{...account,...normalized,name:`${normalized.firstName} ${normalized.lastName}`}:account)
}

export function deleteAccount(accounts:DemoAccount[], id:string, actorRole:Role):DemoAccount[] {
  requireAdmin(actorRole)
  if(!accounts.some(account=>account.id===id)) throw new Error('ไม่พบบัญชีผู้ใช้งาน')
  return accounts.filter(account=>account.id!==id)
}

export function decideAccount(accounts: DemoAccount[], email: string, actorRole: Role, decision: AccountDecision): DemoAccount[] {
  requireAdmin(actorRole)
  const target = accounts.find(account => account.email === email)
  if (!target) throw new Error('ไม่พบบัญชีผู้ใช้งาน')
  if (target.status !== 'PENDING') throw new Error('บัญชีนี้ไม่ได้อยู่ในสถานะรออนุมัติ')
  return accounts.map(account => account.email === email ? { ...account, status: decision } : account)
}
