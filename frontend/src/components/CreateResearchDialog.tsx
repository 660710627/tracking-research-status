import { useEffect, useRef, useState, type FormEvent } from 'react'
import { createResearch, listResearches, type Research } from '../api/client'
import { getAPIErrorCode } from '../api/errors'

type FieldErrors = {
  title?: string
  description?: string
  continuation?: string
  form?: string
}

type CreateResearchDialogProps = {
  researches: Research[]
  onClose: () => void
  onCreated: (research: Research) => void
}

export function CreateResearchDialog({ researches, onClose, onCreated }: CreateResearchDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const titleRef = useRef<HTMLInputElement>(null)
  const descriptionRef = useRef<HTMLTextAreaElement>(null)
  const continuationRef = useRef<HTMLSelectElement>(null)
  const submissionInFlightRef = useRef(false)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [continuationID, setContinuationID] = useState('')
  const [continuationOptions, setContinuationOptions] = useState(researches)
  const [errors, setErrors] = useState<FieldErrors>({})
  const [submitting, setSubmitting] = useState(false)

  const dirty = title.length > 0 || description.length > 0 || continuationID !== ''

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
    if (dirty && !window.confirm('ยกเลิกและทิ้งข้อมูลที่กรอกไว้หรือไม่')) return
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
      const result = await createResearch({
        body: {
          title,
          description,
          continuationOfId: continuationID === '' ? null : Number(continuationID),
        },
      })
      if (result.error) {
        await handleAPIError(result.error)
        return
      }
      onCreated(result.data)
    } catch (error) {
      await handleAPIError(error)
    } finally {
      submissionInFlightRef.current = false
      setSubmitting(false)
    }
  }

  async function handleAPIError(error: unknown) {
    const code = getAPIErrorCode(error)
    switch (code) {
      case 'TITLE_ALREADY_EXISTS':
        setErrors({ title: 'มีชื่องานวิจัยนี้แล้ว หากเป็นงานต่อเนื่องให้เลือกงานวิจัยต้นทาง' })
        titleRef.current?.focus()
        break
      case 'CONTINUATION_NOT_FOUND':
        setContinuationID('')
        setErrors({ continuation: 'งานวิจัยต้นทางไม่มีอยู่แล้ว กรุณาเลือกใหม่' })
        await refreshContinuationOptions()
        continuationRef.current?.focus()
        break
      case 'VALIDATION_ERROR':
        setErrors({ form: 'ข้อมูลยังไม่ผ่านการตรวจสอบ กรุณาตรวจชื่อและรายละเอียดอีกครั้ง' })
        titleRef.current?.focus()
        break
      case 'PAYLOAD_TOO_LARGE':
        setErrors({ form: 'ข้อมูลมีขนาดใหญ่เกินไป กรุณาลดความยาวของชื่อหรือรายละเอียด' })
        break
      default:
        setErrors({ form: 'ยังไม่สามารถเพิ่มงานวิจัยได้ กรุณาตรวจสอบการเชื่อมต่อแล้วลองอีกครั้ง' })
    }
  }

  async function refreshContinuationOptions() {
    try {
      const result = await listResearches()
      if (!result.error) setContinuationOptions(result.data)
    } catch {
      // The actionable continuation error remains visible when refresh fails.
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
      aria-labelledby="create-dialog-title"
      aria-describedby="create-dialog-summary"
      onCancel={(event) => {
        event.preventDefault()
        requestClose()
      }}
    >
      <div className="dialog-heading">
        <div>
          <p className="dialog-kicker">รายการใหม่</p>
          <h2 id="create-dialog-title">เพิ่มงานวิจัย</h2>
          <p id="create-dialog-summary">กรอกข้อมูลที่จำเป็นและระบุว่าเป็นงานต้นฉบับหรืองานต่อเนื่อง</p>
        </div>
        <button className="dialog-close" type="button" onClick={requestClose} disabled={submitting} aria-label="ปิดหน้าต่างเพิ่มงานวิจัย">
          <span aria-hidden="true">×</span>
        </button>
      </div>

      <form className="research-form" onSubmit={handleSubmit} noValidate>
        {errors.form && <div className="form-alert" role="alert">{errors.form}</div>}

        <div className="field-group">
          <div className="field-label-row">
            <label htmlFor="research-title">ชื่องานวิจัย <span>จำเป็น</span></label>
            <span className="character-count" aria-hidden="true">{unicodeLength(title).toLocaleString('th-TH')} / 200</span>
          </div>
          <input
            ref={titleRef}
            id="research-title"
            name="title"
            type="text"
            value={title}
            onChange={(event) => {
              setTitle(event.target.value)
              if (errors.title || errors.form) setErrors((current) => ({ ...current, title: undefined, form: undefined }))
            }}
            aria-invalid={Boolean(errors.title)}
            aria-describedby={`research-title-help${errors.title ? ' research-title-error' : ''}`}
            placeholder="เช่น การพัฒนาระบบติดตามงานวิจัย"
            autoComplete="off"
          />
          <p id="research-title-help" className="field-help">1–200 ตัวอักษร และห้ามมีเครื่องหมาย /</p>
          {errors.title && <p id="research-title-error" className="field-error">{errors.title}</p>}
        </div>

        <div className="field-group">
          <div className="field-label-row">
            <label htmlFor="research-description">รายละเอียด <span>จำเป็น</span></label>
            <span className="character-count" aria-hidden="true">{unicodeLength(description).toLocaleString('th-TH')} / 5,000</span>
          </div>
          <textarea
            ref={descriptionRef}
            id="research-description"
            name="description"
            rows={6}
            value={description}
            onChange={(event) => {
              setDescription(event.target.value)
              if (errors.description || errors.form) setErrors((current) => ({ ...current, description: undefined, form: undefined }))
            }}
            aria-invalid={Boolean(errors.description)}
            aria-describedby={`research-description-help${errors.description ? ' research-description-error' : ''}`}
            placeholder="อธิบายวัตถุประสงค์หรือขอบเขตของงานวิจัย"
          />
          <p id="research-description-help" className="field-help">1–5,000 ตัวอักษร สามารถขึ้นบรรทัดใหม่ได้</p>
          {errors.description && <p id="research-description-error" className="field-error">{errors.description}</p>}
        </div>

        <div className="field-group">
          <label htmlFor="research-continuation">งานวิจัยต้นทาง <span>จำเป็น</span></label>
          <select
            ref={continuationRef}
            id="research-continuation"
            name="continuationOfId"
            value={continuationID}
            onChange={(event) => {
              setContinuationID(event.target.value)
              if (errors.continuation) setErrors((current) => ({ ...current, continuation: undefined }))
            }}
            aria-invalid={Boolean(errors.continuation)}
            aria-describedby={`research-continuation-help${errors.continuation ? ' research-continuation-error' : ''}`}
          >
            <option value="">งานวิจัยต้นฉบับ — ไม่มีงานวิจัยต้นทาง</option>
            {continuationOptions.map((research) => (
              <option value={research.id} key={research.id}>{research.title} — รหัส #{research.id}</option>
            ))}
          </select>
          <p id="research-continuation-help" className="field-help">เลือกงานที่นำมาต่อยอด หรือเลือกงานวิจัยต้นฉบับ</p>
          {errors.continuation && <p id="research-continuation-error" className="field-error">{errors.continuation}</p>}
        </div>

        <div className="dialog-actions">
          <button className="text-button" type="button" onClick={requestClose} disabled={submitting}>ยกเลิก</button>
          <button className="primary-button" type="submit" disabled={submitting}>
            {submitting ? 'กำลังบันทึก…' : 'เพิ่มงานวิจัย'}
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
