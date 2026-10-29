import { useId } from 'react'
import { sdgGoals } from './sdgs'

export default function SdgPicker({value,onChange}:{value:number[];onChange:(ids:number[])=>void}) {
  const hintId=useId()
  return <fieldset className="sdg-picker" aria-describedby={hintId}><legend>เลือกเป้าหมาย SDGs ที่เกี่ยวข้อง *</legend><p id={hintId} className="helper-text">ต้องเลือกอย่างน้อย 1 เป้าหมาย · เลือกได้หลายเป้าหมาย · เลือกแล้ว {value.length} เป้าหมาย</p><div className="sdg-options">{sdgGoals.map(goal=><label key={goal.id} className={value.includes(goal.id)?'sdg-option selected':'sdg-option'}><input type="checkbox" checked={value.includes(goal.id)} onChange={event=>onChange(event.target.checked?[...value,goal.id]:value.filter(id=>id!==goal.id))}/><span><b>SDG {goal.id}</b>{goal.name}</span></label>)}</div></fieldset>
}
