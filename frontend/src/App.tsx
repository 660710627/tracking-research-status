import { useMemo, useState } from 'react'
import { canManageResearch, changeResearch, isTerminal, statuses } from './research'
import type { Research, ResearchAction, Role } from './research'
import { ConfirmResearch, EditResearch, StatusPanel } from './ResearchControls'
import AdminPreview from './AdminPreview'
import { adminModules } from './adminModules'
import type { AdminView } from './adminModules'

type View = 'list' | 'detail' | 'create' | AdminView
const processSteps = ['สัญญาโครงการ','บันทึกข้อตกลง','เปิดบัญชีธนาคาร','การเบิกจ่ายเงิน','การจัดสรรค่าธรรมเนียม','การติดตามส่งรายงาน','รายงานสรุปการใช้เงิน','การปิดบัญชีธนาคาร']
const seed: Research[] = [
  {id:101,title:'การพัฒนาวัสดุดูดซับจากเส้นใยธรรมชาติ',contract:'SURDI-2569-014',lead:'รศ. ดร. กานดา วัฒนศิลป์',unit:'คณะวิทยาศาสตร์',budget:480000,status:'กำลังดำเนินการ',process:5,kind:'โครงการหลัก',endDate:'30 ก.ย. 2570'},
  {id:102,title:'ระบบเฝ้าระวังคุณภาพน้ำด้วยปัญญาประดิษฐ์',contract:'NRCT-2569-088',lead:'ผศ. ดร. นรินทร์ ชูใจ',unit:'คณะวิศวกรรมศาสตร์',budget:1250000,status:statuses[1],process:6,kind:'โครงการต่อเนื่อง',endDate:'31 มี.ค. 2571'},
  {id:103,title:'ทุนทางวัฒนธรรมกับเศรษฐกิจสร้างสรรค์ชุมชน',contract:'FF-2569-031',lead:'ดร. พิมพ์ชนก ศรีสุข',unit:'คณะอักษรศาสตร์',budget:320000,status:'กำลังดำเนินการ',process:3,kind:'โครงการหลัก',endDate:'30 มิ.ย. 2570'},
  {id:104,title:'ฐานข้อมูลจิตรกรรมฝาผนังภาคกลาง',contract:'SURDI-2568-042',lead:'รศ. ดร. วิภา มณีรัตน์',unit:'คณะโบราณคดี',budget:275000,status:'โครงการเสร็จสิ้น',process:8,kind:'โครงการหลัก',endDate:'31 ธ.ค. 2569'},
]

function Login({onLogin}:{onLogin:(role:Role)=>void}) {
  const [role,setRole]=useState<Role>('ผู้ประสานงาน')
  return <main className="login-page">
    <section className="login-story">
      <div className="login-identity"><span className="brand-mark" aria-hidden="true">ว</span><span>ระบบติดตามสถานะงานวิจัย</span></div>
      <h1>ทุกโครงการ<br/>เห็นความคืบหน้า</h1>
      <p>เปิดแฟ้มโครงการ ตรวจสถานะ และติดตามขั้นตอนที่ต้องดำเนินการต่อในพื้นที่เดียว</p>
      <p className="route-caption">กระบวนการติดตามโครงการ</p><div className="process-preview" aria-label="ตัวอย่างกระบวนการ 8 ขั้น">{processSteps.map((item,index)=><div key={item}><b>{String(index+1).padStart(2,'0')}</b><span>{item}</span></div>)}</div>
    </section>
    <section className="login-panel">
      <div className="prototype-pill">Prototype สำหรับสาธิต</div>
      <p className="institution">สำนักงานบริหารการวิจัย นวัตกรรมและการสร้างสรรค์</p>
      <h2>เข้าสู่พื้นที่ทำงาน</h2>
      <p className="muted">เลือกบทบาทเพื่อดูตัวอย่างสิทธิ์การใช้งาน ข้อมูลทั้งหมดเป็นข้อมูลจำลอง</p>
      <fieldset className="role-picker"><legend>บทบาทสำหรับสาธิต</legend>
        {(['นักวิจัย','ผู้ประสานงาน','ผู้ดูแลระบบ'] as Role[]).map(item=><label key={item} className={role===item?'role-option selected':'role-option'}><input type="radio" name="role" checked={role===item} onChange={()=>setRole(item)}/><span><b>{item}</b><small>{item==='นักวิจัย'?'ดูรายการ รายละเอียด และติดตามความคืบหน้า':item==='ผู้ดูแลระบบ'?'ดูบัญชีและข้อมูลพื้นฐานได้':'จัดการและติดตามงานวิจัย'}</small></span></label>)}
      </fieldset>
      <button className="primary full" onClick={()=>onLogin(role)}>เข้าสู่ระบบตัวอย่าง</button>
      <p className="demo-note">ไม่มีการเชื่อมต่อ SSO หรือบันทึกข้อมูลจริง</p>
    </section>
  </main>
}

