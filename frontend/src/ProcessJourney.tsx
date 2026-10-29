import type { Research } from './research'
import { demoSubprocesses } from './processDemo'
import './processJourney.css'

export default function ProcessJourney({research,steps}:{research:Research;steps:string[]}) {
  const history=research.processHistory??[]
  return <ol className="expandable-journey">{steps.map((step,index)=>{
    const process=index+1
    const records=history.filter(entry=>entry.process===process)
    return <li key={step} className={process<research.process?'complete':process===research.process?'current':''}>
      <span aria-hidden="true">{process<research.process?'✓':String(process).padStart(2,'0')}</span>
      <details className="process-step-details"><summary><b>{step}</b><span className="process-step-chevron" aria-hidden="true">⌄</span>{process===research.process&&<small>ขั้นตอนปัจจุบัน</small>}</summary>
        <div className="process-step-content">
          <p className="process-subheading">กระบวนการย่อย</p>
          <ul>{(demoSubprocesses[process]??[]).map(option=>{
            const active=process===research.process&&research.processDetails?.subprocess===option.name
            const recorded=records.some(entry=>entry.subprocess===option.name)
            return <li key={option.name} className={active?'subprocess-current':recorded?'subprocess-recorded':''}>{option.name}<small>{active?'ปัจจุบัน':recorded?'มีประวัติการบันทึก':'ยังไม่มีข้อมูลที่บันทึก'}</small></li>
          })}</ul>
        </div>
      </details>
    </li>
  })}</ol>
}
