import { useRef, useState, type FormEvent } from 'react'
import {
  createResearchMultipart,
  type CreateResearchRequest,
  type ResearchMember,
} from '../api/client'
import { getAPIErrorCode } from '../api/errors'
import './CreateResearchPage.css'

type MemberDraft = ResearchMember & { key: number }
type Errors = Record<string, string>

const MAX_FILE_BYTES = 20 * 1024 * 1024

const initialMembers: MemberDraft[] = [
  { key: 1, fullName: '', email: '', affiliation: '', contributionPercent: 50, role: 'LEAD' },
  { key: 2, fullName: '', email: '', affiliation: '', contributionPercent: 50, role: 'CO_RESEARCHER' },
]

const initialForm = () => ({
  title: '',
  isSubsidized: undefined as boolean | undefined,
  fundingType: 'INTERNAL' as CreateResearchRequest['fundingType'],
  fundingSourceName: '',
  contractNumber: '',
  projectType: 'RESEARCH' as CreateResearchRequest['projectType'],
  researchKind: 'BUDGET' as CreateResearchRequest['researchKind'],
  continuationOfId: '',
  responsibleProjectUnit: '',
  responsibleBudgetUnit: '',
  startDate: '',
  endDate: '',
  budgetAmount: '',
  thaiAbstract: '',
  englishAbstract: '',
  objectives: '',
  keywords: '',
})

