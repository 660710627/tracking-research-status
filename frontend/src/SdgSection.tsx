import { useEffect, useRef, useState } from 'react'
import SdgPicker from './SdgPicker'
import { sdgGoals } from './sdgs'

export default function SdgSection({value,disabledReason,onSave}:{value:number[];disabledReason?:string;onSave:(ids:number[])=>string|null}) {
  const disabled=!!disabledReason
  const [draft,setDraft]=useState<number[]|null>(null)
  const [error,setError]=useState('')
  const [success,setSuccess]=useState('')
  const editor=useRef<HTMLDivElement>(null)
  const editButton=useRef<HTMLButtonElement>(null)
  const wasEditing=useRef(false)
  const editing=draft!==null
  useEffect(()=>{
    if(editing)editor.current?.querySelector('input')?.focus()
    else if(wasEditing.current)editButton.current?.focus()
    wasEditing.current=editing
  },[editing])
  return <section className="sdg-section" aria-labelledby="sdg-heading"><div className="sdg-heading"><div><h2 id="sdg-heading">SDGs · เป้าหมายการพัฒนาที่ยั่งยืน</h2><p className="helper-text">เป้าหมายที่งานวิจัยนี้มีส่วนสนับสนุน</p></div>{!editing&&<button ref={editButton} type="button" className="secondary" disabled={disabled} onClick={()=>{setDraft([...value]);setError('');setSuccess('')}}>แก้ไข SDGs</button>}</div>
    {success&&<p className="sdg-success" role="status">{success}</p>}
    {editing?<div ref={editor}><SdgPicker value={draft} onChange={setDraft}/>{error&&<p className="notice" role="alert">{error}</p>}<div className="form-actions"><button type="button" className="secondary" onClick={()=>{setDraft(null);setError('')}}>ยกเลิกการแก้ไข SDGs</button><button type="button" className="primary" disabled={disabled} onClick={()=>{const message=onSave(draft);setError(message??'');if(!message){setDraft(null);setSuccess('บันทึก SDGs แล้ว')}}}>บันทึก SDGs</button></div></div>:value.length?<ul className="sdg-selected">{sdgGoals.filter(goal=>value.includes(goal.id)).map(goal=><li key={goal.id}><b>SDG {goal.id}</b><span>{goal.name}</span></li>)}</ul>:<p className="sdg-empty">ยังไม่ได้ระบุ SDGs ของงานวิจัยนี้</p>}
    {disabledReason&&<p className="helper-text">{disabledReason}</p>}
  </section>
}
