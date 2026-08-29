import { useEffect, useState } from 'react'
import './App.css'
import { CreateResearchPage } from './pages/CreateResearchPage'
import { ResearchRegisterPage } from './pages/ResearchRegisterPage'

type Page = 'register' | 'create'

function pageFromHash(): Page {
  return window.location.hash === '#/researches/new' ? 'create' : 'register'
}

function createdResearchIdFromHash(): number | undefined {
  const match = /^#\/researches\?created=(\d+)$/.exec(window.location.hash)
  return match ? Number(match[1]) : undefined
}

function App() {
  const [page, setPage] = useState<Page>(pageFromHash)

  useEffect(() => {
    const syncPage = () => setPage(pageFromHash())
    window.addEventListener('hashchange', syncPage)
    return () => window.removeEventListener('hashchange', syncPage)
  }, [])

  const returnToRegister = (createdResearchId?: number) => {
    window.location.hash = createdResearchId ? `/researches?created=${createdResearchId}` : ''
  }

  if (page === 'create') return <CreateResearchPage onReturnToRegister={returnToRegister} />

  return <ResearchRegisterPage createdResearchId={createdResearchIdFromHash()} onCreate={() => { window.location.hash = '/researches/new' }} />
}

export default App
