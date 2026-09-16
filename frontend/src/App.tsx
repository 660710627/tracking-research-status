import { useMemo, useState } from 'react'

type Role = 'ผู้ประสานงาน' | 'ผู้ดูแลระบบ'
type View = 'list' | 'detail' | 'create' | 'users'
type Research = { id:number; title:string; contract:string; lead:string; unit:string; budget:number; status:string; process:number; kind:string; endDate:string }
const processSteps = ['สัญญาโครงการ','บันทึกข้อตกลง','เปิดบัญชีธนาคาร','การเบิกจ่ายเงิน','การจัดสรรค่าธรรมเนียม','การติดตามส่งรายงาน','รายงานสรุปการใช้เงิน','การปิดบัญชีธนาคาร']
const seed: Research[] = [
  {id:101,title:'การพัฒนาวัสดุดูดซับจากเส้นใยธรรมชาติ',contract:'SURDI-2569-014',lead:'รศ. ดร. กานดา วัฒนศิลป์',unit:'คณะวิทยาศาสตร์',budget:480000,status:'กำลังดำเนินการ',process:5,kind:'โครงการหลัก',endDate:'30 ก.ย. 2570'},
  {id:102,title:'ระบบเฝ้าระวังคุณภาพน้ำด้วยปัญญาประดิษฐ์',contract:'NRCT-2569-088',lead:'ผศ. ดร. นรินทร์ ชูใจ',unit:'คณะวิศวกรรมศาสตร์',budget:1250000,status:'ขยายเวลาครั้งที่ 1',process:6,kind:'โครงการต่อเนื่อง',endDate:'31 มี.ค. 2571'},
  {id:103,title:'ทุนทางวัฒนธรรมกับเศรษฐกิจสร้างสรรค์ชุมชน',contract:'FF-2569-031',lead:'ดร. พิมพ์ชนก ศรีสุข',unit:'คณะอักษรศาสตร์',budget:320000,status:'กำลังดำเนินการ',process:3,kind:'โครงการหลัก',endDate:'30 มิ.ย. 2570'},
  {id:104,title:'ฐานข้อมูลจิตรกรรมฝาผนังภาคกลาง',contract:'SURDI-2568-042',lead:'รศ. ดร. วิภา มณีรัตน์',unit:'คณะโบราณคดี',budget:275000,status:'โครงการเสร็จสิ้น',process:8,kind:'โครงการหลัก',endDate:'31 ธ.ค. 2569'},
]

function Login({onLogin}:{onLogin:(role:Role)=>void}) {
  const [role,setRole]=useState<Role>('ผู้ประสานงาน')
  return <main className="login-page">
    <section className="login-story">
      <span className="eyebrow">RESEARCH OPERATIONS / 2569</span>
      <h1>จากสัญญา<br/>ถึงการปิดโครงการ</h1>
      <p>ภาพรวมเดียวสำหรับติดตามงานวิจัย บุคลากร งบประมาณ และกระบวนการตลอดอายุโครงการ</p>
      <div className="process-preview" aria-label="ตัวอย่างกระบวนการ 8 ขั้น">{processSteps.map((item,index)=><div key={item}><b>{String(index+1).padStart(2,'0')}</b><span>{item}</span></div>)}</div>
    </section>
    <section className="login-panel">
      <div className="prototype-pill">Prototype สำหรับสาธิต</div>
      <p className="institution">สำนักงานบริหารการวิจัย นวัตกรรมและการสร้างสรรค์</p>
      <h2>เข้าสู่พื้นที่ทำงาน</h2>
      <p className="muted">เลือกบทบาทเพื่อดูตัวอย่างสิทธิ์การใช้งาน ข้อมูลทั้งหมดเป็นข้อมูลจำลอง</p>
      <fieldset className="role-picker"><legend>บทบาทสำหรับสาธิต</legend>
        {(['ผู้ประสานงาน','ผู้ดูแลระบบ'] as Role[]).map(item=><label key={item} className={role===item?'role-option selected':'role-option'}><input type="radio" name="role" checked={role===item} onChange={()=>setRole(item)}/><span><b>{item}</b><small>{item==='ผู้ดูแลระบบ'?'ดูบัญชีและข้อมูลพื้นฐานได้':'จัดการและติดตามงานวิจัย'}</small></span></label>)}
      </fieldset>
      <button className="primary full" onClick={()=>onLogin(role)}>เข้าสู่ระบบตัวอย่าง <span>→</span></button>
      <p className="demo-note">ไม่มีการเชื่อมต่อ SSO หรือบันทึกข้อมูลจริง</p>
    </section>
  </main>
}