function Sidebar({role,view,onView,onLogout}:{role:Role;view:View;onView:(v:View)=>void;onLogout:()=>void}) {
  const [adminExpanded,setAdminExpanded]=useState(true)
  return <aside className="sidebar">
    <div className="brand"><span className="brand-mark">ว</span><div><b>ระบบติดตามงานวิจัย</b><small>สำนักงานบริหารการวิจัย</small></div></div>
    <nav aria-label="เมนูหลัก">
      <button className={view==='list'||view==='detail'?'active':''} onClick={()=>onView('list')}><span>◫</span>รายการงานวิจัย</button>
      {role!=='นักวิจัย'&&<button className={view==='create'?'active':''} onClick={()=>onView('create')}><span>＋</span>เพิ่มงานวิจัย</button>}
      {role==='ผู้ดูแลระบบ'&&<div className="admin-navigation"><button className="admin-toggle" aria-expanded={adminExpanded} aria-controls="admin-menu" onClick={()=>setAdminExpanded(!adminExpanded)}>ระบบจัดการข้อมูลพื้นฐาน <span aria-hidden="true">{adminExpanded?'−':'＋'}</span></button><div id="admin-menu" hidden={!adminExpanded}>{adminModules.map(module=><button key={module.id} className={view===module.id?'active':''} aria-current={view===module.id?'page':undefined} onClick={()=>onView(module.id)}>{module.title}</button>)}</div></div>}
    </nav>
    <div className="sidebar-foot"><span className="avatar">{role.charAt(0)}</span><div><b>กมลชนก สาธิต</b><small>{role}</small></div><button className="icon-button" aria-label="ออกจากระบบ" onClick={onLogout}>↗</button></div>
  </aside>
}

function Header({title,subtitle}:{title:string;subtitle:string}) {
  return <header className="page-header"><div><h1>{title}</h1><p>{subtitle}</p></div><div className="header-actions"><span className="prototype-pill">Prototype</span></div></header>
}

