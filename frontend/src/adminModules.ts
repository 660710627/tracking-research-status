export const adminModules = [
  {id:'approvals',title:'อนุมัติบัญชีผู้ใช้งาน',description:'ตรวจสอบบัญชีที่เข้าสู่ระบบครั้งแรกก่อนอนุมัติการใช้งาน'},
  {id:'users',title:'บัญชีผู้ใช้งาน',description:'จัดการบัญชีที่อนุมัติแล้วและบัญชีที่ถูกระงับ'},
  {id:'academic-titles',title:'ตำแหน่งทางวิชาการ',description:'รายการตำแหน่งสำหรับเลือกใช้กับข้อมูลนักวิจัย'},
  {id:'about',title:'เกี่ยวกับเรา',description:'ข้อมูลสำนักงานบริหารการวิจัย นวัตกรรมและการสร้างสรรค์'},
  {id:'organization',title:'หน่วยงาน',description:'ข้อมูลมหาวิทยาลัย สำนักงาน ชื่อระบบ และที่อยู่'},
  {id:'logs',title:'รายงานข้อมูล log',description:'ตรวจสอบประวัติการเปลี่ยนแปลงข้อมูลในระบบ'},
] as const
export type AdminView = typeof adminModules[number]['id']
