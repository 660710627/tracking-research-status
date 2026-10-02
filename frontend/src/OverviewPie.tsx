const colors=['#00877b','#80c9bb','#edb65e','#7e91b5','#ae8aba','#dc8b85']
export default function OverviewPie({title,value,unit,segments,money=false}:{title:string;value:string;unit:string;segments:{label:string;value:number}[];money?:boolean}) {
  const total=segments.reduce((sum,segment)=>sum+segment.value,0)
  const stops=segments.map((segment,index)=>{
    const start=total?segments.slice(0,index).reduce((sum,item)=>sum+item.value,0)/total*100:0
    const end=start+(total?segment.value/total*100:0)
    return `${colors[index%colors.length]} ${start}% ${end}%`
  })
  const format=(number:number)=>money?`${number.toLocaleString('th-TH')} บาท`:`${number.toLocaleString('th-TH')} โครงการ`
  return <article className="overview-pie-card">
    <h2>{title}</h2>
    <p className="overview-pie-total"><strong>{value}</strong> <span>{unit}</span></p>
    {total>0?<>
      <div className="overview-pie" role="img" aria-label={`${title}: ${segments.map(segment=>`${segment.label} ${format(segment.value)} (${(segment.value/total*100).toFixed(1)}%)`).join(', ')}`} style={{background:`conic-gradient(${stops.join(',')})`}}/>
      <ul className="overview-pie-legend">{segments.map((segment,index)=><li key={segment.label}><span className="pie-key" style={{background:colors[index%colors.length]}} aria-hidden="true"/><span>{segment.label}<small>{format(segment.value)}</small></span><b>{(segment.value/total*100).toFixed(1)}%</b></li>)}</ul>
    </>:<p className="overview-pie-empty">ยังไม่มีข้อมูลสำหรับแสดงกราฟ</p>}
  </article>
}
