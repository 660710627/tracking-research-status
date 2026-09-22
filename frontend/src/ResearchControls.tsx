import { useEffect, useId, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { isTerminal, nextStatuses } from './research'
import type { Research, ResearchAction } from './research'

function Modal({title, children, onCancel}:{title:string; children:ReactNode; onCancel:()=>void}) {
  const ref = useRef<HTMLDialogElement>(null)
  const titleId = useId()
  useEffect(() => {
    const dialog = ref.current!
    const previousFocus = document.activeElement as HTMLElement | null
    dialog.showModal()
    dialog.querySelector<HTMLButtonElement>('[data-cancel]')?.focus()
    return () => { dialog.close(); (previousFocus?.isConnected ? previousFocus : document.getElementById('main'))?.focus() }
  }, [])
  return <dialog ref={ref} className="confirmation" aria-labelledby={titleId} onCancel={event=>{event.preventDefault();onCancel()}} onKeyDown={event=>{
    if(event.key!=='Tab')return
    const controls=Array.from(event.currentTarget.querySelectorAll<HTMLElement>('button:not(:disabled),input:not(:disabled),select:not(:disabled),textarea:not(:disabled),[tabindex="0"]'))
    const first=controls[0], last=controls[controls.length-1]
    if(event.shiftKey&&document.activeElement===first){event.preventDefault();last?.focus()}
    else if(!event.shiftKey&&document.activeElement===last){event.preventDefault();first?.focus()}
  }}>
    <h2 id={titleId}>{title}</h2>{children}
  </dialog>
}

export function ConfirmResearch({research,action,onCancel,onConfirm}:{research:Research;action:Exclude<ResearchAction,{type:'edit'}>;onCancel:()=>void;onConfirm:()=>void}) {
  const deleting=action.type==='delete'
  const ending=action.type==='status' && isTerminal(action.status)
  const title=deleting?'ยืนยันการลบงานวิจัย':action.type==='process'?'ยืนยันการปรับกระบวนการ':action.status==='ยุติโครงการ'?'ยืนยันการยุติโครงการ':'ยืนยันการปรับสถานะ'
  return <Modal title={title} onCancel={onCancel}>
    <p className="confirmation-project">{research.title}<small>{research.contract}</small></p>
    {action.type==='status'&&<dl className="change-summary"><div><dt>สถานะปัจจุบัน</dt><dd>{research.status}</dd></div><div><dt>สถานะใหม่</dt><dd>{action.status}</dd></div></dl>}
    {action.type==='process'&&<p>เปลี่ยนจากขั้นตอนที่ {research.process} เป็นขั้นตอนที่ {action.process}</p>}
    <p className="confirmation-warning">{deleting?'งานวิจัยนี้จะถูกนำออกจากรายการข้อมูลตัวอย่าง การลบไม่สามารถเรียกคืนได้ในรอบสาธิตนี้':ending?'เมื่อยืนยันแล้ว จะไม่สามารถปรับสถานะหรือกระบวนการของโครงการนี้ได้อีก':action.type==='status'?'เมื่อยืนยันแล้ว จะไม่สามารถย้อนกลับไปสถานะก่อนหน้าได้':'เมื่อยืนยันแล้ว จะไม่สามารถย้อนกลับไปกระบวนการก่อนหน้าได้'}</p>
    <div className="form-actions"><button data-cancel className="secondary" onClick={onCancel}>ยกเลิก</button><button className={deleting||(action.type==='status'&&action.status==='ยุติโครงการ')?'danger':'primary'} onClick={onConfirm}>{deleting?'ยืนยันการลบ':action.type==='status'&&action.status==='ยุติโครงการ'?'ยืนยันยุติโครงการ':'ยืนยันการเปลี่ยนแปลง'}</button></div>
  </Modal>
}

export function StatusPanel({research,canManage,onRequest}:{research:Research;canManage:boolean;onRequest:(action:ResearchAction)=>void}) {
  const [target,setTarget]=useState('')
  const options=nextStatuses(research.status).filter(status=>status!=='ยุติโครงการ')
  const terminal=isTerminal(research.status)
  return <section className="status-panel" aria-labelledby="research-status-heading">
    <h2 id="research-status-heading">สถานะงานวิจัย</h2>
    <p className="status-caption">สถานะปัจจุบัน</p><p className={terminal?'status done':'status'}>{research.status}</p>
    {canManage&&!terminal?<><label className="status-select">เปลี่ยนสถานะเป็น<select value={target} onChange={event=>setTarget(event.target.value)} aria-describedby="status-help"><option value="">เลือกสถานะใหม่</option>{options.map(status=><option key={status}>{status}</option>)}</select></label><p id="status-help" className="helper-text">เลือกได้เฉพาะสถานะที่ไปข้างหน้า เมื่อบันทึกแล้วจะย้อนกลับไม่ได้</p><div className="status-actions"><button className="primary" disabled={!options.some(status=>status===target)} onClick={()=>onRequest({type:'status',status:target})}>ปรับสถานะ</button><button className="danger-outline" onClick={()=>onRequest({type:'status',status:'ยุติโครงการ'})}>ยุติโครงการ</button></div></>:<p className="helper-text">{terminal?'โครงการสิ้นสุดแล้ว ไม่สามารถปรับสถานะหรือกระบวนการได้':'ผู้ประสานงานหรือผู้ดูแลระบบเป็นผู้ปรับสถานะโครงการ'}</p>}
  </section>
}
