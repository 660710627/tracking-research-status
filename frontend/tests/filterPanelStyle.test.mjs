import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

test('filter panel retains native disclosure with a prominent, keyboard-focusable heading', () => {
  const source = readFileSync(new URL('../src/App.tsx', import.meta.url), 'utf8')
  const css = readFileSync(new URL('../src/styles.css', import.meta.url), 'utf8')
  assert.match(source, /<details className="research-filter-panel" open>/)
  assert.match(source, /className="research-filter-icon"><IconSearch/)
  assert.match(css, /\.research-filter-panel>summary\{[^}]*background:#006b60;color:#fff/)
  assert.match(css, /\.research-filter-panel>summary:focus-visible/)
  assert.match(css, /\[open\]>summary \.research-filter-toggle svg\{transform:rotate\(180deg\)/)
  assert.ok(!css.includes('[open]>summary svg{'))
})
