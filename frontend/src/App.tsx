import { useState } from 'react'
import { ResearchListPage, type ResearchChange } from './pages/ResearchListPage'

function App({ change }: { change?: ResearchChange }) {
  const [expanded, setExpanded] = useState(true)
  return (
    <div className="app-shell">
      <a className="skip-link" href="#research-list">ข้ามไปยังรายการงานวิจัย</a>
      <header className="app-header"><span>ระบบติดตามสถานะงานวิจัย</span></header>
      <div className={expanded ? 'workspace' : 'workspace collapsed'}>
        <aside>
          <button aria-expanded={expanded} aria-controls="navigation" onClick={() => setExpanded(!expanded)}>{expanded ? 'ยุบเมนู' : 'เปิดเมนู'}</button>
          <nav id="navigation" aria-label="เมนูหลัก" hidden={!expanded}><a href="#research-list" aria-current="page">รายการงานวิจัย</a></nav>
        </aside>
        <ResearchListPage key={change?.revision ?? 'initial'} change={change} />
      </div>
    </div>
  )
}

export default App
