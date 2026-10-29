import { useId, useState } from 'react'
import { responsibleSuggestions } from './researchFilters'

export default function ResponsibleSearch({value,names,onChange}:{value:string;names:string[];onChange:(value:string)=>void}) {
  const id=useId()
  const [open,setOpen]=useState(false)
  const [active,setActive]=useState(-1)
  const suggestions=responsibleSuggestions(names,value)
  const expanded=open && Boolean(value.trim())
  const choose=(name:string)=>{onChange(name);setOpen(false);setActive(-1)}
  return <div className="responsible-field">
    <label htmlFor={id}>ผู้รับผิดชอบ</label>
    <div className="responsible-combobox">
      <input id={id} role="combobox" autoComplete="off" aria-autocomplete="list" aria-expanded={expanded} aria-controls={`${id}-options`} aria-describedby={`${id}-hint`} aria-activedescendant={expanded && active>=0 && active<suggestions.length?`${id}-option-${active}`:undefined}
        value={value} placeholder="พิมพ์ชื่อหรือนามสกุล" onFocus={()=>setOpen(true)} onBlur={()=>{setOpen(false);setActive(-1)}}
        onChange={event=>{onChange(event.target.value);setOpen(true);setActive(-1)}}
        onKeyDown={event=>{
          if(event.nativeEvent.isComposing) return
          if(event.key==='Escape'){event.preventDefault();setOpen(false);setActive(-1)}
          else if((event.key==='ArrowDown'||event.key==='ArrowUp') && suggestions.length){
            event.preventDefault();setOpen(true)
            setActive(current=>event.key==='ArrowDown'?(current+1)%suggestions.length:(current<=0?suggestions.length-1:current-1))
          } else if(event.key==='Enter' && expanded && active>=0 && active<suggestions.length){event.preventDefault();choose(suggestions[active])}
        }}/>
      <ul id={`${id}-options`} role="listbox" aria-label="รายชื่อผู้รับผิดชอบที่แนะนำ" hidden={!expanded}>
        {suggestions.map((name,index)=><li id={`${id}-option-${index}`} role="option" aria-selected={active===index} key={name} onMouseDown={event=>event.preventDefault()} onClick={()=>choose(name)}>{name}</li>)}
      </ul>
      {expanded && !suggestions.length && <p className="responsible-no-match" role="status">ไม่พบรายชื่อที่ตรงกัน ลองเปลี่ยนคำค้น</p>}
      <small id={`${id}-hint`}>พิมพ์บางส่วนของชื่อเพื่อดูรายชื่อแนะนำ ใช้ปุ่ม ↑ ↓ และ Enter เพื่อเลือก</small>
    </div>
  </div>
}
