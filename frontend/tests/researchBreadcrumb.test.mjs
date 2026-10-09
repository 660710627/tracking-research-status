import test from 'node:test'
import assert from 'node:assert/strict'
import { renderToStaticMarkup } from 'react-dom/server'
import ResearchBreadcrumb from '../src/ResearchBreadcrumb.ts'
import { readFileSync } from 'node:fs'

test('current breadcrumb page is underlined', () => {
  const css = readFileSync(new URL('../src/styles.css', import.meta.url), 'utf8')
  const currentStyle = css.match(/\.research-breadcrumb \[aria-current=page\]\{([^}]+)\}/)?.[1]
  assert.ok(currentStyle?.includes('text-decoration:underline'))
  assert.ok(currentStyle?.includes('text-underline-offset:3px'))
})

test('both research subpages show a list link and a non-link current page', () => {
  for (const current of ['เพิ่มงานวิจัย', 'รายละเอียดโครงการ']) {
    let navigated = false
    const tree = ResearchBreadcrumb({ current, onBack: () => { navigated = true } })
    const html = renderToStaticMarkup(tree)
    assert.match(html, /<nav[^>]+aria-label="เส้นทางนำทาง"/)
    assert.match(html, /<a href="#main">รายการงานวิจัย<\/a>/)
    assert.ok(html.includes(`<span aria-current="page">${current}</span>`))
    const link = tree.props.children.props.children[0].props.children
    let prevented = false
    link.props.onClick({ preventDefault() { prevented = true } })
    assert.ok(prevented && navigated)
  }
})

test('detail editing preserves the existing disabled back navigation', () => {
  const html = renderToStaticMarkup(ResearchBreadcrumb({
    current: 'รายละเอียดโครงการ', disabled: true,
    onBack: () => assert.fail('must not navigate'),
  }))
  assert.match(html, /aria-disabled="true"/)
  assert.ok(!html.includes('<a '))
  assert.match(html, /aria-current="page"/)
})