export function CreateResearchPage({ onReturnToRegister }: { onReturnToRegister: (createdResearchId?: number) => void }) {
  const [form, setForm] = useState(initialForm)
  const [members, setMembers] = useState(initialMembers)
  const [contractFile, setContractFile] = useState<File>()
  const [errors, setErrors] = useState<Errors>({})
  const [submitting, setSubmitting] = useState(false)
  const [notice, setNotice] = useState('')
  const titleRef = useRef<HTMLInputElement>(null)

  const update = <K extends keyof ReturnType<typeof initialForm>>(key: K, value: ReturnType<typeof initialForm>[K]) => {
    setForm((current) => ({ ...current, [key]: value }))
    setErrors((current) => ({ ...current, [key]: '' }))
    setNotice('')
  }

  const updateMember = <K extends keyof ResearchMember>(key: number, field: K, value: ResearchMember[K]) => {
    setMembers((current) => current.map((member) => member.key === key ? { ...member, [field]: value } : member))
    setErrors((current) => ({ ...current, members: '' }))
  }

  function addMember() {
    setMembers((current) => [...current, {
      key: Math.max(0, ...current.map((member) => member.key)) + 1,
      fullName: '', email: '', affiliation: '', contributionPercent: 1, role: 'CO_RESEARCHER',
    }])
  }

  function removeMember(key: number) {
    setMembers((current) => current.filter((member) => member.key !== key))
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (submitting) return

    const fieldErrors = validate(form, members, contractFile)
    if (Object.keys(fieldErrors).length > 0) {
      setErrors(fieldErrors)
      setNotice('กรุณาตรวจสอบข้อมูลที่มีเครื่องหมายแจ้งเตือนก่อนบันทึก')
      titleRef.current?.focus()
      return
    }

    const request: CreateResearchRequest = {
      title: form.title.trim(),
      continuationOfId: form.researchKind === 'BUDGET' ? null : Number(form.continuationOfId),
      isSubsidized: form.isSubsidized as boolean,
      projectMembers: members.map((member) => ({
        fullName: member.fullName.trim(), email: member.email.trim(), affiliation: member.affiliation.trim(),
        contributionPercent: member.contributionPercent, role: member.role,
      })),
      fundingType: form.fundingType,
      fundingSourceName: form.fundingSourceName.trim(),
      contractNumber: form.contractNumber.trim(),
      contractFile: contractFile as File,
      projectType: form.projectType,
      researchKind: form.researchKind,
      responsibleProjectUnit: form.responsibleProjectUnit.trim(),
      responsibleBudgetUnit: form.responsibleBudgetUnit.trim(),
      startDate: form.startDate.trim(), endDate: form.endDate.trim(),
      budgetAmount: Number(form.budgetAmount),
      thaiAbstract: form.thaiAbstract.trim(), englishAbstract: form.englishAbstract.trim(),
      objectives: form.objectives.trim(), keywords: form.keywords.trim(),
    }

    setSubmitting(true)
    setErrors({})
    setNotice('กำลังตรวจสอบและบันทึกข้อมูลโครงการ…')
    try {
      const result = await createResearchMultipart(request)
      if (result.error) {
        setErrors({ form: apiErrorMessage(result.error) })
        setNotice('บันทึกไม่สำเร็จ ไฟล์ยังอยู่ในอุปกรณ์ของคุณและยังไม่ถูกจัดเก็บ')
        return
      }
      onReturnToRegister(result.data.id)
      return
    } catch (error) {
      setErrors({ form: apiErrorMessage(error) })
      setNotice('บันทึกไม่สำเร็จ ไฟล์ยังอยู่ในอุปกรณ์ของคุณและยังไม่ถูกจัดเก็บ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="create-page">
      <header className="create-hero">
        <div className="hero-seal" aria-hidden="true">SU</div>
        <div>
          <p className="hero-kicker">ทะเบียนวิจัย · รายการใหม่</p>
          <h1>เพิ่มงานวิจัย</h1>
          <p>กรอกข้อมูลโครงการให้ครบในครั้งเดียว ระบบจะจัดเก็บสัญญาเมื่อยืนยันบันทึกเท่านั้น</p>
        </div>
        <div className="hero-side" aria-label="ข้อมูลแบบฟอร์ม"><strong>01</strong><span>ฟอร์มข้อมูลโครงการ</span></div>
        <button className="return-to-register" type="button" disabled={submitting} onClick={() => onReturnToRegister()}>← กลับไปรายการ</button>
      </header>

      <section className="create-shell" aria-labelledby="research-form-title">
        <div className="create-progress" aria-label="ลำดับข้อมูลในแบบฟอร์ม">
          <span className="is-current">1 <b>โครงการ</b></span><span>2 <b>บุคลากร</b></span><span>3 <b>ทุนและสัญญา</b></span><span>4 <b>รายละเอียด</b></span>
        </div>
        <div className="form-heading">
          <div><p className="section-kicker">ข้อมูลที่จำเป็น</p><h2 id="research-form-title">ทะเบียนข้อมูลโครงการ</h2></div>
          <p><i aria-hidden="true">*</i> ต้องกรอกทุกช่องก่อนบันทึก</p>
        </div>

        <form onSubmit={submit} noValidate aria-busy={submitting}>
          {errors.form && <div className="form-alert" role="alert">{errors.form}</div>}
          {notice && <p className="form-notice" role="status">{notice}</p>}

          <fieldset className="form-section">
            <legend><span>01</span> ข้อมูลโครงการ</legend>
            <div className="field-grid two">
              <Field label="ชื่อโครงการ" error={errors.title} required>
                <input ref={titleRef} value={form.title} onChange={(event) => update('title', event.target.value)} maxLength={1000} aria-invalid={Boolean(errors.title)} placeholder="ระบุชื่อโครงการ" />
              </Field>
              <Field label="ลักษณะงบประมาณ" error={errors.isSubsidized} required>
                <div className="choice-row">
                  <Choice label="มีงบประมาณ" checked={form.isSubsidized === true} onChange={() => update('isSubsidized', true)} name="isSubsidized" />
                  <Choice label="ไม่มีงบประมาณ" checked={form.isSubsidized === false} onChange={() => update('isSubsidized', false)} name="isSubsidized" />
                </div>
              </Field>
              <Field label="ประเภทโครงการ" required>
                <div className="choice-row">
                  <Choice label="งานวิจัย" checked={form.projectType === 'RESEARCH'} onChange={() => update('projectType', 'RESEARCH')} name="projectType" />
                  <Choice label="บริการวิชาการ" checked={form.projectType === 'ACADEMIC_SERVICE'} onChange={() => update('projectType', 'ACADEMIC_SERVICE')} name="projectType" />
                </div>
              </Field>
              <Field label="ลักษณะงาน" error={errors.researchKind} required>
                <div className="choice-row">
                  <Choice label="โครงการงบประมาณ" checked={form.researchKind === 'BUDGET'} onChange={() => update('researchKind', 'BUDGET')} name="researchKind" />
                  <Choice label="โครงการต่อเนื่อง" checked={form.researchKind === 'CONTINUATION'} onChange={() => update('researchKind', 'CONTINUATION')} name="researchKind" />
                </div>
              </Field>
              {form.researchKind === 'CONTINUATION' && <Field label="รหัสโครงการต้นทาง" error={errors.continuationOfId} required>
                <input inputMode="numeric" value={form.continuationOfId} onChange={(event) => update('continuationOfId', event.target.value)} aria-invalid={Boolean(errors.continuationOfId)} placeholder="จำนวนเต็มบวก" />
              </Field>}
            </div>
          </fieldset>

          <fieldset className="form-section member-section">
            <legend><span>02</span> บุคลากรโครงการ</legend>
            <p className="section-description">เพิ่มหัวหน้าโครงการและผู้ร่วมโครงการอย่างน้อยอย่างละ 1 คน</p>
            {errors.members && <p className="field-error" role="alert">{errors.members}</p>}
            <div className="member-table" role="region" aria-label="รายชื่อบุคลากรโครงการ" tabIndex={0}>
              <div className="member-table-head"><span>บทบาท</span><span>ชื่อ-นามสกุล</span><span>อีเมล</span><span>หน่วยงานที่สังกัด</span><span>สัดส่วน</span><span aria-label="การจัดการ" /></div>
              {members.map((member, index) => <div className="member-row" key={member.key}>
                <select value={member.role} onChange={(event) => updateMember(member.key, 'role', event.target.value as ResearchMember['role'])} aria-label={`บทบาทบุคลากรคนที่ ${index + 1}`}><option value="LEAD">หัวหน้าโครงการ</option><option value="CO_RESEARCHER">ผู้ร่วมโครงการ</option></select>
                <input value={member.fullName} onChange={(event) => updateMember(member.key, 'fullName', event.target.value)} maxLength={1000} aria-label={`ชื่อ-นามสกุลคนที่ ${index + 1}`} />
                <input type="email" value={member.email} onChange={(event) => updateMember(member.key, 'email', event.target.value)} maxLength={1000} aria-label={`อีเมลคนที่ ${index + 1}`} />
                <input value={member.affiliation} onChange={(event) => updateMember(member.key, 'affiliation', event.target.value)} maxLength={1000} aria-label={`หน่วยงานคนที่ ${index + 1}`} />
                <input type="number" min="0.01" max="100" step="0.01" value={member.contributionPercent} onChange={(event) => updateMember(member.key, 'contributionPercent', Number(event.target.value))} aria-label={`สัดส่วนคนที่ ${index + 1}`} />
                <button className="icon-button" type="button" onClick={() => removeMember(member.key)} aria-label={`ลบบุคลากรคนที่ ${index + 1}`}>×</button>
              </div>)}
            </div>
            <button className="add-member" type="button" onClick={addMember}><span aria-hidden="true">+</span> เพิ่มบุคลากร</button>
          </fieldset>

          <fieldset className="form-section">
            <legend><span>03</span> แหล่งทุนและสัญญา</legend>
            <div className="field-grid two">
              <Field label="ประเภทแหล่งทุน" required><div className="choice-row"><Choice label="ภายใน" checked={form.fundingType === 'INTERNAL'} onChange={() => update('fundingType', 'INTERNAL')} name="fundingType" /><Choice label="ภายนอก" checked={form.fundingType === 'EXTERNAL'} onChange={() => update('fundingType', 'EXTERNAL')} name="fundingType" /></div></Field>
              <Field label="ชื่อแหล่งทุน" error={errors.fundingSourceName} required><input value={form.fundingSourceName} onChange={(event) => update('fundingSourceName', event.target.value)} maxLength={1000} aria-invalid={Boolean(errors.fundingSourceName)} /></Field>
              <Field label="เลขที่สัญญาทุน" error={errors.contractNumber} required><input value={form.contractNumber} onChange={(event) => update('contractNumber', event.target.value)} maxLength={1000} aria-invalid={Boolean(errors.contractNumber)} /></Field>
              <Field label="ไฟล์สัญญาทุน (PDF ไม่เกิน 20 MB)" error={errors.contractFile} required><input type="file" accept="application/pdf,.pdf" onChange={(event) => { setContractFile(event.target.files?.[0]); setErrors((current) => ({ ...current, contractFile: '' })) }} aria-invalid={Boolean(errors.contractFile)} /><small>{contractFile ? `${contractFile.name} · ${formatFileSize(contractFile.size)} (ยังไม่อัปโหลด)` : 'เลือกไฟล์ PDF จากอุปกรณ์'}</small></Field>
            </div>
          </fieldset>

          <fieldset className="form-section">
            <legend><span>04</span> ระยะเวลา งบประมาณ และรายละเอียด</legend>
            <div className="field-grid three">
              <Field label="หน่วยงานรับผิดชอบโครงการ" error={errors.responsibleProjectUnit} required><input value={form.responsibleProjectUnit} onChange={(event) => update('responsibleProjectUnit', event.target.value)} maxLength={1000} aria-invalid={Boolean(errors.responsibleProjectUnit)} /></Field>
              <Field label="หน่วยงานรับผิดชอบงบประมาณ" error={errors.responsibleBudgetUnit} required><input value={form.responsibleBudgetUnit} onChange={(event) => update('responsibleBudgetUnit', event.target.value)} maxLength={1000} aria-invalid={Boolean(errors.responsibleBudgetUnit)} /></Field>
              <Field label="งบประมาณ (บาท)" error={errors.budgetAmount} required><input inputMode="decimal" value={form.budgetAmount} onChange={(event) => update('budgetAmount', event.target.value)} aria-invalid={Boolean(errors.budgetAmount)} placeholder="0.00" /></Field>
              <Field label="วันเริ่มโครงการ (DD/MM/YYYY พ.ศ.)" error={errors.startDate} required><input inputMode="numeric" value={form.startDate} onChange={(event) => update('startDate', event.target.value)} aria-invalid={Boolean(errors.startDate)} placeholder="29/08/2569" /></Field>
              <Field label="วันสิ้นสุดโครงการ (DD/MM/YYYY พ.ศ.)" error={errors.endDate} required><input inputMode="numeric" value={form.endDate} onChange={(event) => update('endDate', event.target.value)} aria-invalid={Boolean(errors.endDate)} placeholder="29/08/2570" /></Field>
            </div>
            <div className="field-grid two prose-fields">
              <Field label="บทคัดย่อ (ไทย)" error={errors.thaiAbstract} required><textarea value={form.thaiAbstract} onChange={(event) => update('thaiAbstract', event.target.value)} maxLength={1000} aria-invalid={Boolean(errors.thaiAbstract)} /></Field>
              <Field label="บทคัดย่อ (อังกฤษ)" error={errors.englishAbstract} required><textarea value={form.englishAbstract} onChange={(event) => update('englishAbstract', event.target.value)} maxLength={1000} aria-invalid={Boolean(errors.englishAbstract)} /></Field>
              <Field label="วัตถุประสงค์โครงการ" error={errors.objectives} required><textarea value={form.objectives} onChange={(event) => update('objectives', event.target.value)} maxLength={1000} aria-invalid={Boolean(errors.objectives)} /></Field>
              <Field label="คำสำคัญ" error={errors.keywords} required><textarea value={form.keywords} onChange={(event) => update('keywords', event.target.value)} maxLength={1000} aria-invalid={Boolean(errors.keywords)} placeholder="คั่นคำด้วยเครื่องหมายจุลภาค" /></Field>
            </div>
          </fieldset>

          <footer className="form-actions"><p>ไฟล์ PDF จะถูกส่งเมื่อกด <strong>บันทึกงานวิจัย</strong> เท่านั้น</p><div><button className="button-quiet" type="button" disabled={submitting} onClick={() => onReturnToRegister()}>ยกเลิก</button><button className="button-primary" type="submit" disabled={submitting}>{submitting ? 'กำลังบันทึก…' : 'บันทึกงานวิจัย'}</button></div></footer>
        </form>
      </section>
    </main>
  )
}

function Field({ label, error, required, children }: { label: string; error?: string; required?: boolean; children: React.ReactNode }) { return <label className="field"><span>{label} {required && <i aria-hidden="true">*</i>}</span>{children}{error && <em className="field-error">{error}</em>}</label> }
function Choice({ label, checked, onChange, name }: { label: string; checked: boolean; onChange: () => void; name: string }) { return <label className="choice"><input type="radio" name={name} checked={checked} onChange={onChange} /><span>{label}</span></label> }

function validate(form: ReturnType<typeof initialForm>, members: MemberDraft[], file?: File): Errors {
  const errors: Errors = {}
  const textFields: Array<keyof ReturnType<typeof initialForm>> = ['title', 'fundingSourceName', 'contractNumber', 'responsibleProjectUnit', 'responsibleBudgetUnit', 'thaiAbstract', 'englishAbstract', 'objectives', 'keywords']
  textFields.forEach((field) => { if (!validText(String(form[field]), field === 'title')) errors[field] = 'กรุณากรอกข้อมูล 1–1,000 ตัวอักษร โดยไม่มีอักขระควบคุม' })
  if (form.isSubsidized === undefined) errors.isSubsidized = 'กรุณาเลือกลักษณะงบประมาณ'
  if (form.researchKind === 'CONTINUATION' && !/^\d+$/.test(form.continuationOfId)) errors.continuationOfId = 'กรุณาระบุรหัสโครงการต้นทางเป็นจำนวนเต็มบวก'
  if (!validDate(form.startDate)) errors.startDate = 'รูปแบบวันที่ต้องเป็น DD/MM/YYYY พ.ศ.'
  if (!validDate(form.endDate)) errors.endDate = 'รูปแบบวันที่ต้องเป็น DD/MM/YYYY พ.ศ.'
  if (!/^\d+(?:\.\d{1,2})?$/.test(form.budgetAmount) || Number(form.budgetAmount) <= 0) errors.budgetAmount = 'ต้องเป็นจำนวนมากกว่า 0 และมีทศนิยมได้ไม่เกิน 2 ตำแหน่ง'
  if (!file) errors.contractFile = 'กรุณาเลือกไฟล์สัญญา PDF'
  else if (file.size > MAX_FILE_BYTES || (!file.type.includes('pdf') && !file.name.toLowerCase().endsWith('.pdf'))) errors.contractFile = 'ไฟล์ต้องเป็น PDF และมีขนาดไม่เกิน 20 MB'
  if (!members.some((member) => member.role === 'LEAD') || !members.some((member) => member.role === 'CO_RESEARCHER') || members.some((member) => !validText(member.fullName) || !validText(member.affiliation) || !validEmail(member.email) || !validPercent(member.contributionPercent))) errors.members = 'ต้องมีหัวหน้าและผู้ร่วมโครงการอย่างน้อยอย่างละ 1 คน โดยข้อมูลแต่ละคนต้องครบและสัดส่วนอยู่ระหว่าง 0–100'
  return errors
}

function validText(value: string, title = false) { const trimmed = value.trim(); return Array.from(trimmed).length > 0 && Array.from(trimmed).length <= 1000 && !hasControlCharacter(trimmed) && (!title || !/[\n\t/]/.test(trimmed)) }
function hasControlCharacter(value: string) { return Array.from(value).some((character) => { const code = character.codePointAt(0) ?? 0; return code <= 0x1f || (code >= 0x7f && code <= 0x9f) }) }
function validEmail(value: string) { return Array.from(value.trim()).length <= 1000 && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim()) }
function validPercent(value: number) { return Number.isFinite(value) && value > 0 && value <= 100 && /^\d+(?:\.\d{1,2})?$/.test(String(value)) }
function validDate(value: string) { const match = /^(\d{2})\/(\d{2})\/(\d{4})$/.exec(value.trim()); if (!match) return false; const day = Number(match[1]); const month = Number(match[2]); const year = Number(match[3]); if (year < 2400 || year > 2800) return false; return day >= 1 && month >= 1 && month <= 12 && day <= new Date(year - 543, month, 0).getDate() }
function formatFileSize(bytes: number) { return `${(bytes / 1024 / 1024).toFixed(1)} MB` }
function apiErrorMessage(error: unknown) { switch (getAPIErrorCode(error)) { case 'TITLE_ALREADY_EXISTS': return 'ชื่องานวิจัยนี้มีอยู่แล้ว'; case 'CONTINUATION_NOT_FOUND': return 'ไม่พบโครงการต้นทางที่ระบุ'; case 'PAYLOAD_TOO_LARGE': return 'ไฟล์หรือคำขอมีขนาดเกินกำหนด'; case 'UNSUPPORTED_MEDIA_TYPE': return 'สัญญาต้องเป็นไฟล์ PDF'; case 'VALIDATION_ERROR': return 'ข้อมูลไม่ผ่านการตรวจสอบของระบบ'; default: return 'ยังไม่สามารถบันทึกได้ โปรดลองอีกครั้ง' } }