function ResearchList({items,onOpen,onCreate,canManage}:{canManage:boolean;items:Research[];onOpen:(id:number)=>void;onCreate:()=>void}) {
  const [query,setQuery]=useState('')
  const [status,setStatus]=useState('ทั้งหมด')
  const [scenario,setScenario]=useState<'normal'|'loading'|'error'|'empty'>('normal')
  const filtered=useMemo(()=>items.filter(r=>(r.title.includes(query)||r.contract.toLowerCase().includes(query.toLowerCase()))&&(status==='ทั้งหมด'||r.status===status)),[items,query,status])
  const active=items.filter(r=>!isTerminal(r.status)).length
  return <main className="page">
    <Header title="ทะเบียนงานวิจัย" subtitle="ติดตามทุกโครงการจากสัญญาถึงการปิดบัญชี"/>
    <section className="ledger-summary" aria-label="สรุปโครงการ"><div><span>โครงการทั้งหมด</span><strong>{items.length}</strong><small>รายการในระบบ</small></div><div><span>กำลังดำเนินการ</span><strong>{active}</strong><small>ต้องติดตามต่อ</small></div><div><span>งบประมาณรวม</span><strong>{(items.reduce((s,r)=>s+r.budget,0)/1000000).toFixed(2)}</strong><small>ล้านบาท</small></div><div className="deadline"><span>เสร็จสิ้นแล้ว</span><strong>{items.filter(item=>item.status==='โครงการเสร็จสิ้น').length}</strong><small>โครงการ</small></div></section>
    <section className="register">
      <div className="register-head"><div><h2>รายการโครงการ</h2><p>พบ {filtered.length} จาก {items.length} รายการ</p></div>{canManage&&<button className="primary" onClick={onCreate}>＋ เพิ่มงานวิจัย</button>}</div>
      <div className="filters"><label><span className="sr-only">ค้นหางานวิจัย</span><input value={query} onChange={e=>setQuery(e.target.value)} placeholder="ค้นหาชื่อโครงการหรือเลขสัญญา…"/></label><label><span className="sr-only">กรองสถานะ</span><select value={status} onChange={e=>setStatus(e.target.value)}><option>ทั้งหมด</option>{statuses.map(item=><option key={item}>{item}</option>)}</select></label><label className="scenario"><span>สถานะสาธิต</span><select value={scenario} onChange={e=>setScenario(e.target.value as typeof scenario)}><option value="normal">ข้อมูลพร้อม</option><option value="loading">กำลังโหลด</option><option value="error">เกิดข้อผิดพลาด</option><option value="empty">ยังไม่มีข้อมูล</option></select></label></div>
      {scenario==='loading'?<div className="state-panel" role="status"><span className="loader"/><b>กำลังโหลดทะเบียนงานวิจัย…</b><p>ระบบกำลังเตรียมข้อมูลล่าสุด</p></div>:scenario==='error'?<div className="state-panel error" role="alert"><b>ไม่สามารถโหลดข้อมูลได้</b><p>การเชื่อมต่อขัดข้อง กรุณาลองอีกครั้ง</p><button className="secondary" onClick={()=>setScenario('normal')}>ลองอีกครั้ง</button></div>:scenario==='empty'?<div className="empty-state"><b>ยังไม่มีงานวิจัยในระบบ</b><p>{canManage?'เริ่มต้นทะเบียนด้วยการเพิ่มโครงการแรก':'เมื่อผู้ประสานงานเพิ่มโครงการแล้ว รายการจะแสดงที่นี่'}</p>{canManage&&<button className="primary" onClick={onCreate}>เพิ่มงานวิจัย</button>}</div>:filtered.length===0?<div className="empty-state"><b>ไม่พบโครงการที่ตรงกับคำค้น</b><p>ลองเปลี่ยนคำค้นหรือล้างตัวกรอง</p><button onClick={()=>{setQuery('');setStatus('ทั้งหมด')}}>ล้างตัวกรอง</button></div>:
      <div className="table-wrap"><table><thead><tr><th>เลขสัญญา</th><th>โครงการ / ผู้รับผิดชอบ</th><th>สถานะ</th><th>กระบวนการปัจจุบัน</th><th>สิ้นสุด</th><th><span className="sr-only">เปิด</span></th></tr></thead><tbody>{filtered.map(r=><tr key={r.id}><td><b className="contract">{r.contract}</b><small>{r.kind}</small></td><td><button className="title-link" onClick={()=>onOpen(r.id)}>{r.title}</button><small>{r.lead} · {r.unit}</small></td><td><span className={isTerminal(r.status)?'status done':r.status.includes('ขยาย')?'status warn':'status'}>{r.status}</span></td><td><div className="progress-mini"><span style={{width:String(r.process/8*100)+'%'}}/></div><small>{r.process}/8 · {processSteps[r.process-1]}</small></td><td>{r.endDate}</td><td><button className="row-action" aria-label={'เปิด '+r.title} onClick={()=>onOpen(r.id)}>→</button></td></tr>)}</tbody></table></div>}
    </section>
  </main>
}

