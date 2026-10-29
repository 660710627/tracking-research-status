import { useState } from 'react'
import { Modal } from './ResearchControls'
import { demoSubprocesses } from './processDemo'
import type { ProcessDetails } from './processDemo'
import './processPicker.css'

export function ProcessPicker({current,steps,onCancel,onSelect}:{current:number;steps:string[];onCancel:()=>void;onSelect:(process:number,details:ProcessDetails)=>void}) {
  const [selected,setSelected]=useState('')
  const [subprocess,setSubprocess]=useState('')
  const [note,setNote]=useState('')
  const [file,setFile]=useState<File|null>(null)
  const options=demoSubprocesses[Number(selected)]??[]
  const option=options.find(item=>item.name===subprocess)
  return <Modal title="เปลี่ยนกระบวนการงานวิจัย" onCancel={onCancel} className="process-picker">
    <p className="process-current">กระบวนการปัจจุบัน<strong>{current}. {steps[current-1]}</strong></p>
    <form onSubmit={event=>{event.preventDefault();if(option)onSelect(Number(selected),{subprocess,note:note.trim(),file})}}>
      <label className="process-field">กระบวนการโครงการวิจัย<select required value={selected} onChange={event=>{setSelected(event.target.value);setSubprocess('');setFile(null)}}><option value="">--เลือกรายการกระบวนการ--</option>{steps.map((step,index)=><option key={step} className={index+1===current?'current-process-option':undefined} value={index+1} disabled={index+1===current}>{index+1}. {step}{index+1===current?' (ปัจจุบัน)':''}</option>)}</select></label>
      {selected&&<div className="process-expanded"><label className="process-field">กระบวนการย่อย<select required value={subprocess} onChange={event=>{setSubprocess(event.target.value);setFile(null)}}><option value="">--เลือกกระบวนการย่อย--</option>{options.map(item=><option key={item.name}>{item.name}</option>)}</select></label>
        {option?.attachment&&<label className="process-field">ไฟล์แนบ (ถ้ามี)<input key={`${selected}-${subprocess}`} type="file" onChange={event=>setFile(event.target.files?.[0]??null)}/></label>}
        <small className="helper-text">รายการกระบวนการย่อยสมมติสำหรับ demo</small>
      </div>}
      <label className="process-field">หมายเหตุ : (ความคิดเห็นถ้ามี)<textarea rows={3} value={note} onChange={event=>setNote(event.target.value)}/></label>
      <div className="form-actions"><button type="button" data-cancel className="secondary" onClick={onCancel}>ยกเลิก</button><button type="submit" className="primary" disabled={!option}>ดำเนินการต่อ</button></div>
    </form>
  </Modal>
}
