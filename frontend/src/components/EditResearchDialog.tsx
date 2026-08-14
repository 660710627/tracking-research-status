import { useEffect, useRef, useState, type FormEvent } from 'react'
import { updateResearch, type Research } from '../api/client'
import { getAPIErrorCode } from '../api/errors'

type FieldErrors = {
  title?: string
  description?: string
  form?: string
}

type EditResearchDialogProps = {
  research: Research
  onClose: () => void
  onNotFound: () => void
  onUpdated: (research: Research) => void
}

export function EditResearchDialog({ research, onClose, onNotFound, onUpdated }: EditResearchDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const titleRef = useRef<HTMLInputElement>(null)
  const descriptionRef = useRef<HTMLTextAreaElement>(null)
  const submissionInFlightRef = useRef(false)
  const [title, setTitle] = useState(research.title)
  const [description, setDescription] = useState(research.description)
  const [errors, setErrors] = useState<FieldErrors>({})
  const [submitting, setSubmitting] = useState(false)

  const dirty = title !== research.title || description !== research.description

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return
    dialog.showModal()
    titleRef.current?.focus()
    return () => {
      if (dialog.open) dialog.close()
    }
  }, [])

  function requestClose() {
    if (submissionInFlightRef.current) return
    if (dirty && !window.confirm('ยกเลิกและทิ้งข้อมูลที่แก้ไขไว้หรือไม่')) return
    onClose()
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (submissionInFlightRef.current) return

    const normalizedTitle = title.trim()
    const normalizedDescription = description.replaceAll('\r\n', '\n').trim()
    const validationErrors = validateResearch(normalizedTitle, normalizedDescription)
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors)
      focusFirstError(validationErrors)
      return
    }

    submissionInFlightRef.current = true
    setErrors({})
    setSubmitting(true)
    try {
      const result = await updateResearch({
        path: { id: research.id },
        body: { title: normalizedTitle, description: normalizedDescription },
      })
      if (result.error) {
        handleAPIError(result.error)
        return
      }
      onUpdated(result.data)
    } catch (error) {
      handleAPIError(error)
    } finally {
      submissionInFlightRef.current = false
      setSubmitting(false)
    }
  }

  function handleAPIError(error: unknown) {
    const code = getAPIErrorCode(error)
    switch (code) {
      case 'TITLE_ALREADY_EXISTS':
        setErrors({ title: 'มีชื่องานวิจัยนี้แล้ว กรุณาใช้ชื่ออื่น' })
        titleRef.current?.focus()
        break
      case 'VALIDATION_ERROR':
        setErrors({ form: 'ข้อมูลยังไม่ผ่านการตรวจสอบ กรุณาตรวจชื่อและรายละเอียดอีกครั้ง' })
        titleRef.current?.focus()
        break
      case 'RESEARCH_NOT_FOUND':
        onNotFound()
        break
      case 'PAYLOAD_TOO_LARGE':
        setErrors({ form: 'ข้อมูลมีขนาดใหญ่เกินไป กรุณาลดความยาวของชื่อหรือรายละเอียด' })
        break
      default:
        setErrors({ form: 'ยังไม่สามารถบันทึกการแก้ไขได้ กรุณาตรวจสอบการเชื่อมต่อแล้วลองอีกครั้ง' })
    }
  }

  function focusFirstError(validationErrors: FieldErrors) {
    if (validationErrors.title) titleRef.current?.focus()
    else if (validationErrors.description) descriptionRef.current?.focus()
  }

  return (
    <dialog
      ref={dialogRef}
      className="research-dialog"
      aria-labelledby="edit-dialog-title"
      aria-describedby="edit-dialog-summary"
      onCancel={(event) => {
        event.preventDefault()
        requestClose()
      }}
    >
      <div className="dialog-heading">
        <div>
          <p className="dialog-kicker">แก้ไขรายการ #{research.id}</p>
          <h2 id="edit-dialog-title">แก้ไขงานวิจัย</h2>
          <p id="edit-dialog-summary">แก้ไขชื่อและรายละเอียด โดยข้อมูลอ้างอิงของรายการจะคงเดิม</p>
        </div>
        <button className="dialog-close" type="button" onClick={requestClose} disabled={submitting} aria-label="ปิดหน้าต่างแก้ไขงานวิจัย">
          <span aria-hidden="true">×</span>
        </button>
      </div>

      <form className="research-form" onSubmit={handleSubmit} noValidate aria-busy={submitting}>
        <div className="immutable-note">
          <span>รายการอ้างอิง</span>
          <strong>{research.continuationOfId === null ? 'งานวิจัยต้นฉบับ' : `งานต่อเนื่องจากรหัส #${research.continuationOfId}`}</strong>
          <small>ID และงานวิจัยต้นทางไม่สามารถเปลี่ยนผ่านแบบฟอร์มนี้</small>
        </div>

        {errors.form && <div className="form-alert" role="alert">{errors.form}</div>}

        <div className="field-group">
          <div className="field-label-row">
            <label htmlFor="edit-research-title">ชื่องานวิจัย <span>จำเป็น</span></label>
            <span className="character-count" aria-hidden="true">{unicodeLength(title).toLocaleString('th-TH')} / 200</span>
          </div>
          <input
            ref={titleRef}
            id="edit-research-title"
            name="title"
            type="text"
            value={title}
            onChange={(event) => {
              setTitle(event.target.value)
              if (errors.title || errors.form) setErrors((current) => ({ ...current, title: undefined, form: undefined }))
            }}
            aria-invalid={Boolean(errors.title)}
            aria-describedby={`edit-research-title-help${errors.title ? ' edit-research-title-error' : ''}`}
            autoComplete="off"
          />
          <p id="edit-research-title-help" className="field-help">1–200 ตัวอักษร และห้ามมีเครื่องหมาย /</p>
          {errors.title && <p id="edit-research-title-error" className="field-error">{errors.title}</p>}
        </div>

        <div className="field-group">
          <div className="field-label-row">
            <label htmlFor="edit-research-description">รายละเอียด <span>จำเป็น</span></label>
            <span className="character-count" aria-hidden="true">{unicodeLength(description).toLocaleString('th-TH')} / 5,000</span>
          </div>
          <textarea
            ref={descriptionRef}
            id="edit-research-description"
            name="description"
            rows={6}
            value={description}
            onChange={(event) => {
              setDescription(event.target.value)
              if (errors.description || errors.form) setErrors((current) => ({ ...current, description: undefined, form: undefined }))
            }}
            aria-invalid={Boolean(errors.description)}
            aria-describedby={`edit-research-description-help${errors.description ? ' edit-research-description-error' : ''}`}
          />
          <p id="edit-research-description-help" className="field-help">1–5,000 ตัวอักษร สามารถขึ้นบรรทัดใหม่ได้</p>
          {errors.description && <p id="edit-research-description-error" className="field-error">{errors.description}</p>}
        </div>

        <div className="dialog-actions">
          <button className="text-button" type="button" onClick={requestClose} disabled={submitting}>ยกเลิก</button>
          <button className="primary-button" type="submit" disabled={submitting}>
            {submitting ? 'กำลังบันทึก…' : 'บันทึกการแก้ไข'}
          </button>
        </div>
      </form>
    </dialog>
  )
}

