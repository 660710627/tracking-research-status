import type { Role } from './research'

export type AcademicTitleDraft = {
  thaiAbbreviation: string
  englishAbbreviation: string
  thaiTitle: string
  englishTitle: string
}

export type AcademicTitle = AcademicTitleDraft & { id: string }

export const academicTitlesDemo = [
  { thaiAbbreviation: 'ดร.', englishAbbreviation: 'Dr.', thaiTitle: 'อาจารย์', englishTitle: 'Instructor' },
  { thaiAbbreviation: 'ผศ.ดร.', englishAbbreviation: 'Asst. Prof. Dr.', thaiTitle: 'ผู้ช่วยศาสตราจารย์', englishTitle: 'Assistant Professor Doctor' },
  { thaiAbbreviation: 'รศ.ดร.', englishAbbreviation: 'Assoc. Prof. Dr.', thaiTitle: 'รองศาสตราจารย์', englishTitle: 'Associate Professor Doctor' },
  { thaiAbbreviation: 'ศ.ดร.', englishAbbreviation: 'Prof. Dr.', thaiTitle: 'ศาสตราจารย์', englishTitle: 'Professor Doctor' },
  { thaiAbbreviation: 'ผศ.', englishAbbreviation: 'Asst. Prof.', thaiTitle: 'ผู้ช่วยศาสตราจารย์', englishTitle: 'Assistant Professor' },
  { thaiAbbreviation: 'รศ.', englishAbbreviation: 'Assoc. Prof.', thaiTitle: 'รองศาสตราจารย์', englishTitle: 'Associate Professor' },
  { thaiAbbreviation: 'ศ.', englishAbbreviation: 'Prof.', thaiTitle: 'ศาสตราจารย์', englishTitle: 'Professor' },
] as const

function requireAdmin(role: Role) {
  if (role !== 'ผู้ดูแลระบบ') throw new Error('ไม่มีสิทธิ์จัดการตำแหน่งทางวิชาการ')
}

function normalizeDraft(draft: AcademicTitleDraft): AcademicTitleDraft {
  const normalized = {
    thaiAbbreviation: draft.thaiAbbreviation.trim(),
    englishAbbreviation: draft.englishAbbreviation.trim(),
    thaiTitle: draft.thaiTitle.trim(),
    englishTitle: draft.englishTitle.trim(),
  }
  if (Object.values(normalized).some(value => !value)) throw new Error('กรุณากรอกข้อมูลตำแหน่งให้ครบทุกช่อง')
  return normalized
}

export function addAcademicTitle(titles: AcademicTitle[], id: string, draft: AcademicTitleDraft, role: Role): AcademicTitle[] {
  requireAdmin(role)
  if (titles.some(title => title.id === id)) throw new Error('รหัสตำแหน่งซ้ำ')
  return [...titles, { id, ...normalizeDraft(draft) }]
}

export function updateAcademicTitle(titles: AcademicTitle[], id: string, draft: AcademicTitleDraft, role: Role): AcademicTitle[] {
  requireAdmin(role)
  if (!titles.some(title => title.id === id)) throw new Error('ไม่พบตำแหน่งทางวิชาการ')
  const normalized = normalizeDraft(draft)
  return titles.map(title => title.id === id ? { ...title, ...normalized } : title)
}

export function deleteAcademicTitle(titles: AcademicTitle[], id: string, role: Role): AcademicTitle[] {
  requireAdmin(role)
  if (!titles.some(title => title.id === id)) throw new Error('ไม่พบตำแหน่งทางวิชาการ')
  return titles.filter(title => title.id !== id)
}
