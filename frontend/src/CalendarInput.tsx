import { fromCalendarDate, toCalendarDate } from './dateInput'

export default function CalendarInput({label,value,onChange,required=false}:{label:string;value:string;onChange:(value:string)=>void;required?:boolean}) {
  return <input type="date" aria-label={label} required={required} value={toCalendarDate(value)}
    onClick={event=>{try{event.currentTarget.showPicker?.()}catch{/* Native calendar icon remains available. */}}}
    onChange={event=>onChange(fromCalendarDate(event.target.value))}/>
}
