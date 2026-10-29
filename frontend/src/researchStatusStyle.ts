export function researchStatusClass(status: string): string {
  switch (status) {
    case 'กำลังดำเนินการ(ขยายเวลาครั้งที่ 1)': return 'status warn'
    case 'กำลังดำเนินการ(ขยายเวลาครั้งที่ 2)': return 'status research-extended-twice'
    case 'กำลังดำเนินการ(ขยายเวลามากกว่า 2 ครั้ง)': return 'status research-extended-many'
    case 'โครงการเสร็จสิ้น': return 'status research-completed'
    case 'ยุติโครงการ': return 'status research-terminated'
    default: return 'status'
  }
}