function Sidebar({role,view,onView,onLogout}:{role:Role;view:View;onView:(v:View)=>void;onLogout:()=>void}) {
  return <aside className="sidebar">
    <div className="brand"><span className="brand-mark">ว</span><div><b>ระบบติดตามงานวิจัย</b><small>Research Operations</small></div></div>
    <nav aria-label="เมนูหลัก">
      <button className={view==='list'||view==='detail'?'active':''} onClick={()=>onView('list')}><span>◫</span>รายการงานวิจัย</button>
      <button className={view==='create'?'active':''} onClick={()=>onView('create')}><span>＋</span>เพิ่มงานวิจัย</button>
      {role==='ผู้ดูแลระบบ'&&<button className={view==='users'?'active':''} onClick={()=>onView('users')}><span>◎</span>บัญชีผู้ใช้งาน</button>}
    </nav>
    <div className="sidebar-foot"><span className="avatar">{role.charAt(0)}</span><div><b>กมลชนก สาธิต</b><small>{role}</small></div><button className="icon-button" aria-label="ออกจากระบบ" onClick={onLogout}>↗</button></div>
  </aside>
}

function Header({title,subtitle}:{title:string;subtitle:string}) {
  return <header className="page-header"><div><span className="eyebrow">ปีงบประมาณ 2569</span><h1>{title}</h1><p>{subtitle}</p></div><div className="header-actions"><span className="prototype-pill">Prototype</span><button className="icon-button" aria-label="การแจ้งเตือน">○</button></div></header>
}

