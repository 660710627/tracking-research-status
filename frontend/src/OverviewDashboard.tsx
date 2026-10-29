import { useState } from 'react'
import { IconChartBar, IconChartPie } from '@tabler/icons-react'
import type { Research } from './research'
import { overviewData } from './overviewData'
import type { OverviewGroup } from './overviewData'
import './overviewDashboard.css'
const palette=['#007f73','#d3a323','#e57d35','#c54747','#3c98c4','#84909e']
const groups=[{id:'status',label:'สถานะโครงการ'},{id:'kind',label:'ประเภทโครงการ'},{id:'budget',label:'งบประมาณตามหน่วยงาน'}] as const
const fmt=(value:number)=>value.toLocaleString('th-TH')
export default function OverviewDashboard({items}:{items:Research[]}) {
  const [group,setGroup]=useState<OverviewGroup>('status')
  const [mode,setMode]=useState<'bar'|'pie'>('bar')
  const [selected,setSelected]=useState<string|null>(null)
  const [page,setPage]=useState(0)
  const [mobileView,setMobileView]=useState<'chart'|'details'>('chart')
  const data=overviewData(items,group),unit=group==='budget'?'บาท':'โครงการ'
  const total=data.segments.reduce((sum,item)=>sum+item.value,0),max=Math.max(1,...data.segments.map(item=>item.value))
  const currentPage=Math.min(page,Math.max(0,Math.ceil(data.segments.length/6)-1))
  const visible=data.segments.slice(currentPage*6,currentPage*6+6)
  const selection=data.segments.find(item=>item.label===selected)
  const percent=(value:number)=>total?value/total*100:0
  const gradient=data.segments.map((item,index)=>{const start=data.segments.slice(0,index).reduce((sum,s)=>sum+s.value,0);return `${palette[index%palette.length]} ${percent(start)}% ${percent(start+item.value)}%`}).join(',')
  const choose=(label:string)=>setSelected(selected===label?null:label)
  return <main className={'overview-dashboard mobile-'+mobileView}>
    <header className="od-heading"><div><h1>ภาพรวมงานวิจัย</h1><p>สรุปจากโครงการที่คุณมีสิทธิ์ดู · ไม่รวมถังขยะ</p></div><span>ข้อมูลโครงการในระบบ</span></header>
    <dl className="od-summary"><div><dt>ทั้งหมด</dt><dd>{fmt(data.total)}<small>โครงการ</small></dd></div><div><dt>กำลังดำเนินการรวมขยายเวลา</dt><dd>{fmt(data.active)}<small>โครงการ</small></dd></div><div><dt>เสร็จสิ้น</dt><dd>{fmt(data.completed)}<small>โครงการ</small></dd></div><div><dt>งบประมาณรวม</dt><dd>{fmt(data.budget)}<small>บาท</small></dd></div></dl>
    <div className="od-tools"><div className="od-tabs" role="group" aria-label="มุมมองภาพรวม">{groups.map(option=><button key={option.id} aria-pressed={group===option.id} onClick={()=>{setGroup(option.id);setPage(0);setSelected(null)}}>{option.label}</button>)}</div><div className="od-chart-switch" role="group" aria-label="รูปแบบกราฟ"><button aria-pressed={mode==='bar'} onClick={()=>setMode('bar')}><IconChartBar size={16} aria-hidden="true"/>แท่ง</button><button aria-pressed={mode==='pie'} onClick={()=>setMode('pie')}><IconChartPie size={16} aria-hidden="true"/>พาย</button></div></div>
    <div className="od-mobile-view" role="group" aria-label="ส่วนข้อมูลที่ต้องการแสดง"><button aria-pressed={mobileView==='chart'} onClick={()=>setMobileView('chart')}>ดูกราฟ</button><button aria-pressed={mobileView==='details'} onClick={()=>setMobileView('details')}>ดูตัวเลขและสัดส่วน</button></div>
    <section className="od-report" aria-label={groups.find(option=>option.id===group)?.label}>
      <div className="od-chart"><div className="od-chart-heading"><h2>{groups.find(option=>option.id===group)?.label}</h2><span>{mode==='bar'?`หน่วย: ${unit} · ความยาวแท่งเทียบกับค่าสูงสุด`:'สัดส่วนจากข้อมูลทั้งหมดในมุมมองนี้'}</span></div>
      {!data.total?<div className="od-empty">ยังไม่มีโครงการสำหรับแสดงกราฟ</div>:mode==='bar'?<div className="od-bars">{visible.map((item,index)=><button key={item.label} className="od-bar-row" aria-pressed={selected===item.label} onClick={()=>choose(item.label)}><span className="od-bar-label">{item.label.replace('กำลังดำเนินการ(','').replace(')','')}</span><span className="od-track"><span style={{width:`${item.value/max*100}%`,background:palette[(currentPage*6+index)%palette.length],opacity:selected&&selected!==item.label?0.4:1}}/></span><strong>{fmt(item.value)}</strong></button>)}</div>:<div className="od-donut-area"><div className="od-donut" style={{background:total?`conic-gradient(${gradient})`:'#e5eeeb'}} role="img" aria-label={data.segments.map(item=>`${item.label} ${fmt(item.value)} ${unit}`).join(', ')}><div><strong>{group==='budget'?(data.budget/1000000).toLocaleString('th-TH',{maximumFractionDigits:2}):fmt(data.total)}</strong><span>{group==='budget'?'ล้านบาท':'โครงการ'}</span></div></div></div>}
      </div>
      <aside className="od-breakdown"><header><h2>รายละเอียดประกอบกราฟ</h2><span>{unit} / สัดส่วน</span></header><div className="od-data-list">{visible.map((item,index)=><button key={item.label} aria-pressed={selected===item.label} onClick={()=>choose(item.label)}><i style={{background:palette[(currentPage*6+index)%palette.length]}} aria-hidden="true"/><span>{item.label.replace('กำลังดำเนินการ(','').replace(')','')}{group==='budget'&&<small>{item.count} โครงการ</small>}</span><strong>{fmt(item.value)}<small>{percent(item.value).toFixed(1)}%</small></strong></button>)}</div>{data.segments.length>6&&<label className="od-page-select">เลือกข้อมูลที่แสดง<select value={currentPage} onChange={event=>setPage(Number(event.target.value))}>{Array.from({length:Math.ceil(data.segments.length/6)},(_,index)=><option key={index} value={index}>รายการ {index*6+1}–{Math.min((index+1)*6,data.segments.length)} จาก {data.segments.length}</option>)}</select></label>}</aside>
    </section>
    <footer className="od-detail" role="status">{selection?`${selection.label}: ${fmt(selection.value)} ${unit} · ${percent(selection.value).toFixed(1)}% ของทั้งหมด`:'เลือกมุมมองเพื่อเปรียบเทียบข้อมูล · คลิกแท่งกราฟหรือรายการเพื่อเน้นข้อมูลที่สนใจ'}</footer>
  </main>
}
