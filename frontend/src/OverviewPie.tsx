import { useState } from 'react'
import { IconChartBar, IconChartPie } from '@tabler/icons-react'
const colors=['#00877b','#80c9bb','#edb65e','#7e91b5','#ae8aba','#dc8b85']
export default function OverviewPie({title,value,unit,segments,money=false}:{title:string;value:string;unit:string;segments:{label:string;value:number}[];money?:boolean}) {
  const [mode,setMode]=useState<'pie'|'bar'>('pie')
  const [selected,setSelected]=useState<number|null>(null)
  const [hovered,setHovered]=useState<number|null>(null)
  const active=hovered??selected
  const detail=active===null?undefined:segments[active]
  const choose=(index:number)=>setSelected(current=>current===index?null:index)
  const total=segments.reduce((sum,segment)=>sum+segment.value,0)
  const max=Math.max(1,...segments.map(segment=>segment.value))
  const stops=segments.map((segment,index)=>{
    const start=total?segments.slice(0,index).reduce((sum,item)=>sum+item.value,0)/total*100:0
    const end=start+(total?segment.value/total*100:0)
    return `${active!==null&&active!==index?'#e4eeec':colors[index%colors.length]} ${start}% ${end}%`
  })
  const format=(number:number)=>money?`${number.toLocaleString('th-TH')} บาท`:`${number.toLocaleString('th-TH')} โครงการ`
  return <article className="overview-pie-card">
    <div className="chart-card-heading"><h2>{title}</h2><div className="chart-mode" role="group" aria-label={`รูปแบบกราฟ ${title}`}><button type="button" aria-pressed={mode==='pie'} onClick={()=>setMode('pie')}><IconChartPie size={16} aria-hidden="true"/>พาย</button><button type="button" aria-pressed={mode==='bar'} onClick={()=>setMode('bar')}><IconChartBar size={16} aria-hidden="true"/>แท่ง</button></div></div>
    <p className="overview-pie-total"><strong>{value}</strong> <span>{unit}</span></p>
    {total>0?<>
      {mode==='pie'?<div className="overview-pie interactive-pie" role="img" aria-label={`${title}: ${segments.map(segment=>`${segment.label} ${format(segment.value)} (${(segment.value/total*100).toFixed(1)}%)`).join(', ')}`} style={{background:`conic-gradient(${stops.join(',')})`}} onMouseMove={event=>{const rect=event.currentTarget.getBoundingClientRect();const angle=(Math.atan2(event.clientX-rect.left-rect.width/2,-(event.clientY-rect.top-rect.height/2))*180/Math.PI+360)%360;const index=segments.findIndex((_,i)=>angle<segments.slice(0,i+1).reduce((sum,s)=>sum+s.value,0)/total*360);setHovered(index<0?null:index)}} onMouseLeave={()=>setHovered(null)} onClick={()=>{if(hovered!==null)choose(hovered)}}></div>:<div className="chart-bars">{segments.map((segment,index)=><button type="button" className="chart-bar-row" key={segment.label} aria-pressed={selected===index} onClick={()=>choose(index)} onMouseEnter={()=>setHovered(index)} onMouseLeave={()=>setHovered(null)} onFocus={()=>setHovered(index)} onBlur={()=>setHovered(null)}><span>{segment.label}</span><span className="chart-bar-track"><span style={{width:`${segment.value/max*100}%`,background:colors[index%colors.length],opacity:active===null||active===index?1:.35}}/></span><small>{format(segment.value)}</small></button>)}<p className="chart-axis-caption">{money?'งบประมาณ (บาท)':'จำนวนโครงการ'} · ความยาวแท่งเทียบกับค่าสูงสุด</p></div>}
      <div className="chart-insight" role="status">{detail?<><b>{detail.label}</b><span>{format(detail.value)} · {(detail.value/total*100).toFixed(1)}% ของทั้งหมด</span></>:<span>ชี้หรือเลือกรายการเพื่อดูรายละเอียด · คลิกซ้ำเพื่อยกเลิก</span>}</div>
      <ul className="overview-pie-legend chart-interactive-legend">{segments.map((segment,index)=><li key={segment.label}><button type="button" aria-pressed={selected===index} onClick={()=>choose(index)} onMouseEnter={()=>setHovered(index)} onMouseLeave={()=>setHovered(null)} onFocus={()=>setHovered(index)} onBlur={()=>setHovered(null)}><span className="pie-key" style={{background:colors[index%colors.length]}} aria-hidden="true"/><span>{segment.label}<small>{format(segment.value)}</small></span><b>{(segment.value/total*100).toFixed(1)}%</b></button></li>)}</ul>
    </>:<p className="overview-pie-empty">ยังไม่มีข้อมูลสำหรับแสดงกราฟ</p>}
  </article>
}
