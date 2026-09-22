import { useState } from 'react'
import { adminModules } from './adminModules'
import type { AdminView } from './adminModules'

export default function AdminPreview({view}:{view:AdminView}) {
  const module=adminModules.find(item=>item.id===view)!
  const [scenario,setScenario]=useState('normal')
  return <main className="page"><header className="page-header"><div><p>ระบบจัดการข้อมูลพื้นฐาน</p><h1>{module.title}</h1><p>{module.description}</p></div><div className="admin-page-actions"><span className="prototype-pill">สำหรับผู้ดูแลระบบ</span>{['academic-titles','about','organization'].includes(view)&&<button type="button" className="primary admin-edit-preview" disabled title="ยังไม่เปิดใช้งานใน demo">แก้ไข</button>}</div></header>
    <section className="register admin-preview"><div className="register-head"><p className="helper-text">หน้าตัวอย่างสำหรับสาธิต ยังไม่มีการบันทึกหรือเปลี่ยนแปลงข้อมูลจริง</p><label className="scenario">สถานะสาธิต<select value={scenario} onChange={event=>setScenario(event.target.value)}><option value="normal">ข้อมูลพร้อม</option><option value="loading">กำลังโหลด</option><option value="error">เกิดข้อผิดพลาด</option><option value="empty">ยังไม่มีข้อมูล</option></select></label></div>
      {scenario==='loading'?<div className="state-panel" role="status"><span className="loader"/><b>กำลังโหลดข้อมูล</b></div>:scenario==='error'?<div className="state-panel error" role="alert"><b>โหลดข้อมูลไม่สำเร็จ</b><p>กรุณาลองอีกครั้ง</p><button className="secondary" onClick={()=>setScenario('normal')}>ลองใหม่</button></div>:scenario==='empty'?<div className="empty-state"><b>ยังไม่มีข้อมูล{module.title}</b><p>ข้อมูลจะแสดงที่นี่เมื่อมีรายการในระบบ</p></div>:<PreviewContent view={view}/>}
    </section></main>
}

function PreviewContent({view}:{view:AdminView}) {
  if(view==='approvals'||view==='users') return <div className="table-wrap"><table><thead><tr><th>ชื่อผู้ใช้งาน</th><th>ตำแหน่ง / หน่วยงาน</th><th>อีเมล</th><th>สิทธิ์การใช้งาน</th><th>สถานะบัญชี</th></tr></thead><tbody>{[{name:'อภิชาติ ตั้งมั่น',title:'ดร.',unit:'คณะวิทยาศาสตร์',email:'researcher@example.org',role:'นักวิจัย'},{name:'ณัฐชา พูนผล',title:'เจ้าหน้าที่',unit:'สำนักงานบริหารการวิจัย',email:'coordinator@example.org',role:'ผู้ประสานงาน'}].map(user=><tr key={user.email}><td>{user.name}</td><td>{user.title}<small>{user.unit}</small></td><td>{user.email}</td><td>{user.role}</td><td><span className={view==='approvals'?'status warn':'status'}>{view==='approvals'?'รออนุมัติ':'อนุมัติแล้ว'}</span></td></tr>)}</tbody></table></div>
  if(view==='academic-titles') return <ul className="academic-list">{['ศาสตราจารย์ (ศ.)','รองศาสตราจารย์ (รศ.)','ผู้ช่วยศาสตราจารย์ (ผศ.)','อาจารย์'].map(title=><li key={title}>{title}</li>)}</ul>
  if(view==='about') return <div className="admin-copy"><h2>สำนักงานบริหารการวิจัย นวัตกรรมและการสร้างสรรค์</h2><p>มหาวิทยาลัยศิลปากร</p><p className="muted">พื้นที่ตัวอย่างสำหรับรายละเอียดเกี่ยวกับสำนักงาน สวนส.</p></div>
  if(view==='organization') return <dl className="organization-facts">{[['มหาวิทยาลัย','มหาวิทยาลัยศิลปากร'],['สำนักงาน','สำนักงานบริหารการวิจัย นวัตกรรมและการสร้างสรรค์'],['ชื่อระบบ','ระบบติดตามสถานะงานวิจัย'],['ที่อยู่สำนักงาน','ยังไม่ได้ระบุในข้อมูลตัวอย่าง']].map(([label,value])=><div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>
  return <div className="table-wrap"><table><caption className="helper-text">เหตุการณ์สมมติสำหรับแสดงรูปแบบรายงาน</caption><thead><tr><th>วันและเวลา</th><th>ผู้ดำเนินการ</th><th>การกระทำ</th><th>ข้อมูลที่เกี่ยวข้อง</th></tr></thead><tbody><tr><td>17 ก.ย. 2569 09:30</td><td>กมลชนก สาธิต</td><td>ปรับกระบวนการงานวิจัย</td><td>SURDI-2569-014<small>การจัดสรรค่าธรรมเนียม → การติดตามส่งรายงาน</small></td></tr></tbody></table></div>
}
