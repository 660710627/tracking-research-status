import { useCallback, useRef, useState } from 'react'
export default function AbstractEditor({thai,english,onSave}:{thai:string;english:string;onSave:(patch:{thai:string;english:string})=>string|null}) {
  const [draft,setDraft]=useState<{thai:string;english:string}|null>(null)
  const [message,setMessage]=useState('')
  const focusThai=useCallback((element:HTMLTextAreaElement|null)=>{element?.focus()},[])
  const button=useRef<HTMLButtonElement>(null)
  const close=()=>{setDraft(null);requestAnimationFrame(()=>button.current?.focus())}
  if(!draft)return <div className="research-toolbar"><button ref={button} className="secondary" onClick={()=>{setDraft({thai,english});setMessage('')}}>แก้ไขบทคัดย่อ</button>{message&&<span role="status">{message}</span>}</div>
  return <form className="form-sheet" aria-label="แก้ไขบทคัดย่อ" onSubmit={event=>{event.preventDefault();const error=onSave(draft);if(error){setMessage(error);return}close();setMessage('บันทึกบทคัดย่อแล้ว')}}>
    <h2>แก้ไขบทคัดย่อ</h2><p className="helper-text">แก้ไขได้เฉพาะบทคัดย่อภาษาไทยและภาษาอังกฤษของงานวิจัยนี้</p>
    {message&&<p role="alert" className="notice">{message}</p>}
    <div className="research-texts"><label>บทคัดย่อ (ไทย)<textarea ref={focusThai} rows={8} value={draft.thai} onChange={e=>setDraft({...draft,thai:e.target.value})}/></label><label>บทคัดย่อ (อังกฤษ)<textarea lang="en" rows={8} value={draft.english} onChange={e=>setDraft({...draft,english:e.target.value})}/></label></div>
    <div className="form-actions"><button type="button" className="secondary" onClick={()=>{close();setMessage('')}}>ยกเลิก</button><button className="primary">บันทึกบทคัดย่อ</button></div>
  </form>
}
