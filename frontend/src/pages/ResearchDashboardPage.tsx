import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { listResearches, type Research } from '../api/client'
import { getAPIErrorCode } from '../api/errors'
import { CreateResearchDialog } from '../components/CreateResearchDialog'
import { DeleteResearchDialog } from '../components/DeleteResearchDialog'
import { EditResearchDialog } from '../components/EditResearchDialog'

type LoadState = { kind: 'loading' } | { kind: 'error'; message: string } | { kind: 'ready'; researches: Research[] }
type Notice = { tone: 'success' | 'error'; message: string }
const EMPTY: Research[] = []
const terminalStatuses = new Set(['โครงการเสร็จสิ้น', 'ยุติโครงการ'])

export function ResearchDashboardPage() {
  const [state, setState] = useState<LoadState>({ kind: 'loading' })
  const [loadKey, setLoadKey] = useState(0)
  const [query, setQuery] = useState('')
  const [status, setStatus] = useState('')
  const [process, setProcess] = useState('')
  const [notice, setNotice] = useState<Notice>()
  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<Research>()
  const [deleting, setDeleting] = useState<Research>()
  const [changedId, setChangedId] = useState<number>()
  const createButton = useRef<HTMLButtonElement>(null)
  const editButton = useRef<HTMLButtonElement>(null)
  const deleteButton = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    const controller = new AbortController()
    async function load() {
      setState({ kind: 'loading' })
      try {
        const result = await listResearches({ signal: controller.signal })
        if (result.error) setState({ kind: 'error', message: errorMessage(result.error) })
        else setState({ kind: 'ready', researches: result.data })
      } catch (error) {
        if (!controller.signal.aborted) setState({ kind: 'error', message: errorMessage(error) })
      }
    }
    void load()
    return () => controller.abort()
  }, [loadKey])

  useEffect(() => {
    if (!notice) return
    const timeout = window.setTimeout(() => setNotice(undefined), 5000)
    return () => window.clearTimeout(timeout)
  }, [notice])

  const researches = state.kind === 'ready' ? state.researches : EMPTY
  const filtered = useMemo(() => {
    const term = query.trim().toLocaleLowerCase('th-TH')
    return researches.filter((research) => {
      const matchesText = term === '' || [research.id, research.title, research.description, research.status, research.process]
        .some((value) => String(value).toLocaleLowerCase('th-TH').includes(term))
      return matchesText && (status === '' || research.status === status) && (process === '' || research.process === process)
    })
  }, [process, query, researches, status])
  const statuses = [...new Set(researches.map((research) => research.status))]
  const processes = [...new Set(researches.map((research) => research.process))]
  const active = researches.filter((research) => !terminalStatuses.has(research.status)).length

  const openCreate = (button: HTMLButtonElement) => { createButton.current = button; setCreateOpen(true) }
  const closeCreate = () => { setCreateOpen(false); window.requestAnimationFrame(() => createButton.current?.focus()) }
  const closeEdit = () => { setEditing(undefined); window.requestAnimationFrame(() => editButton.current?.focus()) }
  const closeDelete = () => { setDeleting(undefined); window.requestAnimationFrame(() => deleteButton.current?.focus()) }
  const retry = useCallback(() => setLoadKey((key) => key + 1), [])
  const resetFilters = () => { setQuery(''); setStatus(''); setProcess('') }

  function updated(research: Research) {
    setEditing(undefined); setChangedId(research.id); setNotice({ tone: 'success', message: `บันทึก “${research.title}” แล้ว` })
    setState((current) => current.kind === 'ready' ? { kind: 'ready', researches: current.researches.map((item) => item.id === research.id ? research : item) } : current)
  }
  function created(research: Research) { setCreateOpen(false); setChangedId(research.id); setNotice({ tone: 'success', message: `สร้าง “${research.title}” แล้ว` }); retry() }
  function deleted(research: Research) {
    setDeleting(undefined); setNotice({ tone: 'success', message: `ลบ “${research.title}” แล้ว` })
    setState((current) => current.kind === 'ready' ? { kind: 'ready', researches: current.researches.filter((item) => item.id !== research.id) } : current)
  }
  function notFound() { setEditing(undefined); setDeleting(undefined); setNotice({ tone: 'error', message: 'ข้อมูลรายการเปลี่ยนไปแล้ว กำลังโหลดทะเบียนล่าสุด' }); retry() }

  return <div className="min-h-screen bg-[#f3f7ff] text-slate-950">
    <div className="mx-auto grid min-h-screen w-[1440px] grid-cols-[264px_minmax(0,1fr)] bg-white shadow-[0_0_0_1px_#dbeafe]">
      <aside className="flex flex-col bg-[#061833] px-4 py-6 text-blue-100" aria-label="เมนูระบบ">
        <div className="flex items-center gap-3 border-b border-white/10 pb-7"><span className="grid size-11 place-items-center rounded-xl bg-gradient-to-br from-sky-300 to-blue-600 font-black text-[#061833] shadow-xl shadow-blue-950/50">RF</span><div><p className="font-bold tracking-tight text-white">ResearchFlow</p><p className="text-xs text-blue-300">Research operations</p></div></div>
        <nav className="mt-8 space-y-2" aria-label="การนำทาง"><p className="px-3 pb-2 text-[10px] font-bold uppercase tracking-[.18em] text-blue-400">Workspace</p><a href="#register" aria-current="page" className="flex h-11 items-center gap-3 rounded-lg bg-blue-500/25 px-3 text-sm font-bold text-white shadow-[inset_0_0_0_1px_rgba(147,197,253,.2)]"><span aria-hidden="true">▦</span> ทะเบียนโครงการ</a><button type="button" onClick={(event) => openCreate(event.currentTarget)} className="flex h-11 w-full items-center gap-3 rounded-lg px-3 text-left text-sm font-semibold text-blue-200 hover:bg-white/10 hover:text-white focus-visible:bg-white/10"><span aria-hidden="true">＋</span> สร้างโครงการใหม่</button></nav>
        <div className="mt-auto rounded-xl border border-white/10 bg-white/5 p-4"><p className="text-[10px] font-bold uppercase tracking-[.15em] text-blue-300">Shared register</p><p className="mt-2 text-sm leading-5 text-blue-100">มุมมองกลางสำหรับติดตามทุกโครงการวิจัย</p></div>
      </aside>
      <div className="min-w-0">
        <header className="relative overflow-hidden bg-[linear-gradient(118deg,#061833_0%,#123f8b_60%,#2563eb_100%)] px-12 py-11 text-white"><div className="absolute -right-24 -top-28 size-[430px] rounded-full border border-sky-100/40 shadow-[0_0_0_36px_rgba(147,197,253,.1),0_0_0_72px_rgba(147,197,253,.05)]" /><div className="relative flex items-end justify-between"><div className="max-w-2xl"><p className="text-xs font-bold uppercase tracking-[.16em] text-sky-300">Research trajectory · live register</p><h1 className="mt-4 text-5xl font-bold tracking-[-.065em]">มองเห็นทุกโครงการ<br /><span className="text-sky-300">และจุดที่ต้องขับเคลื่อน</span></h1><p className="mt-5 max-w-xl text-base leading-7 text-blue-100">ทะเบียนโครงการที่ออกแบบเพื่อช่วยให้ทีมค้นหา ตรวจสอบ และจัดการงานวิจัยได้ในพื้นที่เดียว</p></div><button ref={createButton} type="button" onClick={(event) => openCreate(event.currentTarget)} className="relative rounded-xl bg-white px-5 py-3.5 text-sm font-bold text-blue-950 shadow-xl shadow-blue-950/25 hover:-translate-y-0.5 hover:bg-sky-100 focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-sky-200">＋ สร้างโครงการ</button></div><div className="relative mt-11 flex w-[430px] gap-1"><i className="h-1 flex-[.35] rounded-full bg-sky-300" /><i className="h-1 flex-[.25] rounded-full bg-white/65" /><i className="h-1 flex-1 rounded-full bg-white/20" /></div></header>
        <main id="register" className="px-10 pb-12">
          {notice && <div className={`mt-5 flex items-center justify-between rounded-xl border px-4 py-3 text-sm font-semibold ${notice.tone === 'success' ? 'border-emerald-200 bg-emerald-50 text-emerald-900' : 'border-red-200 bg-red-50 text-red-900'}`} role={notice.tone === 'error' ? 'alert' : 'status'}><span>{notice.message}</span><button type="button" onClick={() => setNotice(undefined)} aria-label="ปิดข้อความ">×</button></div>}
          {state.kind === 'ready' && <section className="relative -mt-5 grid grid-cols-3 gap-4" aria-label="สรุปภาพรวม"><Metric label="โครงการในทะเบียน" value={researches.length} tone="white" /><Metric label="กำลังดำเนินการ" value={active} tone="blue" /><Metric label="สิ้นสุดโครงการ" value={researches.length - active} tone="dark" /></section>}
          {state.kind === 'loading' && <Loading />}
          {state.kind === 'error' && <Failure message={state.message} onRetry={retry} />}
          {state.kind === 'ready' && (researches.length === 0 ? <Empty onAdd={openCreate} /> : <>
            <section className="mt-9 rounded-2xl border border-blue-100 bg-white p-5 shadow-sm"><div className="flex items-end justify-between"><div><p className="text-xs font-bold uppercase tracking-[.15em] text-blue-600">Find a project</p><h2 className="mt-1 text-xl font-bold tracking-tight">ค้นหาและจัดกลุ่มข้อมูล</h2></div><button type="button" onClick={resetFilters} className="text-sm font-bold text-blue-700 hover:text-blue-950">ล้างตัวกรอง</button></div><div className="mt-5 grid grid-cols-[1fr_220px_240px] gap-3"><Field label="ค้นหา"><input className="control" type="search" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="รหัส ชื่อ รายละเอียด สถานะ หรือกระบวนการ" /></Field><Field label="สถานะ"><select className="control" value={status} onChange={(event) => setStatus(event.target.value)}><option value="">ทุกสถานะ</option>{statuses.map((item) => <option key={item} value={item}>{item}</option>)}</select></Field><Field label="กระบวนการ"><select className="control" value={process} onChange={(event) => setProcess(event.target.value)}><option value="">ทุกกระบวนการ</option>{processes.map((item) => <option key={item} value={item}>{item}</option>)}</select></Field></div></section>
            <section className="mt-8"><div className="mb-4 flex items-end justify-between"><div><p className="text-xs font-bold uppercase tracking-[.15em] text-blue-600">Project register</p><h2 className="mt-1 text-2xl font-bold tracking-tight">รายการโครงการ <span className="text-slate-400">{filtered.length.toLocaleString('th-TH')}</span></h2></div><p className="text-sm text-slate-500">จากทั้งหมด {researches.length.toLocaleString('th-TH')} รายการ</p></div><ResearchTable researches={filtered} changedId={changedId} onEdit={(research, button) => { editButton.current = button; setEditing(research) }} onDelete={(research, button) => { deleteButton.current = button; setDeleting(research) }} /></section>
          </>)}
        </main>
      </div>
    </div>
    {createOpen && <CreateResearchDialog onClose={closeCreate} onCreated={created} />}
    {editing && <EditResearchDialog research={editing} onClose={closeEdit} onUpdated={updated} onNotFound={notFound} />}
    {deleting && <DeleteResearchDialog research={deleting} showDuplicateContext={researches.filter((item) => item.title === deleting.title).length > 1} onClose={closeDelete} onDeleted={deleted} onNotFound={notFound} />}
  </div>
}