function ResearchList({items,onOpen,onCreate}:{items:Research[];onOpen:(id:number)=>void;onCreate:()=>void}) {
  const [query,setQuery]=useState('')
  const [status,setStatus]=useState('ทั้งหมด')
  const [scenario,setScenario]=useState<'normal'|'loading'|'error'|'empty'>('normal')
  const filtered=useMemo(()=>items.filter(r=>(r.title.includes(query)||r.contract.toLowerCase().includes(query.toLowerCase()))&&(status==='ทั้งหมด'||r.status===status)),[items,query,status])
  const active=items.filter(r=>!r.status.includes('เสร็จสิ้น')).length
  return <main className="page">
    <Header title="ทะเบียนงานวิจัย" subtitle="ติดตามทุกโครงการจากสัญญาถึงการปิดบัญชี"/>
    <section className="ledger-summary" aria-label="สรุปโครงการ"><div><span>โครงการทั้งหมด</span><strong>{items.length}</strong><small>รายการในระบบ</small></div><div><span>กำลังดำเนินการ</span><strong>{active}</strong><small>ต้องติดตามต่อ</small></div><div><span>งบประมาณรวม</span><strong>{(items.reduce((s,r)=>s+r.budget,0)/1000000).toFixed(2)}</strong><small>ล้านบาท</small></div><div className="deadline"><span>ใกล้สิ้นสุด</span><strong>01</strong><small>ภายใน 90 วัน</small></div></section>
    <section className="register">
      <div className="register-head"><div><h2>รายการโครงการ</h2><p>พบ {filtered.length} จาก {items.length} รายการ</p></div><button className="primary" onClick={onCreate}>＋ เพิ่มงานวิจัย</button></div>
      <div className="filters"><label><span className="sr-only">ค้นหางานวิจัย</span><input value={query} onChange={e=>setQuery(e.target.value)} placeholder="ค้นหาชื่อโครงการหรือเลขสัญญา…"/></label><label><span className="sr-only">กรองสถานะ</span><select value={status} onChange={e=>setStatus(e.target.value)}><option>ทั้งหมด</option><option>กำลังดำเนินการ</option><option>ขยายเวลาครั้งที่ 1</option><option>โครงการเสร็จสิ้น</option></select></label><label className="scenario"><span>สถานะสาธิต</span><select value={scenario} onChange={e=>setScenario(e.target.value as typeof scenario)}><option value="normal">ข้อมูลพร้อม</option><option value="loading">กำลังโหลด</option><option value="error">เกิดข้อผิดพลาด</option><option value="empty">ยังไม่มีข้อมูล</option></select></label></div>
      {scenario==='loading'?<div className="state-panel" role="status"><span className="loader"/><b>กำลังโหลดทะเบียนงานวิจัย…</b><p>ระบบกำลังเตรียมข้อมูลล่าสุด</p></div>:scenario==='error'?<div className="state-panel error" role="alert"><b>ไม่สามารถโหลดข้อมูลได้</b><p>การเชื่อมต่อขัดข้อง กรุณาลองอีกครั้ง</p><button className="secondary" onClick={()=>setScenario('normal')}>ลองอีกครั้ง</button></div>:scenario==='empty'?<div className="empty-state"><b>ยังไม่มีงานวิจัยในระบบ</b><p>เริ่มต้นทะเบียนด้วยการเพิ่มโครงการแรก</p><button className="primary" onClick={onCreate}>เพิ่มงานวิจัย</button></div>:filtered.length===0?<div className="empty-state"><b>ไม่พบโครงการที่ตรงกับคำค้น</b><p>ลองเปลี่ยนคำค้นหรือล้างตัวกรอง</p><button onClick={()=>{setQuery('');setStatus('ทั้งหมด')}}>ล้างตัวกรอง</button></div>:
      <div className="table-wrap"><table><thead><tr><th>เลขสัญญา</th><th>โครงการ / ผู้รับผิดชอบ</th><th>สถานะ</th><th>กระบวนการปัจจุบัน</th><th>สิ้นสุด</th><th><span className="sr-only">เปิด</span></th></tr></thead><tbody>{filtered.map(r=><tr key={r.id}><td><b className="contract">{r.contract}</b><small>{r.kind}</small></td><td><button className="title-link" onClick={()=>onOpen(r.id)}>{r.title}</button><small>{r.lead} · {r.unit}</small></td><td><span className={r.status.includes('เสร็จ')?'status done':r.status.includes('ขยาย')?'status warn':'status'}>{r.status}</span></td><td><div className="progress-mini"><span style={{width:String(r.process/8*100)+'%'}}/></div><small>{r.process}/8 · {processSteps[r.process-1]}</small></td><td>{r.endDate}</td><td><button className="row-action" aria-label={'เปิด '+r.title} onClick={()=>onOpen(r.id)}>→</button></td></tr>)}</tbody></table></div>}
    </section>
  </main>
}

