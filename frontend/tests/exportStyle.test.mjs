import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const css = readFileSync(new URL('../src/styles.css', import.meta.url), 'utf8')

test('selected export state uses blue while the unavailable state stays gray', () => {
  const ready = css.match(/\.research-export\.ready:disabled\{([^}]+)\}/)?.[1]
  const unavailable = css.match(/\.research-export:disabled\{([^}]+)\}/)?.[1]
  assert.ok(ready?.includes('background:#2563eb'))
  assert.ok(ready?.includes('color:#fff'))
  assert.ok(unavailable?.includes('background:#f1f3f5'))
})