const detailSamples: Record<number, { fundingType:string; fund:string; start:string; collaborators:string; thai:string; english:string; objectives:string[]; keywords:string[] }> = {
  101: {fundingType:'ภายใน',fund:'ทุนอุดหนุนการวิจัย มหาวิทยาลัยศิลปากร',start:'1 ต.ค. 2569',collaborators:'ดร. ปราณี ใจดี; ผศ. ดร. ธนา วงศ์วิจัย',thai:'ศึกษาการใช้เส้นใยจากวัสดุเหลือใช้ทางการเกษตรเพื่อผลิตวัสดุดูดซับสารปนเปื้อนในน้ำ เปรียบเทียบวิธีเตรียมเส้นใยและประสิทธิภาพการดูดซับในห้องปฏิบัติการ พร้อมประเมินการนำกลับมาใช้ซ้ำและต้นทุนเบื้องต้น เพื่อพัฒนาแนวทางใช้ประโยชน์จากทรัพยากรท้องถิ่น',english:'This study investigates agricultural waste fibers as adsorbents for water treatment. Preparation methods, adsorption performance, reusability, and preliminary costs are compared to support the practical use of local resources.',objectives:['พัฒนาวัสดุดูดซับจากเส้นใยธรรมชาติ','เปรียบเทียบประสิทธิภาพการดูดซับและการใช้ซ้ำ','ประเมินต้นทุนการผลิตระดับห้องปฏิบัติการ'],keywords:['เส้นใยธรรมชาติ','วัสดุดูดซับ','การบำบัดน้ำ']},
  102: {fundingType:'ภายนอก',fund:'สำนักงานการวิจัยแห่งชาติ (ข้อมูลสาธิต)',start:'1 เม.ย. 2569',collaborators:'ดร. สุธี รักษ์น้ำ; ดร. มาลี พัฒนกิจ',thai:'พัฒนาต้นแบบระบบติดตามคุณภาพน้ำโดยรวบรวมค่าจากเซนเซอร์และวิเคราะห์ด้วยแบบจำลองปัญญาประดิษฐ์ ศึกษาความแม่นยำในการตรวจหาค่าผิดปกติและเปรียบเทียบกับผลตรวจวัดมาตรฐาน เพื่อสนับสนุนการเฝ้าระวังคุณภาพน้ำในพื้นที่ศึกษา',english:'This project develops a water-quality monitoring prototype combining sensor measurements with artificial intelligence. Anomaly detection is evaluated against reference measurements to support monitoring in the study area.',objectives:['พัฒนาต้นแบบระบบเก็บข้อมูลคุณภาพน้ำ','ประเมินแบบจำลองตรวจหาค่าผิดปกติ','ทดสอบการใช้งานร่วมกับหน่วยงานในพื้นที่'],keywords:['คุณภาพน้ำ','ปัญญาประดิษฐ์','เซนเซอร์']},
  103: {fundingType:'ภายใน',fund:'ทุนสนับสนุนงานวิจัยพื้นฐาน (ข้อมูลสาธิต)',start:'1 ก.ค. 2569',collaborators:'ดร. วรางคณา ศิลป์สกุล',thai:'ศึกษาทุนทางวัฒนธรรมของชุมชนผ่านการสัมภาษณ์ การสำรวจ และกระบวนการมีส่วนร่วม เพื่อรวบรวมองค์ความรู้ท้องถิ่นและวิเคราะห์แนวทางพัฒนาผลิตภัณฑ์สร้างสรรค์ โดยคำนึงถึงอัตลักษณ์และความต้องการของคนในชุมชน',english:'This study explores community cultural assets through interviews, surveys, and participatory activities. Local knowledge informs creative product development while preserving community identity and addressing local needs.',objectives:['จัดทำข้อมูลทุนทางวัฒนธรรมชุมชน','วิเคราะห์โอกาสพัฒนาผลิตภัณฑ์สร้างสรรค์','เสนอแนวทางใช้ประโยชน์ร่วมกับชุมชน'],keywords:['ทุนทางวัฒนธรรม','เศรษฐกิจสร้างสรรค์','ชุมชน']},
  104: {fundingType:'ภายใน',fund:'ทุนอุดหนุนการวิจัย มหาวิทยาลัยศิลปากร',start:'1 ม.ค. 2568',collaborators:'ผศ. ดร. อรุณ อนุรักษ์',thai:'รวบรวมและจัดหมวดหมู่ข้อมูลจิตรกรรมฝาผนังในพื้นที่ภาคกลาง โดยบันทึกภาพ รายละเอียดแหล่งที่ตั้ง และลักษณะทางศิลปกรรม จัดทำต้นแบบฐานข้อมูลเพื่อสนับสนุนการศึกษาและการอนุรักษ์มรดกทางวัฒนธรรม',english:'This project documents and classifies mural paintings in central Thailand. Images, locations, and artistic characteristics are organized into a prototype database for research and cultural heritage conservation.',objectives:['สำรวจและบันทึกจิตรกรรมฝาผนัง','จัดหมวดหมู่ข้อมูลทางศิลปกรรม','พัฒนาฐานข้อมูลสำหรับการศึกษาและอนุรักษ์'],keywords:['จิตรกรรมฝาผนัง','ฐานข้อมูล','มรดกวัฒนธรรม']},
}

