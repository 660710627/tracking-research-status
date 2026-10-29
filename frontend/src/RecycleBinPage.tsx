import { useState } from 'react'
import type { TrashedResearch } from './researchTrash'

export default function ResearchTrash({items,onRestore}:{items:TrashedResearch[];onRestore:(id:number)=>string|null}) {
  const [error,setError]=useState('')
  return <main className="page"><header className="page-header"><div><h1>ถังขยะงานวิจัย</h1><p>กู้คืนโครงการที่ลบไป พร้อมข้อมูลและ ID เดิม</p></div></header>
    <p className="helper-text">ข้อมูลถังขยะเป็นข้อมูลชั่วคราวใน demo และจะเริ่มใหม่เมื่อโหลดหน้าเว็บใหม่</p>
    {error&&<p className="notice" role="alert">{error}</p>}
    <section className="register"><div className="register-head"><div><h2>โครงการที่ถูกลบ</h2><p role="status">ทั้งหมด {items.length} โครงการ</p></div></div>
      {items.length===0?<div className="empty-state"><b>ถังขยะว่าง</b><p>โครงการที่ลบจะปรากฏที่นี่เพื่อให้กู้คืนได้</p></div>:<div className="table-wrap"><table><thead><tr><th>เลขสัญญา</th><th>โครงการ / ผู้รับผิดชอบ</th><th>ลบเมื่อ</th><th>ลบโดยบทบาท</th><th>การจัดการ</th></tr></thead><tbody>{items.map(entry=><tr key={entry.research.id}><td>{entry.research.contract}</td><td><b>{entry.research.title}</b><small>{entry.research.lead}</small></td><td>{new Date(entry.deletedAt).toLocaleString('th-TH',{timeZone:'Asia/Bangkok'})}</td><td>{entry.deletedBy}</td><td><button className="primary" onClick={()=>setError(onRestore(entry.research.id)??'')} aria-label={'กู้คืน '+entry.research.title}>กู้คืนโครงการ</button></td></tr>)}</tbody></table></div>}
    </section></main>
}
