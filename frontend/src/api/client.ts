import { client } from './generated/client.gen'

client.setConfig({
  baseUrl: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080',
})

export { client as apiClient }
export {
  createResearch,
  deleteResearch,
  getHealth,
  listResearches,
  updateResearch,
} from './generated/sdk.gen'
export type {
  ErrorResponse,
  HealthResponse,
  Research,
  ResearchInput,
} from './generated/types.gen'