function Detail({research,onBack,onChange,canManage}:{canManage:boolean;research:Research;onBack:()=>void;onChange:(action:ResearchAction)=>string|null}) {
  const [pending,setPending]=useState<Exclude<ResearchAction,{type:'edit'}>|null>(null)
  const [editing,setEditing]=useState(false)
  const [error,setError]=useState('')
  const sample=detailSamples[research.id]
  const unspecified='ยังไม่ได้ระบุ'
  const terminal=isTerminal(research.status)
  return <main className="page"><button className="back-link" onClick={onBack}>← กลับทะเบียนงานวิจัย</button><Header title={research.title} subtitle={research.contract+' · '+research.kind}/>
    {canManage&&<div className="research-toolbar"><button className="secondary" onClick={()=>setEditing(true)}>แก้ไขงานวิจัย</button><button className="danger-outline" onClick={()=>setPending({type:'delete'})}>ลบงานวิจัย</button></div>}
    {error&&<p className="notice" role="alert">{error}</p>}
    {editing&&<EditResearch research={research} onCancel={()=>setEditing(false)} onSave={fields=>{const message=onChange({type:'edit',fields});if(!message)setEditing(false);return message}}/>}
    {pending&&<ConfirmResearch research={research} action={pending} onCancel={()=>setPending(null)} onConfirm={()=>{const message=onChange(pending);setPending(null);setError(message??'')}}/>}
    <div className="detail-grid research-detail"><section className="project-sheet research-sheet" aria-label="รายละเอียดงานวิจัย">
      <p className="detail-demo">ข้อมูลประกอบตัวอย่างสำหรับสาธิต</p>
      <h2>ข้อมูลโครงการ</h2>
      <dl className="facts-grid">
        <div className="wide"><dt>ชื่อโครงการ</dt><dd>{research.title}</dd></div>
        <div><dt>ประเภทโครงการ</dt><dd>งานวิจัย</dd></div>
        <div><dt>ทุนอุดหนุน / ประเภทที่เกี่ยวข้อง</dt><dd>{sample?'ทุนอุดหนุนการวิจัย':unspecified}</dd></div>
        <div><dt>สถานะการดำเนินงาน</dt><dd>{research.kind==='โครงการต่อเนื่อง'?'โครงการต่อเนื่อง':'โครงการในงบประมาณ'}</dd></div>
        <div><dt>หน่วยงานรับผิดชอบโครงการ</dt><dd>{research.unit}</dd></div>
        <div><dt>หน่วยงานรับผิดชอบงบประมาณ</dt><dd>{sample?research.unit:unspecified}</dd></div>
        <div><dt>ระยะเวลาวิจัย เริ่ม–สิ้นสุด</dt><dd>{sample?.start??unspecified} – {research.endDate}</dd></div>
        <div><dt>งบประมาณโครงการ</dt><dd className="budget-value">{research.budget.toLocaleString('th-TH',{minimumFractionDigits:2})} บาท</dd></div>
      </dl>
      <h2>บุคลากรและแหล่งทุน</h2>
      <dl className="facts-grid">
        <div><dt>หัวหน้าโครงการ</dt><dd>{research.lead}</dd></div>
        <div><dt>ผู้ร่วมโครงการ</dt><dd>{sample?.collaborators??unspecified}</dd></div>
        <div><dt>ประเภทแหล่งทุน</dt><dd>{sample?.fundingType??unspecified}</dd></div>
        <div><dt>ชื่อแหล่งทุน</dt><dd>{sample?.fund??unspecified}</dd></div>
        <div><dt>เลขที่สัญญาทุน</dt><dd>{research.contract}</dd></div>
        <div><dt>ไฟล์สัญญารับทุน</dt><dd className="muted">ยังไม่มีไฟล์แนบใน mockup</dd></div>
      </dl>
      <div className="research-texts">
        <section><h2>บทคัดย่อ (ไทย)</h2><p>{sample?.thai??unspecified}</p></section>
        <section><h2>บทคัดย่อ (อังกฤษ)</h2><p lang="en">{sample?.english??unspecified}</p></section>
      </div>
      <h2>วัตถุประสงค์โครงการ</h2>
      {sample?<ol className="objectives">{sample.objectives.map(item=><li key={item}>{item}</li>)}</ol>:<p>{unspecified}</p>}
      <div className="keyword-row"><h2>คำค้น / คำสำคัญ</h2>{sample?sample.keywords.map(word=><span key={word}>{word}</span>):<span>{unspecified}</span>}</div>
    </section>
      <aside className="research-tracking"><StatusPanel key={research.status} research={research} canManage={canManage} onRequest={action=>{if(action.type!=='edit')setPending(action)}}/>
      <section className="journey"><div className="journey-head"><div><span className="journey-count">ขั้นตอน {research.process} จาก 8</span><h2>กระบวนการงานวิจัย</h2></div></div><ol>{processSteps.map((item,index)=><li key={item} className={index+1<research.process?'complete':index+1===research.process?'current':''}><span>{index+1<research.process?'✓':String(index+1).padStart(2,'0')}</span><div><b>{item}</b>{index+1===research.process&&<small>{terminal?'ขั้นตอนเมื่อสิ้นสุดโครงการ':'ขั้นตอนปัจจุบัน'}</small>}</div></li>)}</ol>{canManage&&!terminal&&<div className="journey-actions"><button className="primary" disabled={research.process===8} onClick={()=>setPending({type:'process',process:research.process+1})}>เลื่อนไปขั้นถัดไป</button>{research.process===8&&<p className="helper-text">ถึงกระบวนการสุดท้ายแล้ว ปรับสถานะเป็นโครงการเสร็จสิ้นได้ในส่วนสถานะงานวิจัย</p>}</div>}</section></aside>
    </div></main>
}

