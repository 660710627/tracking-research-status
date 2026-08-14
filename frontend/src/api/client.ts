import { client } from './generated/client.gen'

client.setConfig({
  baseUrl: import.meta.env.VITE_API_BASE_URL ?? window.location.origin,
})

export { client as apiClient }
export {
  createResearch,
  deleteResearch,
  getHealth,
  listResearches,
  updateResearch,
  updateResearchProcess,
  updateResearchStatus,
} from './generated/sdk.gen'
export type {
  CreateResearchRequest,
  ErrorCode,
  ErrorResponse,
  HealthResponse,
  NullableResearchId,
  Research,
  ResearchDescription,
  ResearchId,
  ResearchProcess,
  ResearchStatus,
  ResearchTitle,
  UpdateProcessRequest,
  UpdateResearchRequest,
  UpdateStatusRequest,
} from './generated/types.gen'