function Field({ label, children }: { label: string; children: ReactNode }) { return <label className="grid gap-1.5 text-xs font-bold text-slate-600"><span>{label}</span>{children}</label> }
function Metric({ label, value, tone }: { label: string; value: number; tone: 'white' | 'blue' | 'dark' }) { const color = { white: 'border-blue-100 bg-white text-slate-950', blue: 'border-blue-500 bg-blue-600 text-white', dark: 'border-[#061833] bg-[#061833] text-white' }[tone]; return <article className={`rounded-xl border p-5 shadow-lg shadow-blue-950/10 ${color}`}><p className="text-[11px] font-bold uppercase tracking-[.14em] opacity-70">{label}</p><p className="mt-2 text-4xl font-bold tracking-[-.06em]">{value.toLocaleString('th-TH')}</p></article> }
function ResearchTable({ researches, changedId, onEdit, onDelete }: { researches: Research[]; changedId?: number; onEdit: (research: Research, button: HTMLButtonElement) => void; onDelete: (research: Research, button: HTMLButtonElement) => void }) { if (researches.length === 0) return <div className="rounded-2xl border border-dashed border-blue-200 bg-white py-16 text-center"><h3 className="text-xl font-bold">ไม่พบโครงการที่ตรงกับเงื่อนไข</h3><p className="mt-2 text-slate-500">ลองเปลี่ยนคำค้นหาหรือตัวกรอง แล้วตรวจสอบอีกครั้ง</p></div>; return <div className="overflow-hidden rounded-2xl border border-blue-100 bg-white shadow-sm"><table className="w-full table-fixed"><thead className="bg-slate-950 text-left text-[11px] uppercase tracking-[.13em] text-blue-100"><tr><th className="w-24 px-5 py-4 text-center">รหัส</th><th className="w-[27%] px-5 py-4">โครงการ</th><th className="w-[26%] px-5 py-4">รายละเอียด</th><th className="w-[18%] px-5 py-4">สถานะ</th><th className="px-5 py-4">กระบวนการ</th><th className="w-32 px-5 py-4 text-center">จัดการ</th></tr></thead><tbody className="divide-y divide-blue-50">{researches.map((research) => <tr key={research.id} className={`align-middle transition hover:bg-blue-50/70 ${research.id === changedId ? 'bg-sky-50' : ''}`}><td className="px-5 py-5 text-center"><span className="font-mono text-sm font-bold text-blue-700">#{String(research.id).padStart(3, '0')}</span></td><td className="px-5 py-5"><p className="text-[11px] font-bold uppercase tracking-wide text-blue-600">{research.continuationOfId === null ? 'โครงการหลัก' : `ต่อเนื่องจาก #${research.continuationOfId}`}</p><h3 className="mt-1 font-bold leading-5 text-slate-900">{research.title}</h3></td><td className="px-5 py-5 text-sm leading-5 text-slate-500"><p className="line-clamp-2">{research.description}</p></td><td className="px-5 py-5"><span className={`inline-flex rounded-full px-3 py-1 text-xs font-bold ${terminalStatuses.has(research.status) ? 'bg-slate-100 text-slate-600' : 'bg-blue-100 text-blue-800'}`}>{research.status}</span></td><td className="px-5 py-5 text-sm font-semibold leading-5 text-slate-600">{research.process}</td><td className="px-5 py-5"><div className="flex justify-center gap-2"><button type="button" onClick={(event) => onEdit(research, event.currentTarget)} className="rounded-lg border border-blue-200 px-2.5 py-1.5 text-xs font-bold text-blue-700 hover:bg-blue-700 hover:text-white">แก้ไข</button><button type="button" onClick={(event) => onDelete(research, event.currentTarget)} className="rounded-lg border border-red-200 px-2.5 py-1.5 text-xs font-bold text-red-600 hover:bg-red-600 hover:text-white">ลบ</button></div></td></tr>)}</tbody></table></div> }
function Loading() { return <div className="mt-8 grid grid-cols-3 gap-4" aria-busy="true">{[1, 2, 3].map((item) => <div key={item} className="h-36 animate-pulse rounded-2xl bg-blue-100" />)}</div> }
function Failure({ message, onRetry }: { message: string; onRetry: () => void }) { return <section className="mt-8 rounded-2xl border border-red-200 bg-red-50 p-10 text-center" role="alert"><p className="text-xs font-bold uppercase tracking-widest text-red-600">Connection issue</p><h2 className="mt-2 text-2xl font-bold">ไม่สามารถโหลดทะเบียนโครงการได้</h2><p className="mt-2 text-slate-600">{message}</p><button type="button" onClick={onRetry} className="mt-6 rounded-xl bg-slate-950 px-5 py-3 font-bold text-white">ลองอีกครั้ง</button></section> }
function Empty({ onAdd }: { onAdd: (button: HTMLButtonElement) => void }) { return <section className="mt-8 rounded-2xl border border-dashed border-blue-200 bg-white p-16 text-center"><p className="text-xs font-bold uppercase tracking-widest text-blue-600">Ready to begin</p><h2 className="mt-2 text-3xl font-bold tracking-tight">ทะเบียนยังไม่มีโครงการ</h2><p className="mx-auto mt-3 max-w-md text-slate-500">สร้างโครงการแรกเพื่อเริ่มติดตามความคืบหน้าจากพื้นที่ทำงานกลาง</p><button type="button" onClick={(event) => onAdd(event.currentTarget)} className="mt-7 rounded-xl bg-blue-600 px-5 py-3 font-bold text-white shadow-lg shadow-blue-200 hover:bg-blue-700">สร้างโครงการใหม่</button></section> }
function errorMessage(error: unknown) { switch (getAPIErrorCode(error)) { case 'INTERNAL_ERROR': return 'ระบบยังไม่สามารถดึงข้อมูลได้ โปรดลองอีกครั้ง'; case 'INVALID_REQUEST_BODY': case 'VALIDATION_ERROR': return 'คำขอรายการไม่ถูกต้อง โปรดลองโหลดหน้าใหม่'; default: return 'ตรวจสอบการเชื่อมต่อแล้วลองอีกครั้ง' } }
