import { useState } from 'react'
import { adminModules } from './adminModules'
import type { AdminView } from './adminModules'
import { approvalAccounts, managedAccounts } from './approvalDemo'
import type { AccountDecision, AccountDraft, DemoAccount } from './approvalDemo'
import type { AcademicTitle, AcademicTitleDraft } from './academicTitlesDemo'
import AccountEditor from './AccountEditor'
import { Modal } from './ResearchControls'

type Props = {
  view: AdminView
  accounts: DemoAccount[]
  onDecide: (email: string, decision: AccountDecision) => string | null
  onSaveAccount: (id: string | null, draft: AccountDraft) => string | null
  onDeleteAccount: (id: string) => string | null
  titles: AcademicTitle[]
  onSaveTitle: (id: string | null, draft: AcademicTitleDraft) => string | null
  onDeleteTitle: (id: string) => string | null
}
type TitleDialog = { type: 'create' } | { type: 'edit' | 'delete'; id: string }
type AccountDialog = { type: 'create' } | { type: 'edit' | 'delete'; id: string }
const blankTitle: AcademicTitleDraft = { thaiAbbreviation: '', thaiTitle: '', englishAbbreviation: '', englishTitle: '' }

export default function AdminPreview({view,accounts,onDecide,onSaveAccount,onDeleteAccount,titles,onSaveTitle,onDeleteTitle}:Props) {
  const module=adminModules.find(item=>item.id===view)!
  const visibleAccounts=view==='users'?managedAccounts(accounts):view==='approvals'?approvalAccounts(accounts):accounts
  const [scenario,setScenario]=useState('normal')
  const [error,setError]=useState('')
  const [titleDialog,setTitleDialog]=useState<TitleDialog|null>(null)
  const [titleDraft,setTitleDraft]=useState<AcademicTitleDraft>(blankTitle)
  const [titleError,setTitleError]=useState('')
  const [accountDialog,setAccountDialog]=useState<AccountDialog|null>(null)
  const [accountError,setAccountError]=useState('')
  const selectedAccount=accountDialog&&accountDialog.type!=='create'?accounts.find(account=>account.id===accountDialog.id):undefined
  const selectedTitle=titleDialog&&titleDialog.type!=='create'?titles.find(title=>title.id===titleDialog.id):undefined
  const openCreate=()=>{setTitleDraft(blankTitle);setTitleError('');setTitleDialog({type:'create'})}
  const openEdit=(title:AcademicTitle)=>{setTitleDraft({thaiAbbreviation:title.thaiAbbreviation,thaiTitle:title.thaiTitle,englishAbbreviation:title.englishAbbreviation,englishTitle:title.englishTitle});setTitleError('');setTitleDialog({type:'edit',id:title.id})}
  const openDelete=(id:string)=>{setTitleError('');setTitleDialog({type:'delete',id})}
  const saveTitle=()=>{
    if(!titleDialog||titleDialog.type==='delete')return
    const message=onSaveTitle(titleDialog.type==='create'?null:titleDialog.id,titleDraft)
    setTitleError(message??'')
    if(!message)setTitleDialog(null)
  }
  const deleteTitle=()=>{
    if(!titleDialog||titleDialog.type!=='delete')return
    const message=onDeleteTitle(titleDialog.id)
    setTitleError(message??'')
    if(!message)setTitleDialog(null)
  }
  const saveAccount=(draft:AccountDraft):string|null=>{
    if(!accountDialog||accountDialog.type==='delete')return 'ไม่สามารถบันทึกบัญชีผู้ใช้งานได้'
    const message=onSaveAccount(accountDialog.type==='create'?null:accountDialog.id,draft)
    if(!message)setAccountDialog(null)
    return message
  }
  const deleteAccount=()=>{
    if(!accountDialog||accountDialog.type!=='delete')return
    const message=onDeleteAccount(accountDialog.id)
    setAccountError(message??'')
    if(!message)setAccountDialog(null)
  }
  const decide=(email:string,decision:AccountDecision):string|null=>{
    setError(onDecide(email,decision)??'')
    return null
  }
  return <main className="page"><header className="page-header"><div><p>ระบบจัดการข้อมูลพื้นฐาน</p><h1>{module.title}</h1><p>{module.description}</p></div><div className="admin-page-actions"><span className="prototype-pill">สำหรับผู้ดูแลระบบ</span>{view==='users'?<button type="button" className="primary admin-edit-preview" onClick={()=>{setAccountError('');setAccountDialog({type:'create'})}}>＋ เพิ่มบัญชี</button>:view==='academic-titles'?<button type="button" className="primary admin-edit-preview" onClick={openCreate}>＋ เพิ่มตำแหน่ง</button>:['about','organization'].includes(view)&&<button type="button" className="primary admin-edit-preview" disabled title="ยังไม่เปิดใช้งานใน demo">แก้ไข</button>}</div></header>
    <section className="register admin-preview"><div className="register-head"><p className="helper-text">{view==='academic-titles'||view==='users'?'ข้อมูลตัวอย่าง เปลี่ยนแปลงชั่วคราวและจะคืนค่าเมื่อโหลดหน้าใหม่':'หน้าตัวอย่างสำหรับสาธิต ยังไม่มีการบันทึกหรือเปลี่ยนแปลงข้อมูลจริง'}</p><label className="scenario">สถานะสาธิต<select value={scenario} onChange={event=>setScenario(event.target.value)}><option value="normal">ข้อมูลพร้อม</option><option value="loading">กำลังโหลด</option><option value="error">เกิดข้อผิดพลาด</option><option value="empty">ยังไม่มีข้อมูล</option></select></label></div>
      {error&&<p className="notice" role="alert">{error}</p>}
      {scenario==='loading'?<div className="state-panel" role="status"><span className="loader"/><b>กำลังโหลดข้อมูล</b></div>:scenario==='error'?<div className="state-panel error" role="alert"><b>โหลดข้อมูลไม่สำเร็จ</b><p>กรุณาลองอีกครั้ง</p><button className="secondary" onClick={()=>setScenario('normal')}>ลองใหม่</button></div>:scenario==='empty'?<div className="empty-state"><b>ยังไม่มีข้อมูล{module.title}</b><p>ข้อมูลจะแสดงที่นี่เมื่อมีรายการในระบบ</p></div>:view==='users'&&visibleAccounts.length===0?<div className="empty-state"><b>ยังไม่มีบัญชีที่จัดการได้</b><p>เมื่ออนุมัติบัญชีหรือเพิ่มบัญชีใหม่ รายการจะแสดงที่นี่</p><button className="primary" onClick={()=>setAccountDialog({type:'create'})}>เพิ่มบัญชี</button></div>:view==='academic-titles'&&titles.length===0?<div className="empty-state"><b>ยังไม่มีตำแหน่งทางวิชาการ</b><p>เริ่มต้นโดยเพิ่มตำแหน่งทางวิชาการ</p><button className="primary" onClick={openCreate}>เพิ่มตำแหน่ง</button></div>:<PreviewContent view={view} accounts={visibleAccounts} onDecide={decide} titles={titles} onEditTitle={openEdit} onDeleteTitle={openDelete} onEditAccount={account=>{setAccountError('');setAccountDialog({type:'edit',id:account.id})}} onDeleteAccount={id=>{setAccountError('');setAccountDialog({type:'delete',id})}}/>}
    </section>
    {accountDialog&&accountDialog.type!=='delete'&&<AccountEditor key={accountDialog.type==='create'?'new':accountDialog.id} account={accountDialog.type==='create'?null:selectedAccount??null} titles={titles} onCancel={()=>setAccountDialog(null)} onSave={saveAccount}/>}
    {accountDialog?.type==='delete'&&<Modal title="ยืนยันการลบบัญชีผู้ใช้งาน" onCancel={()=>setAccountDialog(null)}>
      {accountError&&<p className="notice" role="alert">{accountError}</p>}
      <p className="confirmation-project">{selectedAccount?.name}<small>{selectedAccount?.email}</small></p>
      <p className="confirmation-warning">บัญชีนี้จะถูกลบออกจากข้อมูลสาธิตในรอบการใช้งานนี้</p>
      <div className="form-actions"><button type="button" data-cancel className="secondary" onClick={()=>setAccountDialog(null)}>ยกเลิก</button><button type="button" className="danger" onClick={deleteAccount}>ยืนยันการลบ</button></div>
    </Modal>}
    {titleDialog&&<Modal title={titleDialog.type==='create'?'เพิ่มตำแหน่งทางวิชาการ':titleDialog.type==='edit'?'แก้ไขตำแหน่งทางวิชาการ':'ยืนยันการลบตำแหน่ง'} onCancel={()=>setTitleDialog(null)}>
      {titleError&&<p className="notice" role="alert">{titleError}</p>}
      {titleDialog.type==='delete'?<>
        <p className="confirmation-project">{selectedTitle?.thaiAbbreviation} · {selectedTitle?.thaiTitle}<small>{selectedTitle?.englishAbbreviation} · {selectedTitle?.englishTitle}</small></p>
        <p className="confirmation-warning">รายการนี้จะถูกลบออกจากข้อมูลสาธิตในรอบการใช้งานนี้</p>
        <div className="form-actions"><button type="button" data-cancel className="secondary" onClick={()=>setTitleDialog(null)}>ยกเลิก</button><button type="button" className="danger" onClick={deleteTitle}>ยืนยันการลบ</button></div>
      </>:<form onSubmit={event=>{event.preventDefault();saveTitle()}}>
        <p className="helper-text">กรอกข้อมูลให้ครบทั้ง 4 ช่อง</p>
        <div className="academic-title-fields">
          <label>แบบย่อ-ไทย *<input required value={titleDraft.thaiAbbreviation} onChange={event=>setTitleDraft({...titleDraft,thaiAbbreviation:event.target.value})}/></label>
          <label>ตำแหน่งทางวิชาการ (ไทย) *<input required value={titleDraft.thaiTitle} onChange={event=>setTitleDraft({...titleDraft,thaiTitle:event.target.value})}/></label>
          <label>แบบย่อ-อังกฤษ *<input required value={titleDraft.englishAbbreviation} onChange={event=>setTitleDraft({...titleDraft,englishAbbreviation:event.target.value})}/></label>
          <label>ตำแหน่งทางวิชาการ (อังกฤษ) *<input required value={titleDraft.englishTitle} onChange={event=>setTitleDraft({...titleDraft,englishTitle:event.target.value})}/></label>
        </div>
        <div className="form-actions"><button type="button" data-cancel className="secondary" onClick={()=>setTitleDialog(null)}>ยกเลิก</button><button type="submit" className="primary">{titleDialog.type==='create'?'เพิ่มตำแหน่ง':'บันทึกการแก้ไข'}</button></div>
      </form>}
    </Modal>}
  </main>
}

