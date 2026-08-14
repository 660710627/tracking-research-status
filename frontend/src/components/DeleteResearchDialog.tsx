import { useEffect, useRef, useState } from 'react'
import { deleteResearch, type Research } from '../api/client'
import { getAPIErrorCode } from '../api/errors'

type DeleteResearchDialogProps = {
  research: Research
  showDuplicateContext: boolean
  onClose: () => void
  onDeleted: (research: Research) => void
  onNotFound: () => void
}

export function DeleteResearchDialog({ research, showDuplicateContext, onClose, onDeleted, onNotFound }: DeleteResearchDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const cancelRef = useRef<HTMLButtonElement>(null)
  const requestInFlightRef = useRef(false)
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState<string>()

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return
    dialog.showModal()
    cancelRef.current?.focus()
    return () => {
      if (dialog.open) dialog.close()
    }
  }, [])

  function requestClose() {
    if (!requestInFlightRef.current) onClose()
  }

  async function handleDelete() {
    if (requestInFlightRef.current) return
    requestInFlightRef.current = true
    setDeleting(true)
    setError(undefined)

    try {
      const result = await deleteResearch({ path: { id: research.id } })
      if (result.error) {
        handleAPIError(result.error)
        return
      }
      onDeleted(research)
    } catch (requestError) {
      handleAPIError(requestError)
    } finally {
      requestInFlightRef.current = false
      setDeleting(false)
    }
  }

  function handleAPIError(requestError: unknown) {
    const code = getAPIErrorCode(requestError)
    switch (code) {
      case 'RESEARCH_NOT_FOUND':
        onNotFound()
        break
      case 'RESEARCH_HAS_CONTINUATIONS':
        setError('ยังลบงานวิจัยนี้ไม่ได้ เพราะมีงานวิจัยต่อเนื่องอ้างถึงอยู่')
        cancelRef.current?.focus()
        break
      default:
        setError('ยังไม่สามารถลบงานวิจัยได้ กรุณาตรวจสอบการเชื่อมต่อแล้วลองอีกครั้ง')
    }
  }

  return (
    <dialog
      ref={dialogRef}
      className="research-dialog delete-research-dialog"
      aria-labelledby="delete-dialog-title"
      aria-describedby="delete-dialog-summary delete-dialog-warning"
      aria-busy={deleting}
      onCancel={(event) => {
        event.preventDefault()
        requestClose()
      }}
    >
      <div className="dialog-heading delete-dialog-heading">
        <div>
          <p className="dialog-kicker delete-kicker">ยืนยันการลบ</p>
          <h2 id="delete-dialog-title">ลบงานวิจัย</h2>
          <p id="delete-dialog-summary">คุณกำลังจะลบ “{research.title}”</p>
        </div>
      </div>

      <div className="delete-dialog-content">
        {showDuplicateContext && (
          <div className="delete-record-context">
            <span>รหัสงานวิจัย #{research.id}</span>
            <p>{research.description}</p>
          </div>
        )}

        <p id="delete-dialog-warning" className="destructive-note">การลบนี้ไม่สามารถย้อนกลับได้</p>
        {error && <div className="form-alert" role="alert">{error}</div>}

        <div className="dialog-actions">
          <button ref={cancelRef} className="text-button" type="button" onClick={requestClose} disabled={deleting}>
            ยกเลิก
          </button>
          <button className="danger-button" type="button" onClick={() => void handleDelete()} disabled={deleting}>
            {deleting ? 'กำลังลบ…' : 'ลบงานวิจัย'}
          </button>
        </div>
      </div>
    </dialog>
  )
}
