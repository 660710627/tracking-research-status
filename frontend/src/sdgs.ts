// Short Thai labels for the 17 UN Sustainable Development Goals.
// Reference: https://thailand.un.org/th/sdgs
export const sdgGoals = [
  'ขจัดความยากจน', 'ขจัดความหิวโหย', 'สุขภาพและความเป็นอยู่ที่ดี',
  'การศึกษาที่มีคุณภาพ', 'ความเท่าเทียมทางเพศ', 'น้ำสะอาดและการสุขาภิบาล',
  'พลังงานสะอาดที่ทุกคนเข้าถึงได้', 'งานที่มีคุณค่าและการเติบโตทางเศรษฐกิจ',
  'อุตสาหกรรม นวัตกรรม และโครงสร้างพื้นฐาน', 'ลดความเหลื่อมล้ำ',
  'เมืองและชุมชนที่ยั่งยืน', 'การผลิตและการบริโภคที่ยั่งยืน',
  'การรับมือการเปลี่ยนแปลงสภาพภูมิอากาศ', 'ทรัพยากรทางทะเล',
  'ระบบนิเวศบนบก', 'สันติภาพ ความยุติธรรม และสถาบันที่เข้มแข็ง',
  'ความร่วมมือเพื่อการพัฒนาที่ยั่งยืน',
].map((name,index)=>({id:index+1,name}))

export function normalizeSdgs(ids:number[]):number[] {
  if(!Array.isArray(ids)||ids.some(id=>!Number.isInteger(id)||id<1||id>17))throw new Error('เลือก SDGs จากเป้าหมายที่ 1–17 เท่านั้น')
  if(ids.length===0)throw new Error('กรุณาเลือก SDGs อย่างน้อย 1 เป้าหมาย')
  return [...new Set(ids)].sort((a,b)=>a-b)
}
