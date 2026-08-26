import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { listResearches, type Research } from '../api/client'
import { getAPIErrorCode } from '../api/errors'
import { CatalogHeading } from '../components/CatalogHeading'
import { CreateResearchDialog } from '../components/CreateResearchDialog'
import { DeleteResearchDialog } from '../components/DeleteResearchDialog'
import { EditResearchDialog } from '../components/EditResearchDialog'
import { ResearchList } from '../components/ResearchList'

type ListState =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'success'; researches: Research[] }

type Feedback = { tone: 'success' | 'error'; message: string }

const EMPTY_RESEARCHES: Research[] = []

export function ResearchListPage() {
  const [requestKey, setRequestKey] = useState(0)
  const [state, setState] = useState<ListState>({ status: 'loading' })
  const [createOpen, setCreateOpen] = useState(false)
  const [editingResearch, setEditingResearch] = useState<Research>()
  const [deletingResearch, setDeletingResearch] = useState<Research>()
  const [feedback, setFeedback] = useState<Feedback>()
  const [highlightedID, setHighlightedID] = useState<number>()
  const [navigationOpen, setNavigationOpen] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [processFilter, setProcessFilter] = useState('')
  const [filterOpen, setFilterOpen] = useState(false)
  const createOpenerRef = useRef<HTMLButtonElement>(null)
  const editOpenerRef = useRef<HTMLButtonElement>(null)
  const deleteOpenerRef = useRef<HTMLButtonElement>(null)
  const headerActionRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    const controller = new AbortController()

    async function loadResearches() {
      setState({ status: 'loading' })
      try {
        const result = await listResearches({ signal: controller.signal })
        if (result.error) {
          setState({ status: 'error', message: listErrorMessage(result.error) })
          return
        }
        setState({ status: 'success', researches: result.data })
      } catch (error) {
        if (!controller.signal.aborted) {
          setState({ status: 'error', message: listErrorMessage(error) })
        }
      }
    }

    void loadResearches()
    return () => controller.abort()
  }, [requestKey])

  const retry = useCallback(() => setRequestKey((key) => key + 1), [])

  useEffect(() => {
    if (!feedback) return
    const timeout = window.setTimeout(() => setFeedback(undefined), 5000)
    return () => window.clearTimeout(timeout)
  }, [feedback])

  useEffect(() => {
    if (highlightedID === undefined || state.status !== 'success') return
    const timeout = window.setTimeout(() => setHighlightedID(undefined), 2600)
    return () => window.clearTimeout(timeout)
  }, [highlightedID, state.status])

  const researches = state.status === 'success' ? state.researches : EMPTY_RESEARCHES
  const filteredResearches = useMemo(() => {
    const query = searchTerm.trim().toLocaleLowerCase('th-TH')
    return researches.filter((research) => {
      const matchesQuery = query === '' || [research.id, research.title, research.description, research.status, research.process]
        .some((value) => String(value).toLocaleLowerCase('th-TH').includes(query))
      return matchesQuery && (statusFilter === '' || research.status === statusFilter) && (processFilter === '' || research.process === processFilter)
    })
  }, [processFilter, researches, searchTerm, statusFilter])
  const statusOptions = [...new Set(researches.map((research) => research.status))]
  const processOptions = [...new Set(researches.map((research) => research.process))]
  const filtersActive = searchTerm !== '' || statusFilter !== '' || processFilter !== ''

  function clearFilters() {
    setSearchTerm('')
    setStatusFilter('')
    setProcessFilter('')
  }

  function openCreateDialog(opener: HTMLButtonElement) {
    createOpenerRef.current = opener
    setCreateOpen(true)
  }

  function closeCreateDialog() {
    setCreateOpen(false)
    window.requestAnimationFrame(() => createOpenerRef.current?.focus())
  }

  function handleCreated(research: Research) {
    setCreateOpen(false)
    setFeedback({ tone: 'success', message: `เพิ่มงานวิจัย “${research.title}” แล้ว` })
    setHighlightedID(research.id)
    setRequestKey((key) => key + 1)
    window.requestAnimationFrame(() => createOpenerRef.current?.focus())
  }

  function openEditDialog(research: Research, opener: HTMLButtonElement) {
    editOpenerRef.current = opener
    setEditingResearch(research)
  }

  function closeEditDialog() {
    setEditingResearch(undefined)
    window.requestAnimationFrame(() => editOpenerRef.current?.focus())
  }

  function handleUpdated(research: Research) {
    setEditingResearch(undefined)
    setFeedback({ tone: 'success', message: `บันทึกการแก้ไข “${research.title}” แล้ว` })
    setHighlightedID(research.id)
    setState((current) => current.status === 'success'
      ? { status: 'success', researches: current.researches.map((item) => item.id === research.id ? research : item) }
      : current)
    window.requestAnimationFrame(() => editOpenerRef.current?.focus())
    void refreshResearchesInPlace()
  }

  function handleEditedResearchNotFound() {
    setEditingResearch(undefined)
    setFeedback({ tone: 'error', message: 'รายการนี้ไม่มีอยู่แล้ว ระบบกำลังโหลดรายการล่าสุด' })
    void refreshResearchesInPlace().finally(() => headerActionRef.current?.focus())
  }

  function openDeleteDialog(research: Research, opener: HTMLButtonElement) {
    deleteOpenerRef.current = opener
    setDeletingResearch(research)
  }

  function closeDeleteDialog() {
    setDeletingResearch(undefined)
    window.requestAnimationFrame(() => deleteOpenerRef.current?.focus())
  }

  function handleDeleted(research: Research) {
    setDeletingResearch(undefined)
    setFeedback({ tone: 'success', message: `ลบงานวิจัย “${research.title}” แล้ว` })
    setState((current) => current.status === 'success'
      ? { status: 'success', researches: current.researches.filter((item) => item.id !== research.id) }
      : current)
    window.requestAnimationFrame(() => headerActionRef.current?.focus())
    void refreshResearchesInPlace()
  }

  function handleDeletedResearchNotFound() {
    setDeletingResearch(undefined)
    setFeedback({ tone: 'error', message: 'รายการนี้ไม่มีอยู่แล้ว ระบบกำลังโหลดรายการล่าสุด' })
    void refreshResearchesInPlace().finally(() => headerActionRef.current?.focus())
  }

  async function refreshResearchesInPlace() {
    try {
      const result = await listResearches()
      if (!result.error) setState({ status: 'success', researches: result.data })
    } catch {
      // Keep the last confirmed list visible; feedback from the completed action remains actionable.
    }
  }

  return (
    <div className={`app-shell ${navigationOpen ? 'navigation-open' : 'navigation-collapsed'}`}>
      <aside className="module-sidebar" aria-label="เมนู module">
        <div className="module-sidebar-brand">
          <button
            className="module-sidebar-toggle"
            type="button"
            aria-expanded={navigationOpen}
            aria-label={navigationOpen ? 'ยุบแถบ module' : 'ขยายแถบ module'}
            title={navigationOpen ? 'ยุบแถบ module' : 'ขยายแถบ module'}
            onClick={() => setNavigationOpen((open) => !open)}
          >
            <span className="hamburger-icon" aria-hidden="true"><i /><i /><i /></span>
          </button>
          <div className="module-brand-copy">
            <strong>ทะเบียนวิจัย</strong>
            <span>ระบบติดตามสถานะ</span>
          </div>
        </div>
        <nav className="module-navigation" aria-label="การนำทาง module">
          <p className="module-navigation-label">Module</p>
          <a className="module-navigation-link is-active" href="#main-content" aria-current="page" title="ทะเบียนงานวิจัย">
            <span className="module-navigation-icon" aria-hidden="true">▦</span>
            <span className="module-navigation-copy">ทะเบียนงานวิจัย</span>
          </a>
          <button className="module-navigation-link" type="button" onClick={(event) => openCreateDialog(event.currentTarget)} title="เพิ่มงานวิจัย">
            <span className="module-navigation-icon" aria-hidden="true">+</span>
            <span className="module-navigation-copy">เพิ่มงานวิจัย</span>
          </button>
        </nav>
      </aside>

      <div className="page-content">
      <header className="page-header">
        <div className="page-intro">
          <div className="page-mark" aria-hidden="true"><span>SU</span><i /></div>
          <p className="page-eyebrow">มหาวิทยาลัยศิลปากร · พระราชวังสนามจันทร์</p>
          <h1>ระบบติดตาม<br /><em>สถานะงานวิจัย</em></h1>
          <p className="page-summary">จัดการและดูรายการงานวิจัยที่ใช้งานร่วมกัน</p>
        </div>
        <div className="header-controls">
          <p className="header-note"><span />Green campus · ทะเบียนกลาง</p>
          <button ref={headerActionRef} className="primary-button header-action" type="button" onClick={(event) => openCreateDialog(event.currentTarget)}>
            <span aria-hidden="true">+</span> เพิ่มงานวิจัย
          </button>
        </div>
      </header>

      <div className="feedback-region" aria-live="polite" aria-atomic="true">
        {feedback && (
          <div className={`action-feedback action-feedback-${feedback.tone}`} role={feedback.tone === 'error' ? 'alert' : 'status'}>
            <span>{feedback.message}</span>
            <button type="button" onClick={() => setFeedback(undefined)} aria-label="ปิดข้อความแจ้งเตือน">×</button>
          </div>
        )}
      </div>

      <main id="main-content" className="research-workspace">
        {state.status === 'loading' && <ResearchListLoading />}
        {state.status === 'error' && <ResearchListError message={state.message} onRetry={retry} />}
        {state.status === 'success' &&
          (state.researches.length === 0 ? (
            <ResearchListEmpty onAdd={openCreateDialog} />
          ) : (
            <>
              <ResearchList
                researches={filteredResearches}
                totalCount={researches.length}
                filtersActive={filtersActive}
                onClearFilters={clearFilters}
                highlightedID={highlightedID}
                onEdit={openEditDialog}
                onDelete={openDeleteDialog}
                filterControls={
                  <div className={`research-filter ${filterOpen ? 'is-open' : ''}`} role="search" aria-label="ค้นหาและกรองงานวิจัย">
                    <div className="research-filter-summary">
                      <button
                        className={`filter-toggle ${filtersActive ? 'has-active-filters' : ''}`}
                        type="button"
                        aria-expanded={filterOpen}
                        aria-controls="research-filter-panel"
                        onClick={() => setFilterOpen((open) => !open)}
                      >
                        <span className="filter-toggle-glyph" aria-hidden="true">⌄</span>
                        <span>กรองงานวิจัย</span>
                        {filtersActive && <span className="filter-toggle-badge">กำลังกรอง</span>}
                      </button>
                      <p className="filter-result-count" role="status">พบ {filteredResearches.length.toLocaleString('th-TH')} จาก {researches.length.toLocaleString('th-TH')} รายการ</p>
                    </div>
                    {filterOpen && (
                      <div id="research-filter-panel" className="research-filter-bar">
                        <label className="filter-search">
                          <span>ค้นหา</span>
                          <input
                            type="search"
                            value={searchTerm}
                            onChange={(event) => setSearchTerm(event.target.value)}
                            placeholder="รหัส ชื่อ รายละเอียด"
                          />
                        </label>
                        <label className="filter-select">
                          <span>สถานะ</span>
                          <select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)}>
                            <option value="">ทุกสถานะ</option>
                            {statusOptions.map((status) => <option key={status} value={status}>{status}</option>)}
                          </select>
                        </label>
                        <label className="filter-select">
                          <span>กระบวนการ</span>
                          <select value={processFilter} onChange={(event) => setProcessFilter(event.target.value)}>
                            <option value="">ทุกกระบวนการ</option>
                            {processOptions.map((process) => <option key={process} value={process}>{process}</option>)}
                          </select>
                        </label>
                        <button className="filter-reset" type="button" onClick={clearFilters} disabled={!filtersActive}>ล้างตัวกรอง</button>
                      </div>
                    )}
                  </div>
                }
              />
            </>
          ))}
      </main>

      {createOpen && <CreateResearchDialog onClose={closeCreateDialog} onCreated={handleCreated} />}
      {editingResearch && (
        <EditResearchDialog
          research={editingResearch}
          onClose={closeEditDialog}
          onNotFound={handleEditedResearchNotFound}
          onUpdated={handleUpdated}
        />
      )}
      {deletingResearch && (
        <DeleteResearchDialog
          research={deletingResearch}
          showDuplicateContext={researches.filter((item) => item.title === deletingResearch.title).length > 1}
          onClose={closeDeleteDialog}
          onDeleted={handleDeleted}
          onNotFound={handleDeletedResearchNotFound}
        />
      )}
      </div>
    </div>
  )
}

