import { createElement as h } from 'react'
import type { MouseEvent } from 'react'

export default function ResearchBreadcrumb({ current, onBack, disabled = false }: {
  current: 'เพิ่มงานวิจัย' | 'รายละเอียดโครงการ'
  onBack: () => void
  disabled?: boolean
}) {
  const parent = disabled
    ? h('span', { 'aria-disabled': true, title: 'บันทึกหรือยกเลิกการแก้ไขก่อนกลับรายการงานวิจัย' }, 'รายการงานวิจัย')
    : h('a', { href: '#main', onClick: (event: MouseEvent<HTMLAnchorElement>) => {
      event.preventDefault()
      onBack()
    } }, 'รายการงานวิจัย')
  return h('nav', { className: 'research-breadcrumb', 'aria-label': 'เส้นทางนำทาง' },
    h('ol', null,
      h('li', null, parent),
      h('li', null, h('span', { 'aria-current': 'page' }, current)),
    ),
  )
}