function Create({onCancel,onSave}:{onCancel:()=>void;onSave:(r:Research)=>void}) {
  const [title,setTitle]=useState(''); const [contract,setContract]=useState(''); const [lead,setLead]=useState(''); const [error,setError]=useState('')
  const save=()=>{if(!title.trim()||!contract.trim()||!lead.trim()){setError('กรอกชื่อโครงการ เลขสัญญา และหัวหน้าโครงการให้ครบ');return}onSave({id:Date.now(),title,contract,lead,unit:'คณะวิทยาศาสตร์',budget:350000,status:'กำลังดำเนินการ',process:1,kind:'โครงการหลัก',endDate:'30 ก.ย. 2571'})}
  return <main className="page"><button className="back-link" onClick={onCancel}>← กลับทะเบียนงานวิจัย</button><Header title="เพิ่มงานวิจัย" subtitle="บันทึกข้อมูลสำคัญสำหรับเริ่มติดตามโครงการ"/>{error&&<div className="notice" role="alert">{error}</div>}<section className="form-sheet">
    <FormHeading number="01" title="ข้อมูลโครงการ" text="ชื่อและประเภทของโครงการ"/><div className="form-grid"><label className="wide">ชื่อโครงการ *<input value={title} onChange={e=>setTitle(e.target.value)}/></label><label>ประเภทโครงการ<select><option>งานวิจัย</option><option>บริการวิชาการ</option></select></label><label>ลักษณะโครงการ<select><option>โครงการในงบประมาณ</option><option>โครงการต่อเนื่อง</option></select></label></div>
    <FormHeading number="02" title="บุคลากรและสัญญา" text="ผู้รับผิดชอบและเอกสารอ้างอิง"/><div className="form-grid"><label>หัวหน้าโครงการ *<input value={lead} onChange={e=>setLead(e.target.value)}/></label><label>ผู้ร่วมโครงการ<input defaultValue="ดร. ปราณี ใจดี"/></label><label>เลขที่สัญญาทุน *<input value={contract} onChange={e=>setContract(e.target.value)}/></label><label>ไฟล์สัญญา PDF<input type="file" accept=".pdf"/></label></div>
    <FormHeading number="03" title="ระยะเวลาและงบประมาณ" text="ข้อมูลสำหรับการติดตาม"/><div className="form-grid"><label>วันเริ่มต้น<input defaultValue="01/10/2569"/></label><label>วันสิ้นสุด<input defaultValue="30/09/2571"/></label><label>งบประมาณ (บาท)<input defaultValue="350,000"/></label><label>หน่วยงาน<select><option>คณะวิทยาศาสตร์</option><option>คณะวิศวกรรมศาสตร์</option></select></label></div>
    <div className="form-actions"><button className="secondary" onClick={onCancel}>ยกเลิก</button><button className="primary" onClick={save}>บันทึกงานวิจัย</button></div></section></main>
}