function Detail({research,onBack,onUpdate}:{research:Research;onBack:()=>void;onUpdate:(r:Research)=>void}) {
  const terminal=research.status.includes('เสร็จสิ้น')||research.status.includes('ยุติ')
  return <main className="page"><button className="back-link" onClick={onBack}>← กลับทะเบียนงานวิจัย</button><Header title={research.title} subtitle={research.contract+' · '+research.kind}/>
    <div className="detail-grid"><section className="project-sheet"><div className="sheet-label">ข้อมูลโครงการ</div><dl><div><dt>หัวหน้าโครงการ</dt><dd>{research.lead}</dd></div><div><dt>หน่วยงาน</dt><dd>{research.unit}</dd></div><div><dt>งบประมาณ</dt><dd>{research.budget.toLocaleString('th-TH')} บาท</dd></div><div><dt>วันสิ้นสุด</dt><dd>{research.endDate}</dd></div></dl><div className="abstract"><h2>บทคัดย่อ</h2><p>โครงการวิจัยนี้ศึกษาการประยุกต์ใช้องค์ความรู้เพื่อสร้างผลลัพธ์ที่นำไปใช้ประโยชน์ได้จริง โดยเชื่อมโยงนักวิจัย หน่วยงาน และชุมชนอย่างเป็นระบบ</p></div><button className="secondary">เปิดไฟล์สัญญาทุน ↗</button></section>
      <section className="journey"><div className="journey-head"><div><span className="eyebrow">PROJECT JOURNEY</span><h2>เส้นทางโครงการ</h2></div><span className={terminal?'status done':'status'}>{research.status}</span></div><ol>{processSteps.map((item,index)=><li key={item} className={index+1<research.process?'complete':index+1===research.process?'current':''}><span>{index+1<research.process?'✓':String(index+1).padStart(2,'0')}</span><div><b>{item}</b>{index+1===research.process&&<small>ขั้นตอนปัจจุบัน · รอดำเนินการ</small>}</div></li>)}</ol>{!terminal&&<div className="journey-actions"><button className="primary" disabled={research.process===8} onClick={()=>onUpdate({...research,process:Math.min(8,research.process+1)})}>เลื่อนไปขั้นถัดไป</button><button className="secondary" onClick={()=>onUpdate({...research,status:'โครงการเสร็จสิ้น'})}>ปิดโครงการ</button></div>}</section>
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

function FormHeading({number,title,text}:{number:string;title:string;text:string}) {return <div className="form-section"><span>{number}</span><div><h2>{title}</h2><p>{text}</p></div></div>}
function Users(){return <main className="page"><Header title="บัญชีผู้ใช้งาน" subtitle="อนุมัติและดูแลสิทธิ์ของผู้ใช้งานในระบบ"/><section className="register"><div className="register-head"><div><h2>รอการอนุมัติ</h2><p>2 บัญชีต้องตรวจสอบ</p></div><span className="status warn">ADMIN ONLY</span></div><div className="user-list">{['ดร. อภิชาติ ตั้งมั่น|นักวิจัย|คณะวิทยาศาสตร์','นางสาวณัฐชา พูนผล|ผู้ประสานงาน|สำนักงานบริหารการวิจัย'].map(row=>{const [name,role,unit]=row.split('|');return <article key={name}><span className="avatar">{name.charAt(0)}</span><div><b>{name}</b><small>{role} · {unit}</small></div><button className="secondary">ดูข้อมูล</button><button className="primary">อนุมัติ</button></article>})}</div></section></main>}

export default function App(){
  const [role,setRole]=useState<Role|null>(null); const [view,setView]=useState<View>('list'); const [researches,setResearches]=useState(seed); const [selected,setSelected]=useState(101); const [toast,setToast]=useState('')
  if(!role)return <Login onLogin={setRole}/>
  const update=(item:Research)=>{setResearches(items=>items.map(r=>r.id===item.id?item:r));setToast('อัปเดตข้อมูลตัวอย่างแล้ว')}
  return <div className="app-shell"><a className="skip-link" href="#main">ข้ามไปยังเนื้อหา</a><Sidebar role={role} view={view} onView={setView} onLogout={()=>setRole(null)}/><div id="main" className="content">{toast&&<div className="toast" role="status"><b>สำเร็จ</b>{toast}<button aria-label="ปิดข้อความ" onClick={()=>setToast('')}>×</button></div>}{view==='list'&&<ResearchList items={researches} onOpen={id=>{setSelected(id);setView('detail')}} onCreate={()=>setView('create')}/>} {view==='detail'&&<Detail research={researches.find(r=>r.id===selected)??researches[0]} onBack={()=>setView('list')} onUpdate={update}/>} {view==='create'&&<Create onCancel={()=>setView('list')} onSave={r=>{setResearches(items=>[r,...items]);setToast('เพิ่มงานวิจัยตัวอย่างแล้ว');setView('list')}}/>} {view==='users'&&<Users/>}</div></div>
}
