const months = ['ม.ค.','ก.พ.','มี.ค.','เม.ย.','พ.ค.','มิ.ย.','ก.ค.','ส.ค.','ก.ย.','ต.ค.','พ.ย.','ธ.ค.']

export function toCalendarDate(value:string):string {
  if (/^\d{4}-\d{2}-\d{2}$/.test(value)) return value
  const numeric = value.match(/^(\d{1,2})\/(\d{1,2})\/(\d{4})$/)
  const thai = value.trim().split(/\s+/)
  const day = numeric ? Number(numeric[1]) : Number(thai[0])
  const month = numeric ? Number(numeric[2]) : months.indexOf(thai[1]) + 1
  const year = Number(numeric ? numeric[3] : thai[2]) - 543
  if (!day || !month || !Number.isFinite(year) || year < 1) return ''
  return `${String(year).padStart(4,'0')}-${String(month).padStart(2,'0')}-${String(day).padStart(2,'0')}`
}

export function fromCalendarDate(value:string):string {
  if (!value) return ''
  const [year,month,day] = value.split('-').map(Number)
  return `${day} ${months[month-1]} ${year+543}`
}