type PreviewProps = Pick<Props,'view'|'accounts'|'onDecide'|'titles'> & { onEditTitle:(title:AcademicTitle)=>void; onDeleteTitle:(id:string)=>void; onEditAccount:(account:DemoAccount)=>void; onDeleteAccount:(id:string)=>void }

function PreviewContent({view,accounts,onDecide,titles,onEditTitle,onDeleteTitle,onEditAccount,onDeleteAccount}:PreviewProps) {
  if(view==='approvals') return <div className="table-wrap"><table className="approval-table">
    <caption className="sr-only">รายการพิจารณาบัญชีผู้ใช้</caption>
    <thead><tr><th scope="col">วันที่เข้าสู่ระบบ</th><th scope="col">ชื่อผู้ใช้งาน</th><th scope="col">ตำแหน่ง/หน่วยงาน</th><th scope="col">อีเมล</th><th scope="col">สิทธิ์การใช้งาน</th><th scope="col">สถานะ</th><th scope="col">การดำเนินการ</th></tr></thead>
    <tbody>{accounts.map(user=><tr key={user.id}>
      <td>{user.firstLoginAt}</td><td>{user.name}</td><td>{user.title}<small>{user.unit}</small></td><td>{user.email}</td><td>{user.role}</td>
      <td><AccountStatus status={user.status}/></td>
      <td>{user.status==='PENDING'&&<div className="approval-actions"><button type="button" className="approval-accept" aria-label={`อนุมัติบัญชี ${user.name}`} onClick={()=>onDecide(user.email,'APPROVED')}>อนุมัติ</button><button type="button" className="approval-reject" aria-label={`ปฏิเสธบัญชี ${user.name}`} onClick={()=>onDecide(user.email,'REJECTED')}>ปฏิเสธ</button></div>}</td>
    </tr>)}</tbody>
  </table></div>
  if(view==='users') return <div className="table-wrap"><table className="users-table"><caption className="sr-only">รายการบัญชีผู้ใช้งานตัวอย่าง</caption><thead><tr><th scope="col">ชื่อผู้ใช้งาน</th><th scope="col">ตำแหน่ง / หน่วยงาน</th><th scope="col">อีเมล</th><th scope="col">สิทธิ์การใช้งาน</th><th scope="col">สถานะบัญชี</th><th scope="col">การดำเนินการ</th></tr></thead><tbody>{accounts.map(user=><tr key={user.id}><td>{user.prefix} {user.name}</td><td>{user.title||'ไม่ระบุ'}<small>{user.unit}</small></td><td>{user.email}</td><td>{user.role}</td><td><AccountStatus status={user.status} activeLabel="ใช้งานปกติ"/></td><td><div className="account-row-actions"><button type="button" className="secondary" aria-label={`แก้ไขบัญชี ${user.name}`} onClick={()=>onEditAccount(user)}>แก้ไข</button><button type="button" className="danger-outline" aria-label={`ลบบัญชี ${user.name}`} onClick={()=>onDeleteAccount(user.id)}>ลบ</button></div></td></tr>)}</tbody></table></div>
  if(view==='academic-titles') return <div className="table-wrap"><table className="academic-title-table">
    <caption className="sr-only">รายการตำแหน่งทางวิชาการตัวอย่าง</caption>
    <thead><tr><th scope="col">แบบย่อ-ไทย</th><th scope="col">ตำแหน่งทางวิชาการ (ไทย)</th><th scope="col">แบบย่อ-อังกฤษ</th><th scope="col">ตำแหน่งทางวิชาการ (อังกฤษ)</th><th scope="col">การดำเนินการ</th></tr></thead>
    <tbody>{titles.map(title=><tr key={title.id}><td>{title.thaiAbbreviation}</td><td>{title.thaiTitle}</td><td>{title.englishAbbreviation}</td><td>{title.englishTitle}</td><td><div className="academic-row-actions"><button type="button" className="secondary" aria-label={`แก้ไขตำแหน่ง ${title.thaiAbbreviation}`} onClick={()=>onEditTitle(title)}>แก้ไข</button><button type="button" className="danger-outline" aria-label={`ลบตำแหน่ง ${title.thaiAbbreviation}`} onClick={()=>onDeleteTitle(title.id)}>ลบ</button></div></td></tr>)}</tbody>
  </table></div>
  if(view==='about') return <div className="admin-copy"><h2>สำนักงานบริหารการวิจัย นวัตกรรมและการสร้างสรรค์</h2><p>มหาวิทยาลัยศิลปากร</p><p className="muted">พื้นที่ตัวอย่างสำหรับรายละเอียดเกี่ยวกับสำนักงาน สวนส.</p></div>
  if(view==='organization') return <dl className="organization-facts">{[['มหาวิทยาลัย','มหาวิทยาลัยศิลปากร'],['สำนักงาน','สำนักงานบริหารการวิจัย นวัตกรรมและการสร้างสรรค์'],['ชื่อระบบ','ระบบติดตามสถานะงานวิจัย'],['ที่อยู่สำนักงาน','ยังไม่ได้ระบุในข้อมูลตัวอย่าง']].map(([label,value])=><div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>
  return <div className="table-wrap"><table><caption className="helper-text">เหตุการณ์สมมติสำหรับแสดงรูปแบบรายงาน</caption><thead><tr><th>วันและเวลา</th><th>ผู้ดำเนินการ</th><th>การกระทำ</th><th>ข้อมูลที่เกี่ยวข้อง</th></tr></thead><tbody><tr><td>17 ก.ย. 2569 09:30</td><td>กมลชนก สาธิต</td><td>ปรับกระบวนการงานวิจัย</td><td>SURDI-2569-014<small>การจัดสรรค่าธรรมเนียม → การติดตามส่งรายงาน</small></td></tr></tbody></table></div>
}

function AccountStatus({status,activeLabel='อนุมัติแล้ว'}:{status:DemoAccount['status'];activeLabel?:string}) {
  return <span className={status==='PENDING'?'status warn':status==='REJECTED'||status==='SUSPENDED'?'status rejected':'status'}>{status==='PENDING'?'รออนุมัติ':status==='APPROVED'?activeLabel:status==='SUSPENDED'?'ระงับการใช้งาน':'ปฏิเสธแล้ว'}</span>
}