function ResearchListLoading() {
  return (
    <section className="catalog" aria-labelledby="catalog-loading-title" aria-busy="true">
      <CatalogHeading id="catalog-loading-title" label="กำลังโหลดรายการงานวิจัย" />
      <p className="sr-only" role="status">กำลังโหลดรายการงานวิจัย</p>
      <div className="list-table skeleton-list" aria-hidden="true">
        {Array.from({ length: 4 }, (_, index) => (
          <div className="research-row skeleton-row" key={index}>
            <div>
              <span className="skeleton-line skeleton-title" />
            </div>
            <div>
              <span className="skeleton-line" />
              <span className="skeleton-line skeleton-line-short" />
            </div>
            <div><span className="skeleton-line skeleton-action" /></div>
          </div>
        ))}
      </div>
    </section>
  )
}

function ResearchListError({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <section className="catalog" aria-labelledby="catalog-error-title">
      <CatalogHeading id="catalog-error-title" label="รายการงานวิจัย" />
      <div className="state-panel state-panel-error" role="alert">
        <p className="state-kicker">โหลดข้อมูลไม่สำเร็จ</p>
        <h2>ไม่สามารถแสดงรายการงานวิจัยได้</h2>
        <p>{message}</p>
        <button className="secondary-button" type="button" onClick={onRetry}>
          ลองอีกครั้ง
        </button>
      </div>
    </section>
  )
}

