import { useEffect, useRef } from 'react'

export default function ResearchSelection({checked,mixed=false,disabled=false,label,onChange}:{checked:boolean;mixed?:boolean;disabled?:boolean;label:string;onChange:()=>void}) {
  const ref=useRef<HTMLInputElement>(null)
  useEffect(()=>{if(ref.current) ref.current.indeterminate=mixed},[mixed])
  return <input ref={ref} className="research-checkbox" type="checkbox" checked={checked} disabled={disabled} aria-label={label} onChange={onChange}/>
}
