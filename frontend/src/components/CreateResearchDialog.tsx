import { useEffect, useRef, useState, type FormEvent } from 'react'
import { createResearch, type Research } from '../api/client'
import { getAPIErrorCode } from '../api/errors'

type FieldErrors = {
  title?: string
  description?: string
  continuation?: string
  form?: string
}

type CreateResearchDialogProps = {
  onClose: () => void
  onCreated: (research: Research) => void
}

export function CreateResearchDialog({ onClose, onCreated }: CreateResearchDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const titleRef = useRef<HTMLInputElement>(null)
  const descriptionRef = useRef<HTMLTextAreaElement>(null)
  const continuationRef = useRef<HTMLInputElement>(null)
  const submissionInFlightRef = useRef(false)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [researchKind, setResearchKind] = useState<'budget' | 'continuation'>('budget')
  const [continuationID, setContinuationID] = useState('')
  const [errors, setErrors] = useState<FieldErrors>({})
  const [submitting, setSubmitting] = useState(false)

  const dirty = title.length > 0 || description.length > 0 || researchKind === 'continuation' || continuationID !== ''

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
    const validationErrors = validateResearch(normalizedTitle, normalizedDescription, researchKind, continuationID)
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
          continuationOfId: researchKind === 'budget' ? null : Number(continuationID),
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

  function focusFirstError(validationErrors: FieldErrors) {
    if (validationErrors.title) titleRef.current?.focus()
    else if (validationErrors.description) descriptionRef.current?.focus()
    else if (validationErrors.continuation) continuationRef.current?.focus()
  }

  return (
    <dialog
      ref={dialogRef}
      className="research-dialog create-research-dialog"
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
          <p id="create-dialog-summary">บันทึกข้อมูลหลักก่อน แล้วระบุว่างานนี้เริ่มต้นเองหรือต่อยอดจากรายการใด</p>
        </div>
        <button className="dialog-close" type="button" onClick={requestClose} disabled={submitting} aria-label="ปิดหน้าต่างเพิ่มงานวิจัย">
          <span className="dialog-close-icon" aria-hidden="true">×</span>
          <span className="dialog-close-label">ปิด</span>
        </button>
      </div>

      <form className="research-form" onSubmit={handleSubmit} noValidate aria-busy={submitting}>
        {errors.form && <div className="form-alert" role="alert">{errors.form}</div>}

        <div className="create-intake-note">
          <span aria-hidden="true">01</span>
          <p><strong>เริ่มต้นที่ข้อมูลสำคัญ</strong> รหัสงานวิจัยจะถูกสร้างโดยระบบหลังบันทึกสำเร็จ</p>
        </div>

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
            disabled={submitting}
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
            disabled={submitting}
          />
          <p id="research-description-help" className="field-help">1–5,000 ตัวอักษร สามารถขึ้นบรรทัดใหม่ได้</p>
          {errors.description && <p id="research-description-error" className="field-error">{errors.description}</p>}
        </div>

        <fieldset className="continuation-choice">
          <legend>ลักษณะการดำเนินงาน</legend>
          <label className="research-kind-option">
            <input
              type="radio"
              name="researchKind"
              value="budget"
              checked={researchKind === 'budget'}
              onChange={() => {
                setResearchKind('budget')
                setContinuationID('')
                if (errors.continuation) setErrors((current) => ({ ...current, continuation: undefined }))
              }}
              disabled={submitting}
            />
            <span>โครงการที่เป็นงบประมาณ</span>
          </label>
          <label className="research-kind-option">
            <input
              type="radio"
              name="researchKind"
              value="continuation"
              checked={researchKind === 'continuation'}
              onChange={() => setResearchKind('continuation')}
              disabled={submitting}
              aria-controls="research-continuation"
            />
            <span>โครงการต่อเนื่อง</span>
          </label>
          {researchKind === 'continuation' && (
            <div className="continuation-control">
              <label htmlFor="research-continuation">ระบุรหัสโครงการก่อนหน้า <span aria-hidden="true">*</span></label>
              <input
                ref={continuationRef}
                id="research-continuation"
                name="continuationOfId"
                type="number"
                min="1"
                step="1"
                inputMode="numeric"
                value={continuationID}
                onChange={(event) => {
                  setContinuationID(event.target.value)
                  if (errors.continuation) setErrors((current) => ({ ...current, continuation: undefined }))
                }}
                aria-invalid={Boolean(errors.continuation)}
                aria-describedby={errors.continuation ? 'research-continuation-error' : undefined}
                autoComplete="off"
                disabled={submitting}
              />
              {errors.continuation && <p id="research-continuation-error" className="field-error">{errors.continuation}</p>}
            </div>
          )}
        </fieldset>

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

function validateResearch(title: string, description: string, researchKind: 'budget' | 'continuation', continuationID: string): FieldErrors {
  const errors: FieldErrors = {}
  const titleLength = unicodeLength(title)
  const descriptionLength = unicodeLength(description)

  if (titleLength === 0) errors.title = 'กรุณากรอกชื่องานวิจัย'
  else if (titleLength > 200) errors.title = 'ชื่องานวิจัยต้องไม่เกิน 200 ตัวอักษร'
  else if (title.includes('/') || containsForbiddenControl(title, false)) errors.title = 'ชื่องานวิจัยห้ามมี / บรรทัดใหม่ แท็บ หรืออักขระควบคุม'

  if (descriptionLength === 0) errors.description = 'กรุณากรอกรายละเอียด'
  else if (descriptionLength > 5000) errors.description = 'รายละเอียดต้องไม่เกิน 5,000 ตัวอักษร'
  else if (containsForbiddenControl(description, true)) errors.description = 'รายละเอียดมีอักขระที่ระบบไม่รองรับ'

  if (researchKind === 'continuation' && (!/^\d+$/.test(continuationID) || Number(continuationID) < 1 || !Number.isSafeInteger(Number(continuationID)))) {
    errors.continuation = 'กรุณาระบุรหัสโครงการก่อนหน้าที่เป็นจำนวนเต็มบวก'
  }

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
