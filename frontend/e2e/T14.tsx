// Test-only entry, excluded from the production index.html bundle.
import { createRoot } from 'react-dom/client'
import App from '../src/App'
import { client } from '../src/api/generated/client.gen'
import fixture from './fixtures/T13_LIST.json'
import type { ResearchChange } from '../src/pages/ResearchListPage'
import '../src/index.css'
import '@fontsource/sarabun/thai-400.css'
import '@fontsource/sarabun/thai-700.css'

let scenario = 'success'
let revision = 0
let release: (() => void) | undefined
const root = createRoot(document.getElementById('root')!)
const controls = document.getElementById('test-controls')!
const log = document.createElement('pre')
log.style.whiteSpace = 'pre-wrap'
log.style.overflowWrap = 'anywhere'
const requests: string[] = []
client.setConfig({ fetch: async (request) => {
  const req = request as Request
  requests.push(req.method + ' ' + new URL(req.url).pathname + ' query=' + new URL(req.url).search + ' body=' + String(req.body))
  log.textContent = requests.join('\n')
  if (scenario === 'network') throw new TypeError('Test network failure')
  if (scenario === 'error') return new Response(JSON.stringify(fixture.error), {status:500,headers:{'Content-Type':'application/json'}})
  if (scenario === 'loading') await new Promise<void>((resolve) => { release = resolve })
  const data = scenario === 'empty' ? fixture.empty : scenario === 'created' ? fixture.afterCreate
    : scenario === 'updated' ? fixture.afterUpdate : scenario === 'long'
      ? Array.from({length:30}, (_, i) => ({...fixture.success[0],id:1000+i,contractNumber:'LONG-'+i+'X'.repeat(100),title:'โครงการทดสอบข้อความยาว'.repeat(30)}))
      : fixture.success
  return new Response(JSON.stringify(data), {status:200,headers:{'Content-Type':'application/json'}})
} })
for (const name of ['success','loading','release','error','network','empty','created','updated','long','retry-success','audit']) {
  const button = document.createElement('button')
  button.textContent = name
  button.onclick = () => {
    if (name === 'audit') {
      const main = document.querySelector('main')!
      const rows = Array.from(main.querySelectorAll('tbody tr'))
      log.textContent += '\nAUDIT ' + JSON.stringify({
        rows:rows.length,
        contracts: rows.slice(0,4).map(r=>r.querySelector('th')?.textContent),
        changed:main.querySelectorAll('.changed').length,
        pageOverflow:document.documentElement.scrollWidth > window.innerWidth,
        focus:document.activeElement?.textContent,
        motion:Array.from(main.querySelectorAll('*')).some(e=>getComputedStyle(e).animationName!=='none'||getComputedStyle(e).transitionDuration!=='0s'),
        text:getComputedStyle(main).color,
        background:getComputedStyle(document.documentElement).backgroundColor
      })
      return
    }
    if (name === 'release') { release?.(); return }
    if (name === 'retry-success') { scenario = 'success'; return }
    scenario = name
    revision++
    const change: ResearchChange | undefined = name === 'created' ? {id:125,kind:'created',revision}
      : name === 'updated' ? {id:108,kind:'updated',revision} : undefined
    root.render(<App key={revision} change={change} />)
  }
  controls.append(button)
}
controls.append(log)
root.render(<App />)
