import { useEffect, useRef, useState } from 'react'
import type { AcademicTitle } from './academicTitlesDemo'
export type ProjectMember = {academicTitle:string;name:string;email:string;contribution:number}

export default function ProjectMemberPage({heading,initial,titles,onSave,onCancel}:{heading:string;initial?:ProjectMember;titles:AcademicTitle[];onSave:(member:ProjectMember)=>void;onCancel:()=>void}) {
  const [member,setMember]=useState<ProjectMember>(initial??{academicTitle:'',name:'',email:'',contribution:100})
  const headingRef=useRef<HTMLHeadingElement>(null)
  useEffect(()=>{headingRef.current?.focus()},[])
  return <main className="page"><button className="back-link" onClick={onCancel}>← กลับหน้าเพิ่มงานวิจัย</button><header className="page-header"><div><h1 tabIndex={-1} ref={headingRef}>{heading}</h1><p>กรอกข้อมูลบุคลากรสำหรับโครงการนี้</p></div></header>
    <form className="form-sheet" onSubmit={event=>{event.preventDefault();if(!member.name.trim())return;onSave({...member,name:member.name.trim(),email:member.email.trim()})}}>
      <div className="form-grid">
        <label>ตำแหน่งทางวิชาการ *<select required value={member.academicTitle} onChange={e=>setMember({...member,academicTitle:e.target.value})}><option value="">เลือกตำแหน่งทางวิชาการ</option>{titles.map(title=><option key={title.id} value={title.thaiAbbreviation}>{title.thaiTitle} ({title.thaiAbbreviation})</option>)}</select></label>
        <label>ชื่อ นามสกุล *<input required value={member.name} onChange={e=>setMember({...member,name:e.target.value})}/></label>
        <label>อีเมล *<input type="email" required value={member.email} onChange={e=>setMember({...member,email:e.target.value})}/></label>
        <label>สัดส่วนของโปรเจค (%) *<input type="number" min="0.01" max="100" step="0.01" required value={Number.isNaN(member.contribution)?'':member.contribution} onChange={e=>setMember({...member,contribution:e.target.valueAsNumber})}/></label>
      </div><p className="helper-text">ระบุสัดส่วนมากกว่า 0 ถึง 100% ทศนิยมไม่เกิน 2 ตำแหน่ง</p>
      <div className="form-actions"><button type="button" className="secondary" onClick={onCancel}>ยกเลิก</button><button className="primary" type="submit">บันทึกข้อมูลบุคลากร</button></div>
    </form></main>
}
