import type { Research } from '../api/client'
import { CatalogHeading } from './CatalogHeading'

type ResearchListProps = {
  researches: Research[]
  highlightedID?: number
  onEdit: (research: Research, opener: HTMLButtonElement) => void
  onDelete: (research: Research, opener: HTMLButtonElement) => void
}

export function ResearchList({ researches, highlightedID, onEdit, onDelete }: ResearchListProps) {
  const duplicateTitles = countDuplicateTitles(researches)

  return (
    <section className="catalog" aria-labelledby="research-list-title">
      <CatalogHeading id="research-list-title" label={`งานวิจัย ${researches.length.toLocaleString('th-TH')} รายการ`} />

      <div className="list-table">
        <div className="list-header" aria-hidden="true">
          <span>ชื่องานวิจัย</span>
          <span>รายละเอียด</span>
          <span>จัดการ</span>
        </div>
        <div className="list-body">
          {researches.map((research) => {
            const showID = (duplicateTitles.get(research.title) ?? 0) > 1
            return (
              <article className={`research-row${research.id === highlightedID ? ' newly-created' : ''}`} key={research.id}>
                <div className="research-identity">
                  <h3>{research.title}</h3>
                  {showID && <p className="research-id">รหัสงานวิจัย #{research.id}</p>}
                </div>
                <p className="research-description">{research.description}</p>
                <div className="research-actions">
                  <button className="row-action" type="button" onClick={(event) => onEdit(research, event.currentTarget)}>
                    แก้ไข
                  </button>
                  <button className="row-action row-action-danger" type="button" onClick={(event) => onDelete(research, event.currentTarget)}>
                    ลบ
                  </button>
                </div>
              </article>
            )
          })}
        </div>
      </div>
    </section>
  )
}

function countDuplicateTitles(researches: Research[]) {
  const counts = new Map<string, number>()
  for (const research of researches) {
    counts.set(research.title, (counts.get(research.title) ?? 0) + 1)
  }
  return counts
}
