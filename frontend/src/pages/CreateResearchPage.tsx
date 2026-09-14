import { useEffect, useRef, useState, type FormEvent } from 'react'
import { api, type CreateResearchRequest, type Research } from '../api'

const textFields = [
  ['title', 'ชื่อโครงการ'], ['fundingSourceName', 'ชื่อแหล่งทุน'], ['contractNumber', 'เลขที่สัญญาทุน'],
  ['responsibleProjectUnit', 'หน่วยงานรับผิดชอบโครงการ'], ['responsibleBudgetUnit', 'หน่วยงานรับผิดชอบงบประมาณ'],
  ['thaiAbstract', 'บทคัดย่อไทย'], ['englishAbstract', 'บทคัดย่ออังกฤษ'], ['objectives', 'วัตถุประสงค์'], ['keywords', 'คำสำคัญ'],
] as const
type Errors = Record<string, string>

export function CreateResearchPage({ onCancel, onCreated }: { onCancel: () => void; onCreated: (id: number) => void }) {
  const [kind, setKind] = useState('BUDGET')
  const [parents, setParents] = useState<Research[]>([])
  const [parentState, setParentState] = useState('loading')
  const [query, setQuery] = useState('')
  const [parentId, setParentId] = useState('')
  const [errors, setErrors] = useState<Errors>({})
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)
  const dirty = useRef(false)
  const submitting = useRef(false)
  const mainRef = useRef<HTMLElement>(null)
  const summaryRef = useRef<HTMLDivElement>(null)
  async function loadParents() {
    setParentState('loading')
    try {
      const response = await api.listResearches()
      if (response.error || !response.data) throw new Error('list')
      setParents(response.data)
      setParentState('success')
    } catch { setParentState('error') }
  }
  useEffect(() => {
    let active = true
    void api.listResearches().then((response) => {
      if (!active) return
      if (response.error || !response.data) setParentState('error')
      else { setParents(response.data); setParentState('success') }
    }).catch(() => { if (active) setParentState('error') })
    return () => { active = false }
  }, [])
  useEffect(() => { mainRef.current?.focus() }, [])
  useEffect(() => {
    const warn = (event: BeforeUnloadEvent) => {
      if (dirty.current) { event.preventDefault(); event.returnValue = '' }
    }
    window.addEventListener('beforeunload', warn)
    return () => window.removeEventListener('beforeunload', warn)
  }, [])
  useEffect(() => { if (message) summaryRef.current?.focus() }, [message, errors])
  function cancel() {
    if (submitting.current) return
    if (!dirty.current || window.confirm('ต้องการทิ้งข้อมูลที่ยังไม่ได้บันทึกหรือไม่?')) onCancel()
  }
  const matches = parents.filter((p) => [p.title, p.contractNumber].some((value) => value.toLowerCase().includes(query.trim().toLowerCase())))
  const errorFor = (name: string) => errors[name] ? <span className="field-error" id={name + '-error'}>{errors[name]}</span> : null
  const inputProps = (name: string) => ({ name, id: name, 'aria-invalid': !!errors[name], 'aria-describedby': errors[name] ? name + '-error' : undefined })
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (submitting.current) return
    const data = new FormData(event.currentTarget)
    const value = (key: string) => String(data.get(key) ?? '').trim()
    const problems: Errors = {}
    const checkText = (key: string, max = 1000) => {
      const v = value(key)
      if (!v || [...v].length > max) problems[key] = 'กรอกข้อมูล 1-' + max + ' ตัวอักษร'
      // Match the contract's control-character rules without silently deleting characters.
      if ([...v].some((c) => { const n = c.codePointAt(0)!; return (n < 32 && !['\n', '\t'].includes(c)) || n === 127 })) problems[key] = 'ข้อมูลมีอักขระควบคุมที่ไม่อนุญาต'
      return v
    }
    for (const [key] of textFields) checkText(key)
    if ([...value('title')].some((c) => c === '/' || c.charCodeAt(0) < 32)) problems.title = 'ชื่อโครงการห้ามมี / หรือการขึ้นบรรทัดใหม่และแท็บ'
    const decimal = (key: string, max = Infinity) => {
      const raw = value(key)
      const n = Number(raw)
      if (!/^\d+(\.\d{1,2})?$/.test(raw) || !Number.isFinite(n) || n <= 0 || n > max) problems[key] = 'กรอกจำนวนบวก ทศนิยมไม่เกิน 2 ตำแหน่ง' + (max === 100 ? ' และไม่เกิน 100' : '')
      return n
    }
    const projectMembers = (['LEAD', 'CO_RESEARCHER'] as const).map((role, i) => {
      const prefix = 'projectMembers[' + i + '].'
      const fullName = checkText(prefix + 'fullName')
      const affiliation = checkText(prefix + 'affiliation')
      const email = checkText(prefix + 'email', 254)
      if (!/^[^\s@]+@[^\s@]+$/.test(email)) problems[prefix + 'email'] = 'กรอกอีเมลให้ถูกต้อง'
      return { fullName, affiliation, email, role, contributionPercent: decimal(prefix + 'contributionPercent', 100) }
    })
    if (projectMembers[0].email.toLowerCase() === projectMembers[1].email.toLowerCase()) problems['projectMembers[1].email'] = 'หัวหน้าและผู้ร่วมโครงการต้องใช้อีเมลต่างกัน'
    if (Math.round(projectMembers.reduce((sum, m) => sum + m.contributionPercent, 0) * 100) !== 10000) problems.projectMembers = 'สัดส่วนของทั้งสองคนต้องรวมเท่ากับ 100%'
    const date = (key: string) => {
      const raw = value(key)
      const parts = raw.split('/').map(Number)
      const d = new Date(Date.UTC(parts[2] - 543, parts[1] - 1, parts[0]))
      if (!/^\d{2}\/\d{2}\/\d{4}$/.test(raw) || d.getUTCDate() !== parts[0] || d.getUTCMonth() !== parts[1] - 1 || d.getUTCFullYear() !== parts[2] - 543) problems[key] = 'กรอกวันที่จริงในรูปแบบ วว/ดด/ปี พ.ศ.'
      return d
    }
    const start = date('startDate')
    const end = date('endDate')
    const anniversary = new Date(start)
    anniversary.setUTCFullYear(start.getUTCFullYear() + 1)
    if (end < anniversary) problems.endDate = 'วันสิ้นสุดต้องห่างจากวันเริ่มอย่างน้อย 1 ปี'
    const budgetAmount = decimal('budgetAmount')
    const file = data.get('contractFile') as File
    if (!file || !file.size) problems.contractFile = 'เลือกไฟล์สัญญา PDF ที่ไม่ว่าง'
    if (kind === 'CONTINUATION' && (!parentId || !parents.some((p) => p.id === Number(parentId)))) problems.continuationOfId = 'เลือกงานวิจัยต้นทาง'
    if (Object.keys(problems).length) {
      setErrors(problems); setMessage('โปรดตรวจสอบข้อมูลที่ระบุด้านล่าง'); return
    }
    const body: CreateResearchRequest = {
      title: value('title'), continuationOfId: kind === 'CONTINUATION' ? Number(parentId) : null,
      isSubsidized: data.get('isSubsidized') === 'on', projectMembers: [projectMembers[0], projectMembers[1]],
      fundingType: value('fundingType') as CreateResearchRequest['fundingType'],
      fundingSourceName: value('fundingSourceName'), contractNumber: value('contractNumber'), contractFile: file,
      projectType: value('projectType') as CreateResearchRequest['projectType'], researchKind: kind as CreateResearchRequest['researchKind'],
      responsibleProjectUnit: value('responsibleProjectUnit'), responsibleBudgetUnit: value('responsibleBudgetUnit'),
      startDate: value('startDate'), endDate: value('endDate'), budgetAmount,
      thaiAbstract: value('thaiAbstract'), englishAbstract: value('englishAbstract'), objectives: value('objectives'), keywords: value('keywords'),
    }
    submitting.current = true; setBusy(true); setErrors({}); setMessage('')
    try {
      const result = await api.createResearch(body)
      if (result.error || !result.data) {
        const fields: Errors = {}
        const error = result.error?.error
        if (error && 'fieldErrors' in error) for (const field of error.fieldErrors) fields[field.field] = field.message
        if (error?.code === 'TITLE_ALREADY_EXISTS') fields.title = 'ชื่อโครงการนี้มีอยู่แล้ว'
        if (error?.code === 'CONTRACT_NUMBER_ALREADY_EXISTS') fields.contractNumber = 'เลขสัญญาทุนนี้มีอยู่แล้ว'
        if (error?.code === 'CONTINUATION_NOT_FOUND') { fields.continuationOfId = 'ไม่พบงานต้นทาง โปรดเลือกใหม่'; setParentId(''); void loadParents() }
        if (error?.code === 'PAYLOAD_TOO_LARGE') fields.contractFile = 'ไฟล์หรือคำขอมีขนาดเกินกำหนด'
        setErrors(fields); setMessage('บันทึกไม่สำเร็จ โปรดตรวจสอบข้อมูลหรือลองอีกครั้ง')
      } else { dirty.current = false; onCreated(result.data.id) }
    } catch { setMessage('เชื่อมต่อไม่สำเร็จ ข้อมูลยังอยู่ในฟอร์ม โปรดลองอีกครั้ง') }
    finally { submitting.current = false; setBusy(false) }
  }
  return <main id="create-research" tabIndex={-1} ref={mainRef}>
    <header className="page-heading"><div><h1>เพิ่มงานวิจัย</h1><p>กรอกข้อมูลโครงการ บุคลากร และแนบสัญญาทุน</p></div><button type="button" disabled={busy} onClick={cancel}>ย้อนกลับ</button></header>
    {message && <div className="notice error" role="alert" tabIndex={-1} ref={summaryRef}><p>{message}</p><ul>{Object.entries(errors).map(([key, error]) => <li key={key}><a href={'#' + key}>{error}</a></li>)}</ul></div>}
    <form noValidate onSubmit={submit} onChange={() => { dirty.current = true }}>
      <fieldset disabled={busy}><legend>ข้อมูลโครงการ</legend>
        <label htmlFor="title">ชื่อโครงการ *</label><input {...inputProps('title')} />{errorFor('title')}
        <label className="checkbox-label"><input name="isSubsidized" type="checkbox" />เป็นทุนอุดหนุน</label>
        <label htmlFor="projectType">ประเภทโครงการ *</label><select id="projectType" name="projectType"><option value="RESEARCH">งานวิจัย</option><option value="ACADEMIC_SERVICE">บริการวิชาการ</option></select>
        <label htmlFor="researchKind">ลักษณะโครงการ *</label><select id="researchKind" name="researchKind" value={kind} onChange={(e) => { setKind(e.target.value); setParentId('') }}><option value="BUDGET">โครงการในงบประมาณ</option><option value="CONTINUATION">โครงการต่อเนื่อง</option></select>
        {kind === 'CONTINUATION' && <div>
          {parentState === 'loading' && <p role="status">กำลังโหลดงานต้นทาง…</p>}
          {parentState === 'error' && <div role="alert"><p>โหลดงานต้นทางไม่สำเร็จ</p><button type="button" onClick={() => void loadParents()}>ลองโหลดงานต้นทางใหม่</button></div>}
          {parentState === 'success' && <><label htmlFor="parent-search">ค้นหางานต้นทางด้วยชื่อหรือเลขสัญญาทุน</label><input id="parent-search" value={query} onChange={(e) => { setQuery(e.target.value); setParentId('') }} />
            <label htmlFor="continuationOfId">งานวิจัยต้นทาง *</label><select {...inputProps('continuationOfId')} value={parentId} onChange={(e) => setParentId(e.target.value)}><option value="">เลือกงานต้นทาง</option>{matches.map((p) => <option key={p.id} value={p.id}>{p.contractNumber} · {p.title}</option>)}</select>
            {!matches.length && <p>{parents.length ? 'ไม่พบงานที่ตรงกับคำค้น' : 'ยังไม่มีงานวิจัยต้นทาง'}</p>}</>}
          {errorFor('continuationOfId')}
        </div>}
      </fieldset>
      <fieldset disabled={busy} id="projectMembers"><legend>บุคลากรในโครงการ</legend><p>หัวหน้า 1 คน และผู้ร่วม 1 คน สัดส่วนรวม 100%</p>{errorFor('projectMembers')}
        {[0, 1].map((i) => <div className="member-fields" key={i}><h2>{i === 0 ? 'หัวหน้าโครงการ' : 'ผู้ร่วมโครงการ'}</h2>{[['fullName', 'ชื่อ-นามสกุล'], ['email', 'อีเมล'], ['affiliation', 'หน่วยงาน'], ['contributionPercent', 'สัดส่วนการมีส่วนร่วม (%)']].map(([key, label]) => {
          const name = 'projectMembers[' + i + '].' + key
          return <div key={key}><label htmlFor={name}>{label} *</label><input {...inputProps(name)} inputMode={key === 'contributionPercent' ? 'decimal' : key === 'email' ? 'email' : 'text'} />{errorFor(name)}</div>
        })}</div>)}
      </fieldset>
      <fieldset disabled={busy}><legend>แหล่งทุนและสัญญา</legend>
        <label htmlFor="fundingType">ประเภทแหล่งทุน *</label><select name="fundingType" id="fundingType"><option value="INTERNAL">ภายใน</option><option value="EXTERNAL">ภายนอก</option></select>
        {textFields.slice(1, 3).map(([key, label]) => <div key={key}><label htmlFor={key}>{label} *</label><input {...inputProps(key)} />{errorFor(key)}</div>)}
        <label htmlFor="contractFile">ไฟล์สัญญาทุน PDF *</label><input {...inputProps('contractFile')} type="file" accept=".pdf,application/pdf" /><p>สูงสุด 20 MiB ระบบจะตรวจเนื้อหา PDF เมื่อบันทึก</p>{errorFor('contractFile')}
      </fieldset>
      <fieldset disabled={busy}><legend>รายละเอียดโครงการ</legend>
        {textFields.slice(3).map(([key, label]) => <div key={key}><label htmlFor={key}>{label} *</label><textarea {...inputProps(key)} rows={key.includes('Unit') ? 2 : 3} />{errorFor(key)}</div>)}
        {[['startDate', 'วันเริ่มต้น (วว/ดด/ปี พ.ศ.)'], ['endDate', 'วันสิ้นสุด (วว/ดด/ปี พ.ศ.)'], ['budgetAmount', 'งบประมาณ (บาท)']].map(([key, label]) => <div key={key}><label htmlFor={key}>{label} *</label><input {...inputProps(key)} inputMode={key === 'budgetAmount' ? 'decimal' : 'text'} />{errorFor(key)}</div>)}
      </fieldset>
      <div className="form-actions"><button type="submit" disabled={busy}>{busy ? 'กำลังบันทึก…' : 'บันทึกงานวิจัย'}</button><button type="button" onClick={cancel} disabled={busy}>ยกเลิก</button></div>
      {busy && <p role="status">กำลังตรวจสอบและบันทึกโครงการ โปรดรอสักครู่</p>}
    </form>
  </main>
}
