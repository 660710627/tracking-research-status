import { useState } from 'react'
import { ResearchListPage, type ResearchChange } from './pages/ResearchListPage'
import { CreateResearchPage } from './pages/CreateResearchPage'

function App({ change }: { change?: ResearchChange }) {
  const [expanded, setExpanded] = useState(true)
  const [creating, setCreating] = useState(false)
  const [latestChange, setLatestChange] = useState<ResearchChange>()
  return (
    <div className="app-shell">
      <a className="skip-link" href={creating ? '#create-research' : '#research-list'}>ข้ามไปยังเนื้อหา</a>
      <header className="app-header"><span>ระบบติดตามสถานะงานวิจัย</span></header>
      <div className={expanded ? 'workspace' : 'workspace collapsed'}>
        <aside>
          <button aria-expanded={expanded} aria-controls="navigation" onClick={() => setExpanded(!expanded)}>{expanded ? 'ยุบเมนู' : 'เปิดเมนู'}</button>
          <nav id="navigation" aria-label="เมนูหลัก" hidden={!expanded}>{creating ? <span>เพิ่มงานวิจัย</span> : <a href="#research-list" aria-current="page">รายการงานวิจัย</a>}</nav>
        </aside>
        {creating ? <CreateResearchPage onCancel={() => setCreating(false)} onCreated={(id) => {
          setLatestChange({ id, kind: 'created', revision: Date.now() })
          setCreating(false)
        }} /> : <ResearchListPage key={(latestChange ?? change)?.revision ?? 'initial'} change={latestChange ?? change} onCreate={() => setCreating(true)} />}
      </div>
    </div>
  )
}

export default App
