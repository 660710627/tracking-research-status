import { useEffect, useId, useRef } from 'react'
import type { ReactNode } from 'react'
import type { Research, ResearchDraft } from './research'

function Fact({label,wide,children}:{label:string;wide?:boolean;children:ReactNode}) {
  const id=useId()
  return <div className={wide?'wide':undefined}><dt id={id}>{label}</dt><dd>{children}</dd></div>
}

export default function ResearchSheet({research,fields,editing,onFields,onSave,onCancel}:{research:Research;fields:ResearchDraft;editing:boolean;onFields:(next:ResearchDraft)=>void;onSave:()=>void;onCancel:()=>void}) {
  const titleRef=useRef<HTMLInputElement>(null)
  useEffect(()=>{if(editing)titleRef.current?.focus()},[editing])
  const unspecified='ยังไม่ได้ระบุ'
  const text=(key:'title'|'unit'|'budgetUnit'|'startDate'|'endDate'|'lead'|'collaborators'|'fund'|'contract',label:string)=>editing?<input ref={key==='title'?titleRef:undefined} aria-label={label} required={['title','unit','endDate','lead','contract'].includes(key)} value={fields[key]} onChange={event=>onFields({...fields,[key]:event.target.value})}/>:fields[key]||unspecified
  const select=(key:'projectType'|'fundingType',label:string,options:string[])=>editing?<select aria-label={label} value={fields[key]} onChange={event=>onFields({...fields,[key]:event.target.value})}>{options.map(value=><option key={value}>{value}</option>)}</select>:fields[key]
  return <form id="research-edit-form" className={editing?'project-sheet research-sheet inline-editing':'project-sheet research-sheet'} aria-label="รายละเอียดงานวิจัย" onSubmit={event=>{event.preventDefault();if(editing)onSave()}}>
    <p className="detail-demo">{editing?'กำลังแก้ไขข้อมูล • กดบันทึกเพื่อใช้ข้อมูลใหม่ หรือยกเลิกเพื่อคืนค่าเดิม':'ข้อมูลประกอบตัวอย่างสำหรับสาธิต'}</p>
    <h2>ข้อมูลโครงการ</h2>
    <dl className="facts-grid">
      <Fact label="ชื่อโครงการ" wide>{text('title','ชื่อโครงการ')}</Fact>
      <Fact label="ประเภทโครงการ">{select('projectType','ประเภทโครงการ',['งานวิจัย','บริการวิชาการ'])}</Fact>
      <Fact label="ทุนอุดหนุน / ประเภทที่เกี่ยวข้อง">{editing?<select aria-label="ทุนอุดหนุน / ประเภทที่เกี่ยวข้อง" value={String(fields.isSubsidized)} onChange={event=>onFields({...fields,isSubsidized:event.target.value==='true'})}><option value="true">ทุนอุดหนุนการวิจัย</option><option value="false">ไม่ใช่ทุนอุดหนุน</option></select>:fields.isSubsidized?'ทุนอุดหนุนการวิจัย':'ไม่ใช่ทุนอุดหนุน'}</Fact>
      <Fact label="สถานะการดำเนินงาน">{research.kind==='โครงการต่อเนื่อง'?'โครงการต่อเนื่อง':'โครงการในงบประมาณ'}{editing&&<small className="readonly-note">กำหนดเมื่อสร้างโครงการ เปลี่ยนภายหลังไม่ได้</small>}</Fact>
      <Fact label="หน่วยงานรับผิดชอบโครงการ">{text('unit','หน่วยงานรับผิดชอบโครงการ')}</Fact>
      <Fact label="หน่วยงานรับผิดชอบงบประมาณ">{text('budgetUnit','หน่วยงานรับผิดชอบงบประมาณ')}</Fact>
      <Fact label="ระยะเวลาวิจัย เริ่ม–สิ้นสุด">{editing?<div className="date-edit"><label>วันเริ่มต้น{text('startDate','วันเริ่มต้น')}</label><label>วันสิ้นสุด{text('endDate','วันสิ้นสุด')}</label></div>:`${fields.startDate||unspecified} – ${fields.endDate}`}</Fact>
      <Fact label="งบประมาณโครงการ">{editing?<input aria-label="งบประมาณโครงการ (บาท)" required type="number" min="0.01" step="0.01" value={Number.isNaN(fields.budget)?'':fields.budget} onChange={event=>onFields({...fields,budget:event.target.valueAsNumber})}/>:<span className="budget-value">{fields.budget.toLocaleString('th-TH',{minimumFractionDigits:2})} บาท</span>}</Fact>
    </dl>
    <h2>บุคลากรและแหล่งทุน</h2>
    <dl className="facts-grid">
      <Fact label="หัวหน้าโครงการ">{text('lead','หัวหน้าโครงการ')}</Fact>
      <Fact label="ผู้ร่วมโครงการ">{text('collaborators','ผู้ร่วมโครงการ')}</Fact>
      <Fact label="ประเภทแหล่งทุน">{select('fundingType','ประเภทแหล่งทุน',['ภายใน','ภายนอก'])}</Fact>
      <Fact label="ชื่อแหล่งทุน">{text('fund','ชื่อแหล่งทุน')}</Fact>
      <Fact label="เลขที่สัญญาทุน">{text('contract','เลขที่สัญญาทุน')}</Fact>
      <Fact label="ไฟล์สัญญารับทุน"><span>{fields.contractFile?.name??'ยังไม่มีไฟล์แนบ'}</span>{editing&&<><input aria-label="เลือกไฟล์สัญญารับทุน PDF" type="file" accept=".pdf,application/pdf" onChange={event=>{const file=event.target.files?.[0];if(file)onFields({...fields,contractFile:file})}}/><small className="readonly-note">เลือก PDF ไม่เกิน 20 MB • เก็บใน demo จนกว่าจะรีโหลดหน้า</small></>}</Fact>
    </dl>
    <div className="research-texts">
      <section><h2>บทคัดย่อ (ไทย)</h2>{editing?<textarea aria-label="บทคัดย่อ (ไทย)" rows={9} value={fields.thai} onChange={event=>onFields({...fields,thai:event.target.value})}/>:<p className="preserve-lines">{fields.thai||unspecified}</p>}</section>
      <section><h2>บทคัดย่อ (อังกฤษ)</h2>{editing?<textarea aria-label="บทคัดย่อ (อังกฤษ)" lang="en" rows={9} value={fields.english} onChange={event=>onFields({...fields,english:event.target.value})}/>:<p lang="en" className="preserve-lines">{fields.english||unspecified}</p>}</section>
    </div>
    <h2>วัตถุประสงค์โครงการ</h2>
    {editing?<><textarea aria-label="วัตถุประสงค์โครงการ" rows={4} value={fields.objectives.join('\n')} onChange={event=>onFields({...fields,objectives:event.target.value.split('\n')})}/><p className="helper-text">แยกแต่ละข้อด้วยการขึ้นบรรทัดใหม่</p></>:fields.objectives.length?<ol className="objectives">{fields.objectives.map((item,index)=><li key={index}>{item}</li>)}</ol>:<p>{unspecified}</p>}
    <div className="keyword-row"><h2>คำค้น / คำสำคัญ</h2>{editing?<input aria-label="คำค้น / คำสำคัญ" placeholder="คั่นแต่ละคำด้วยเครื่องหมาย ," value={fields.keywords.join(',')} onChange={event=>onFields({...fields,keywords:event.target.value.split(',')})}/>:fields.keywords.length?fields.keywords.map((word,index)=><span key={index}>{word}</span>):<span>{unspecified}</span>}</div>
    {editing&&<div className="form-actions"><button className="secondary" type="button" onClick={onCancel}>ยกเลิกการแก้ไข</button><button className="primary" type="submit">บันทึกการแก้ไข</button></div>}
  </form>
}
