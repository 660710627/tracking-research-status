// Test-only transport controls; normal requests reach the isolated real backend.
import { createRoot } from 'react-dom/client'
import App from '../src/App'
import { client } from '../src/api/generated/client.gen'
import '../src/index.css'
import '@fontsource/sarabun/thai-400.css'
import '@fontsource/sarabun/thai-700.css'

let mode = 'real'
let release: (() => void) | undefined
let sequence = 0
const controls = document.getElementById('test-controls')!
const log = document.createElement('pre')
log.id = 'request-log'
log.style.whiteSpace = 'pre-wrap'
log.style.overflowWrap = 'anywhere'
function record(value: unknown) { log.textContent += JSON.stringify(value) + '\n' }
for (const name of ['real', 'post-delay', 'post-500', 'post-network', 'get-delay', 'get-500', 'get-empty', 'release']) {
  const button = document.createElement('button')
  button.textContent = 'TEST ' + name
  button.onclick = () => { if (name === 'release') release?.(); else mode = name }
  controls.append(button)
}
const label = document.createElement('label')
label.htmlFor = 'payload-overrides'
label.textContent = 'TEST multipart overrides (JSON; empty for normal requests)'
const overrides = document.createElement('textarea')
overrides.id = 'payload-overrides'
controls.append(label, overrides, log)
const realFetch = window.fetch.bind(window)
client.setConfig({ fetch: async (request) => {
  let req = request as Request
  const id = ++sequence
  const post = req.method === 'POST'
  if (post) {
    const data = await req.clone().formData()
    if (overrides.value.trim()) {
      const changes = JSON.parse(overrides.value) as Record<string, unknown>
      for (const [key, value] of Object.entries(changes)) {
        if (value === null) data.delete(key)
        else if (key === 'projectMembers') data.set(key, new Blob([JSON.stringify(value)], {type: 'application/json'}), 'members.json')
        else data.set(key, String(value))
      }
      const headers = new Headers(req.headers)
      headers.delete('content-type')
      req = new Request(req.url, {method: 'POST', headers, body: data})
    }
    record({id, method:req.method, parts:Array.from(data.entries()).map(([key,value]) => [key, typeof value === 'string' ? value : {name:value.name,size:value.size,type:value.type}])})
  } else record({id,method:req.method})
  if (mode === (post ? 'post-delay' : 'get-delay')) await new Promise<void>((resolve) => { release = resolve })
  if (post && mode === 'post-network') { record({id,networkError:true}); throw new TypeError('Test network failure') }
  const response = mode === (post ? 'post-500' : 'get-500')
    ? new Response(JSON.stringify({error:{code:'INTERNAL_ERROR',message:'Unable to complete request'}}), {status:500,headers:{'Content-Type':'application/json'}})
    : !post && mode === 'get-empty' ? new Response('[]', {headers:{'Content-Type':'application/json'}})
      : await realFetch(req)
  record({id,status:response.status,body:await response.clone().json()})
  return response
} })
createRoot(document.getElementById('root')!).render(<App />)
