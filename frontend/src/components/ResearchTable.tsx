import type { Research } from '../api'

export function ResearchTable({ items, changedId }: { items: Research[]; changedId?: number }) {
  return (
    <div className="table-scroll" role="region" aria-label="ตารางงานวิจัย เลื่อนแนวนอนเพื่อดูทุกคอลัมน์" tabIndex={0}>
      <table>
        <caption className="sr-only">งานวิจัยเรียงตามชื่อ และลำดับรายการเมื่อชื่อซ้ำ</caption>
        <thead><tr><th scope="col">เลขสัญญาทุน</th><th scope="col">ประเภทโครงการ</th><th scope="col">ชื่อโครงการ</th><th scope="col">สถานะ</th><th scope="col">กระบวนการ</th></tr></thead>
        <tbody>{items.map((item) => (
          <tr key={item.id} className={item.id === changedId ? 'changed' : undefined}>
            <th scope="row">{item.contractNumber}{item.id === changedId && <span className="changed-label">เปลี่ยนแปลงล่าสุด</span>}</th>
            <td>{item.researchKind === 'CONTINUATION' ? 'โครงการต่อเนื่อง' : 'โครงการหลัก'}</td>
            <td className="project-title">{item.title}</td>
            <td>{item.status}</td><td>{item.process}</td>
          </tr>
        ))}</tbody>
      </table>
    </div>
  )
}
