import { useCallback, useEffect, useState } from 'react'
import { listResearches, type Research, type ResearchMember } from '../api/client'
import { getAPIErrorCode } from '../api/errors'
import './ResearchRegisterPage.css'

type RegisterState =
  | { kind: 'loading' }
  | { kind: 'error'; message: string }
  | { kind: 'success'; researches: Research[] }

export function ResearchRegisterPage({ createdResearchId, onCreate }: { createdResearchId?: number; onCreate: () => void }) {
  const [state, setState] = useState<RegisterState>({ kind: 'loading' })
  const [refreshKey, setRefreshKey] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    void loadRegister(controller.signal, setState)
    return () => controller.abort()
  }, [refreshKey])

  const retry = useCallback(() => setRefreshKey((key) => key + 1), [])

  return (
    <main className="register-page">
      <header className="register-masthead">
        <div className="register-mark" aria-hidden="true"><span>SU</span><i /></div>
        <div>
          <p className="register-kicker">ทะเบียนงานวิจัย · มหาวิทยาลัยศิลปากร</p>
          <h1>บัญชีรายการกลาง</h1>
          <p>รายการโครงการที่ใช้ร่วมกัน พร้อมข้อมูลอ้างอิงและความคืบหน้าล่าสุด</p>
        </div>
        <div className="register-masthead-tools">
          <div className="register-masthead-note"><span aria-hidden="true" />ข้อมูลทะเบียนกลาง</div>
          <button className="register-create-button" type="button" onClick={onCreate}><span aria-hidden="true">+</span> เพิ่มงานวิจัย</button>
        </div>
      </header>

      <section className="register-content" aria-labelledby="register-title">
        {state.kind === 'loading' && <RegisterLoading />}
        {state.kind === 'error' && <RegisterError message={state.message} onRetry={retry} />}
        {state.kind === 'success' && <RegisterSuccess createdResearchId={createdResearchId} researches={state.researches} onCreate={onCreate} />}
      </section>
    </main>
  )
}

async function loadRegister(signal: AbortSignal, setState: (state: RegisterState) => void) {
  setState({ kind: 'loading' })
  try {
    const result = await listResearches({ signal })
    if (result.error) {
      setState({ kind: 'error', message: listErrorMessage(result.error) })
      return
    }
    setState({ kind: 'success', researches: result.data })
  } catch (error) {
    if (!signal.aborted) setState({ kind: 'error', message: listErrorMessage(error) })
  }
}

function RegisterLoading() {
  return <>
    <RegisterHeading label="กำลังโหลดทะเบียนงานวิจัย" />
    <p className="sr-only" role="status">กำลังโหลดรายการงานวิจัย</p>
    <div className="register-skeleton" aria-hidden="true">
      {Array.from({ length: 5 }, (_, index) => <div className="register-skeleton-row" key={index}><i /><i /><i /><i /></div>)}
    </div>
  </>
}

function RegisterError({ message, onRetry }: { message: string; onRetry: () => void }) {
  return <>
    <RegisterHeading label="ทะเบียนงานวิจัย" />
    <div className="register-state register-state-error" role="alert">
      <p className="state-label">ไม่สามารถโหลดทะเบียนได้</p>
      <h2>ตรวจสอบการเชื่อมต่อแล้วลองอีกครั้ง</h2>
      <p>{message}</p>
      <button className="register-button" type="button" onClick={onRetry}>ลองโหลดใหม่</button>
    </div>
  </>
}

function RegisterSuccess({ createdResearchId, researches, onCreate }: { createdResearchId?: number; researches: Research[]; onCreate: () => void }) {
  if (researches.length === 0) {
    return <>
      <RegisterHeading label="งานวิจัย 0 รายการ" />
      <div className="register-state register-state-empty">
        <p className="state-label">ทะเบียนยังว่าง</p>
        <h2>ยังไม่มีงานวิจัยในทะเบียนกลาง</h2>
        <p>เมื่อมีการเพิ่มโครงการ รายการทั้งหมดจะแสดงในหน้านี้</p>
        <button className="register-button" type="button" onClick={onCreate}><span aria-hidden="true">+</span> เพิ่มงานวิจัย</button>
      </div>
    </>
  }

  return <>
    <RegisterHeading label={`งานวิจัย ${researches.length.toLocaleString('th-TH')} รายการ`} />
    {createdResearchId && <p className="register-created-notice" role="status">เพิ่มงานวิจัยรหัส {createdResearchId} แล้ว รายการด้านล่างเป็นข้อมูลล่าสุด</p>}
    <p className="register-reading-note" id="register-description">เรียงตามชื่อโครงการ และใช้รหัสงานวิจัยสำหรับอ้างอิงรายการที่มีชื่อซ้ำ</p>
    <div className="register-table-wrap" role="region" aria-labelledby="register-title" aria-describedby="register-description" tabIndex={0}>
      <table className="register-table">
        <thead><tr><th scope="col">รหัส</th><th scope="col">โครงการ</th><th scope="col">บุคลากร</th><th scope="col">ระยะเวลา</th><th scope="col">สถานะและกระบวนการ</th><th scope="col">สัญญาทุน</th><th scope="col"><span className="sr-only">รายละเอียด</span></th></tr></thead>
        <tbody>{researches.map((research) => <ResearchRow key={research.id} research={research} />)}</tbody>
      </table>
    </div>
  </>
}