function validateResearch(title: string, description: string): FieldErrors {
  const errors: FieldErrors = {}
  const titleLength = unicodeLength(title)
  const descriptionLength = unicodeLength(description)

  if (titleLength === 0) errors.title = 'กรุณากรอกชื่องานวิจัย'
  else if (titleLength > 200) errors.title = 'ชื่องานวิจัยต้องไม่เกิน 200 ตัวอักษร'
  else if (title.includes('/') || containsForbiddenControl(title, false)) errors.title = 'ชื่องานวิจัยห้ามมี / บรรทัดใหม่ แท็บ หรืออักขระควบคุม'

  if (descriptionLength === 0) errors.description = 'กรุณากรอกรายละเอียด'
  else if (descriptionLength > 5000) errors.description = 'รายละเอียดต้องไม่เกิน 5,000 ตัวอักษร'
  else if (containsForbiddenControl(description, true)) errors.description = 'รายละเอียดมีอักขระที่ระบบไม่รองรับ'

  return errors
}

function containsForbiddenControl(value: string, allowTextLayout: boolean) {
  for (const character of value) {
    const code = character.codePointAt(0) ?? 0
    const control = code <= 0x1f || (code >= 0x7f && code <= 0x9f)
    if (control && !(allowTextLayout && (character === '\n' || character === '\t'))) return true
  }
  return false
}

function unicodeLength(value: string) {
  return Array.from(value).length
}
