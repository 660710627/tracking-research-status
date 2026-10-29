import type { Research } from './research'
import { Modal } from './ResearchControls'

export default function ProcessLog({research,steps,onClose}:{research:Research;steps:string[];onClose:()=>void}) {
  const entries=[...(research.processHistory??[])].reverse()
  return <Modal title="ประวัติการเปลี่ยนกระบวนการ" onCancel={onClose} className="process-log">
    <p className="confirmation-project">{research.title}<small>{research.contract}</small></p>
    {entries.length?<div className="table-wrap"><table><caption className="sr-only">ประวัติการเปลี่ยนกระบวนการ เรียงจากล่าสุด</caption><thead><tr><th>วันที่และเวลา</th><th>ผู้เปลี่ยน</th><th>การเปลี่ยนแปลง</th></tr></thead><tbody>{entries.map((entry,index)=><tr key={index}>
      <td><time dateTime={entry.recordedAt}>{new Date(entry.recordedAt).toLocaleString('th-TH')}</time></td>
      <td>{entry.changedBy??'ไม่ได้บันทึกชื่อผู้เปลี่ยน'}{entry.changedRole&&<small>{entry.changedRole}</small>}</td>
      <td><dl><dt>เดิม</dt><dd>{entry.fromProcess?steps[entry.fromProcess-1]:'ไม่มีข้อมูล'}</dd><dt>ใหม่</dt><dd>{steps[entry.process-1]}</dd><dt>กระบวนการย่อย</dt><dd>{entry.subprocess}</dd></dl>{entry.note&&<p>หมายเหตุ: {entry.note}</p>}{entry.file&&<p>ไฟล์แนบ: {entry.file.name}</p>}</td>
    </tr>)}</tbody></table></div>:<p className="process-log-empty">ยังไม่มีประวัติการเปลี่ยนกระบวนการ</p>}
    <div className="form-actions"><button type="button" data-cancel className="secondary" onClick={onClose}>ปิด</button></div>
  </Modal>
}