function ResearchListEmpty({ onAdd }: { onAdd: (opener: HTMLButtonElement) => void }) {
  return (
    <section className="catalog" aria-labelledby="catalog-empty-title">
      <CatalogHeading id="catalog-empty-title" label="งานวิจัย 0 รายการ" />
      <div className="state-panel empty-state">
        <p className="state-kicker">ทะเบียนยังว่าง</p>
        <h2>ยังไม่มีงานวิจัย</h2>
        <p>เริ่มเพิ่มรายการแรกเพื่อให้ทุกคนเห็นงานวิจัยชุดเดียวกัน</p>
        <button className="primary-button empty-action" type="button" onClick={(event) => onAdd(event.currentTarget)}>เพิ่มงานวิจัย</button>
      </div>
    </section>
  )
}

function listErrorMessage(error: unknown): string {
  const code = getAPIErrorCode(error)
  switch (code) {
    case 'INTERNAL_ERROR':
      return 'ระบบยังไม่สามารถดึงข้อมูลได้ โปรดลองอีกครั้ง'
    case 'INVALID_REQUEST_BODY':
    case 'VALIDATION_ERROR':
      return 'คำขอรายการไม่ถูกต้อง โปรดลองโหลดหน้าใหม่'
    default:
      return 'ตรวจสอบการเชื่อมต่อแล้วลองอีกครั้ง'
  }
}
