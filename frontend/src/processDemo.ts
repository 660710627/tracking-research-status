export type ProcessDetails={subprocess:string;note:string;file:File|null}
export const demoSubprocesses:Record<number,{name:string;attachment:boolean}[]>={
  1:[{name:'จัดเตรียมสัญญา',attachment:false},{name:'ลงนามสัญญา',attachment:true}],
  2:[{name:'จัดทำบันทึกข้อตกลง',attachment:true},{name:'ตรวจสอบบันทึกข้อตกลง',attachment:false}],
  3:[{name:'ยื่นคำขอเปิดบัญชี',attachment:true},{name:'ยืนยันการเปิดบัญชี',attachment:false}],
  4:[{name:'ยื่นเอกสารเบิกจ่าย',attachment:true},{name:'ตรวจสอบการเบิกจ่าย',attachment:false}],
  5:[{name:'คำนวณค่าธรรมเนียม',attachment:false},{name:'แนบหลักฐานจัดสรรค่าธรรมเนียม',attachment:true}],
  6:[{name:'ติดตามรายงานความก้าวหน้า',attachment:false},{name:'รับรายงานความก้าวหน้า',attachment:true}],
  7:[{name:'จัดทำรายงานสรุปการใช้เงิน',attachment:true},{name:'ตรวจสอบรายงานการใช้เงิน',attachment:false}],
  8:[{name:'ยื่นคำขอปิดบัญชี',attachment:true},{name:'ยืนยันการปิดบัญชี',attachment:false}],
}
export function validateProcessDetails(process:number,details:ProcessDetails){
  const option=demoSubprocesses[process]?.find(item=>item.name===details.subprocess)
  if(!option)throw new Error('เลือกกระบวนการย่อยให้ตรงกับกระบวนการหลัก')
  if(details.file&&!option.attachment)throw new Error('กระบวนการย่อยนี้ไม่รองรับไฟล์แนบ')
  if(details.file&&details.file.size===0)throw new Error('ไม่สามารถแนบไฟล์ว่างได้')
}
