import { useState } from 'react'
import type { AcademicTitle } from './academicTitlesDemo'
import type { AccountDraft, DemoAccount } from './approvalDemo'
import type { Role } from './research'
import { Modal } from './ResearchControls'

type Props = {
  account: DemoAccount | null
  titles: AcademicTitle[]
  onCancel: () => void
  onSave: (draft: AccountDraft) => string | null
}

const emptyDraft: AccountDraft = {
  email: '', prefix: '', firstName: '', lastName: '', title: '', unit: '',
  role: 'นักวิจัย', status: 'APPROVED', notifyResearch: false,
}

export default function AccountEditor({account,titles,onCancel,onSave}:Props) {
  const [draft,setDraft]=useState<AccountDraft>(()=>account?{
    email:account.email,prefix:account.prefix,firstName:account.firstName,lastName:account.lastName,
    title:account.title,unit:account.unit,role:account.role,status:account.status,notifyResearch:account.notifyResearch,
  }:emptyDraft)
  const [error,setError]=useState('')
  const titleOptions=[...new Set([...titles.map(title=>title.thaiAbbreviation),'เจ้าหน้าที่',...(account?.title?[account.title]:[])])]
  const save=()=>{
    const message=onSave(draft)
    setError(message??'')
  }
  return <Modal title={account?'แก้ไขข้อมูลบัญชีผู้ใช้งาน':'เพิ่มบัญชีผู้ใช้งาน'} onCancel={onCancel} className="account-modal">
    <button type="button" data-cancel className="account-modal-close" aria-label="ปิดหน้าต่าง" onClick={onCancel}>×</button>
    <form onSubmit={event=>{event.preventDefault();save()}}>
      {error&&<p className="notice" role="alert">{error}</p>}
      <div className="account-form-grid">
        <label>อีเมล <span className="required">*</span><input type="email" required autoComplete="email" value={draft.email} onChange={event=>setDraft({...draft,email:event.target.value})}/><small>อีเมลที่ระบุจะถูกใช้สำหรับบัญชีเข้าสู่ระบบ</small></label>
        <div className="account-password"><span>รหัสผ่าน</span><label><input type="checkbox" disabled/> กำหนดรหัสผ่านใหม่</label><input type="password" disabled placeholder="ไม่รองรับในเดโม"/><small>เดโมนี้ไม่เปลี่ยนรหัสผ่านหรือเชื่อมระบบเข้าสู่ระบบจริง</small></div>
        <label>คำนำหน้าชื่อ <span className="required">*</span><select required value={draft.prefix} onChange={event=>setDraft({...draft,prefix:event.target.value})}><option value="">-- เลือกคำนำหน้า --</option>{['นาย','นาง','นางสาว','ไม่ระบุ'].map(prefix=><option key={prefix}>{prefix}</option>)}</select></label>
        <label>ตำแหน่งทางวิชาการ / ตำแหน่ง<select value={draft.title} onChange={event=>setDraft({...draft,title:event.target.value})}><option value="">-- ไม่ระบุ --</option>{titleOptions.map(title=><option key={title}>{title}</option>)}</select></label>
        <label>ชื่อ <span className="required">*</span><input required autoComplete="given-name" value={draft.firstName} onChange={event=>setDraft({...draft,firstName:event.target.value})}/></label>
        <label>นามสกุล <span className="required">*</span><input required autoComplete="family-name" value={draft.lastName} onChange={event=>setDraft({...draft,lastName:event.target.value})}/></label>
        <label>หน่วยงาน <span className="required">*</span><input required value={draft.unit} onChange={event=>setDraft({...draft,unit:event.target.value})}/></label>
        <label>ประเภทผู้ใช้งาน <span className="required">*</span><select value={draft.role} onChange={event=>setDraft({...draft,role:event.target.value as Role})}>{(['นักวิจัย','ผู้ประสานงาน','ผู้ดูแลระบบ'] as Role[]).map(role=><option key={role}>{role}</option>)}</select></label>
        <fieldset className="account-status-choice"><legend>สถานะการใช้งาน <span className="required">*</span></legend>
          <label><input type="radio" name="account-status" checked={draft.status==='APPROVED'} onChange={()=>setDraft({...draft,status:'APPROVED'})}/> ใช้งานปกติ</label>
          <label><input type="radio" name="account-status" checked={draft.status==='SUSPENDED'} onChange={()=>setDraft({...draft,status:'SUSPENDED'})}/> ระงับการใช้งาน</label>
        </fieldset>
        <fieldset className="account-notification"><legend>รับแจ้งเตือนจากระบบ</legend><label><input type="checkbox" checked={draft.notifyResearch} onChange={event=>setDraft({...draft,notifyResearch:event.target.checked})}/> ระบบจัดการข้อมูลการบริหารจัดการทุนวิจัย</label><small>บันทึกตัวเลือกในเดโมเท่านั้น ยังไม่มีการส่งแจ้งเตือนจริง</small></fieldset>
      </div>
      <div className="form-actions account-form-actions"><button type="button" className="secondary" onClick={onCancel}>ปิดหน้าต่าง</button><button type="submit" className="primary">{account?'บันทึกการแก้ไข':'เพิ่มบัญชี'}</button></div>
    </form>
  </Modal>
}