function FormHeading({number,title,text}:{number:string;title:string;text:string}) {return <div className="form-section"><span className="section-marker" aria-hidden="true">{number === "01" ? "▤" : number === "02" ? "◎" : "◷"}</span><div><h2>{title}</h2><p>{text}</p></div></div>}
export default function App(){
  const [role,setRole]=useState<Role|null>(null)
  const [view,setView]=useState<View>('list')
  const [researches,setResearches]=useState(seed)
  const [selected,setSelected]=useState(101)
  const [toast,setToast]=useState('')
  if(!role)return <Login onLogin={nextRole=>{setRole(nextRole);setView('list');setToast('')}}/>
  const canManage=canManageResearch(role)
  const research=researches.find(item=>item.id===selected)
  const change=(action:ResearchAction):string|null=>{
    try {
      const next=changeResearch(researches,selected,role,action)
      setResearches(next)
      setToast(action.type==='delete'?'ลบงานวิจัยตัวอย่างแล้ว':action.type==='edit'?'บันทึกการแก้ไขแล้ว':action.type==='status'?'ปรับสถานะงานวิจัยแล้ว':'ปรับกระบวนการแล้ว')
      if(action.type==='delete')setView('list')
      return null
    } catch(error) { return error instanceof Error?error.message:'ไม่สามารถบันทึกข้อมูลได้ กรุณาลองอีกครั้ง' }
  }
  const navigate=(next:View)=>{setView(next);setToast('')}
  return <div className="app-shell"><a className="skip-link" href="#main">ข้ามไปยังเนื้อหา</a><Sidebar role={role} view={view} onView={navigate} onLogout={()=>setRole(null)}/><div id="main" tabIndex={-1} className="content">
    {toast&&<div className="toast" role="status"><b>สำเร็จ</b>{toast}<button aria-label="ปิดข้อความ" onClick={()=>setToast('')}>×</button></div>}
    {view==='list'&&<ResearchList canManage={canManage} items={researches} onOpen={id=>{setSelected(id);navigate('detail')}} onCreate={()=>navigate('create')}/>}
    {view==='detail'&&research&&<Detail key={research.id} canManage={canManage} research={research} onBack={()=>navigate('list')} onChange={change}/>}
    {canManage&&view==='create'&&<Create onCancel={()=>navigate('list')} onSave={r=>{setResearches(items=>[r,...items]);setToast('เพิ่มงานวิจัยตัวอย่างแล้ว');setView('list')}}/>}
    {role==='ผู้ดูแลระบบ'&&adminModules.some(module=>module.id===view)&&<AdminPreview key={view} view={view as AdminView}/>}
  </div></div>
}
