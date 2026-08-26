import type { ReactNode } from 'react'
import type { Research } from '../api/client'
import { CatalogHeading } from './CatalogHeading'

type ResearchListProps = {
  researches: Research[]
  totalCount: number
  filtersActive: boolean
  onClearFilters: () => void
  filterControls: ReactNode
  highlightedID?: number
  onEdit: (research: Research, opener: HTMLButtonElement) => void
  onDelete: (research: Research, opener: HTMLButtonElement) => void
}

export function ResearchList({ researches, totalCount, filtersActive, onClearFilters, filterControls, highlightedID, onEdit, onDelete }: ResearchListProps) {
  const label = researches.length === totalCount
    ? `งานวิจัย ${researches.length.toLocaleString('th-TH')} รายการ`
    : `ผลการค้นหา ${researches.length.toLocaleString('th-TH')} จาก ${totalCount.toLocaleString('th-TH')} รายการ`

  return (
    <section className="catalog" aria-labelledby="research-list-title">
      <CatalogHeading id="research-list-title" label={label} />
      {filterControls}

      {researches.length === 0 ? (
        <div className="filtered-empty" role="status">
          <h3>ไม่พบงานวิจัยที่ตรงกับเงื่อนไข</h3>
          <p>ลองเปลี่ยนคำค้นหาหรือตัวกรอง แล้วตรวจสอบอีกครั้ง</p>
          {filtersActive && <button className="secondary-button" type="button" onClick={onClearFilters}>ล้างตัวกรอง</button>}
        </div>
      ) : (

      <div className="research-ledger">
        <table className="research-table">
          <caption className="sr-only">รายการงานวิจัยทั้งหมด</caption>
          <thead>
            <tr>
              <th scope="col">รหัส</th>
              <th scope="col">งานวิจัย</th>
              <th scope="col">รายละเอียด</th>
              <th scope="col">สถานะ</th>
              <th scope="col">กระบวนการ</th>
              <th scope="col"><span className="sr-only">จัดการรายการ</span></th>
            </tr>
          </thead>
          <tbody>
          {researches.map((research) => (
            <tr key={research.id} className={research.id === highlightedID ? 'newly-created' : ''} aria-label={`งานวิจัยรหัส ${research.id}: ${research.title}`}>
                <td className="research-table-id">
                  <div className="research-register-number" aria-label={`รหัสงานวิจัย ${research.id}`}>
                  <span>research</span>
                  <strong>{String(research.id).padStart(3, '0')}</strong>
                  </div>
                </td>
                <td className="research-table-title">
                  <p className="research-type">{research.continuationOfId === null ? 'โครงการหลัก' : `โครงการต่อเนื่อง · #${research.continuationOfId}`}</p>
                  <h3>{research.title}</h3>
                </td>
                <td className="research-table-description">{research.description}</td>
                <td><span className="research-table-status">{research.status}</span></td>
                <td><span className="research-table-process">{research.process}</span></td>
                <td>
                  <div className="research-actions">
                  <button className="row-action" type="button" onClick={(event) => onEdit(research, event.currentTarget)}>
                    แก้ไข
                  </button>
                  <button className="row-action row-action-danger" type="button" onClick={(event) => onDelete(research, event.currentTarget)}>
                    ลบ
                  </button>
                  </div>
                </td>
            </tr>
          ))}
          </tbody>
        </table>
      </div>
      )}
    </section>
  )
}