function RegisterHeading({ label }: { label: string }) {
  return <div className="register-heading"><div><p className="section-label">ทะเบียนกลาง</p><h2 id="register-title">{label}</h2></div><span aria-hidden="true">{new Date().getFullYear() + 543}</span></div>
}

function ResearchRow({ research }: { research: Research }) {
  return <tr>
    <td className="register-id"><span>RESEARCH</span><strong>{String(research.id).padStart(3, '0')}</strong></td>
    <td className="register-project"><p>{research.projectType === 'RESEARCH' ? 'งานวิจัย' : 'บริการวิชาการ'} · {research.researchKind === 'CONTINUATION' ? 'โครงการต่อเนื่อง' : 'โครงการงบประมาณ'}</p><h3>{research.title}</h3>{research.continuationOfId !== null && <small>ต่อยอดจาก #{research.continuationOfId}</small>}</td>
    <td className="register-members"><strong>{leadMember(research.projectMembers)?.fullName ?? '—'}</strong><span>หัวหน้าโครงการ</span><small>ผู้ร่วม {countCoMembers(research.projectMembers)} คน</small></td>
    <td className="register-period"><strong>{research.startDate}</strong><span>ถึง</span><strong>{research.endDate}</strong><small>{formatCurrency(research.budgetAmount)} บาท</small></td>
    <td className="register-progress"><span className={`status-chip ${statusTone(research.status)}`}>{research.status}</span><strong>{research.process}</strong></td>
    <td className="register-contract"><strong>{research.fundingSourceName}</strong><span>{research.fundingType === 'INTERNAL' ? 'ทุนภายใน' : 'ทุนภายนอก'} · {research.contractNumber}</span><small>{research.contractFile.filename} · {formatFileSize(research.contractFile.sizeBytes)}</small></td>
    <td className="register-expand"><details><summary><span className="sr-only">ดูรายละเอียดของ {research.title}</span><span aria-hidden="true">↗</span></summary><ResearchDetails research={research} /></details></td>
  </tr>
}

function ResearchDetails({ research }: { research: Research }) {
  return <div className="research-details">
    <section><h4>หน่วยงานและทุน</h4><dl><div><dt>หน่วยงานโครงการ</dt><dd>{research.responsibleProjectUnit}</dd></div><div><dt>หน่วยงานงบประมาณ</dt><dd>{research.responsibleBudgetUnit}</dd></div><div><dt>เลขที่สัญญา</dt><dd>{research.contractNumber}</dd></div><div><dt>ไฟล์สัญญา</dt><dd>{research.contractFile.filename} ({research.contractFile.contentType})</dd></div></dl></section>
    <section><h4>บุคลากรโครงการ</h4><ul>{research.projectMembers.map((member, index) => <li key={`${member.email}-${index}`}><strong>{member.fullName}</strong><span>{member.role === 'LEAD' ? 'หัวหน้าโครงการ' : 'ผู้ร่วมโครงการ'} · {member.affiliation} · {member.email} · {member.contributionPercent}%</span></li>)}</ul></section>
    <section className="research-detail-prose"><h4>รายละเอียดโครงการ</h4><dl><div><dt>บทคัดย่อไทย</dt><dd>{research.thaiAbstract}</dd></div><div><dt>บทคัดย่ออังกฤษ</dt><dd>{research.englishAbstract}</dd></div><div><dt>วัตถุประสงค์</dt><dd>{research.objectives}</dd></div><div><dt>คำสำคัญ</dt><dd>{research.keywords}</dd></div></dl></section>
  </div>
}

function leadMember(members: ResearchMember[]) { return members.find((member) => member.role === 'LEAD') }
function countCoMembers(members: ResearchMember[]) { return members.filter((member) => member.role === 'CO_RESEARCHER').length }
function formatCurrency(amount: number) { return amount.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }
function formatFileSize(bytes: number) { return bytes < 1024 * 1024 ? `${Math.ceil(bytes / 1024)} KB` : `${(bytes / 1024 / 1024).toFixed(1)} MB` }
function statusTone(status: Research['status']) { return status === 'โครงการเสร็จสิ้น' ? 'is-complete' : status === 'ยุติโครงการ' ? 'is-ended' : 'is-active' }
function listErrorMessage(error: unknown) { switch (getAPIErrorCode(error)) { case 'INTERNAL_ERROR': return 'ระบบยังไม่สามารถดึงข้อมูลทะเบียนได้'; case 'VALIDATION_ERROR': return 'คำขอรายการไม่ถูกต้อง กรุณาลองโหลดหน้าใหม่'; default: return 'ไม่สามารถเชื่อมต่อกับระบบทะเบียนได้ในขณะนี้' } }
