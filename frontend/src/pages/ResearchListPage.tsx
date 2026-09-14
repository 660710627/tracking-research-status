import { useEffect, useState } from 'react'
import { api, type Research } from '../api'
import { ResearchTable } from '../components/ResearchTable'

export type ResearchChange = { id: number; kind: 'created' | 'updated'; revision: number }

export function ResearchListPage({ change, onCreate }: { change?: ResearchChange; onCreate?: () => void }) {
  const [attempt, setAttempt] = useState(0)
  const [state, setState] = useState<
    { kind: 'loading' } | { kind: 'error' } | { kind: 'success'; items: Research[] }
  >({ kind: 'loading' })
  useEffect(() => {
    let active = true
    void api.listResearches().then((result) => {
      if (!active) return
      if (result.error || !Array.isArray(result.data)) setState({ kind: 'error' })
      else setState({ kind: 'success', items: result.data })
    }).catch(() => { if (active) setState({ kind: 'error' }) })
    return () => { active = false }
  }, [attempt, change])
  function refresh() {
    setState({ kind: 'loading' })
    setAttempt((value) => value + 1)
  }
  const changed = state.kind === 'success' && change
    ? state.items.find((item) => item.id === change.id) : undefined
  return (
    <main id="research-list" tabIndex={-1}>
      <header className="page-heading">
        <div><h1>รายการงานวิจัย</h1><p>ติดตามสถานะและกระบวนการของโครงการทั้งหมด</p></div>
        <div className="form-actions"><button onClick={refresh} disabled={state.kind === 'loading'}>โหลดข้อมูลใหม่</button>{onCreate && <button onClick={onCreate}>เพิ่มงานวิจัย</button>}</div>
      </header>
      <section aria-label="รายการโครงการ" aria-busy={state.kind === 'loading'}>
        <div role="status" aria-live="polite">
          {state.kind === 'loading' && <p className="notice">กำลังโหลดรายการงานวิจัย…</p>}
          {state.kind === 'success' && <p className="count">ทั้งหมด <strong>{state.items.length}</strong> รายการ</p>}
          {changed && <p className="notice success">{change?.kind === 'created' ? 'เพิ่ม' : 'อัปเดต'}โครงการสำเร็จ: {changed.contractNumber}</p>}
        </div>
        {state.kind === 'loading' && <div className="skeleton" aria-hidden="true"><div /><div /><div /></div>}
        {state.kind === 'error' && <div className="notice error" role="alert"><h2>โหลดรายการไม่สำเร็จ</h2><p>โปรดตรวจสอบการเชื่อมต่อ แล้วลองอีกครั้ง</p><button onClick={refresh}>ลองใหม่</button></div>}
        {state.kind === 'success' && (state.items.length === 0
          ? <div className="empty"><h2>ยังไม่มีงานวิจัย</h2><p>เมื่อมีการเพิ่มโครงการ รายการจะแสดงที่นี่</p></div>
          : <ResearchTable items={state.items} changedId={change?.id} />)}
      </section>
    </main>
  )
}
