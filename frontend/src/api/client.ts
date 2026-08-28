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
  ContractFile,
  ContractFileMetadata,
  BuddhistDate,
  BudgetAmount,
  ErrorCode,
  ErrorResponse,
  FundingType,
  HealthResponse,
  NullableResearchId,
  ProjectType,
  Research,
  ResearchId,
  ResearchKind,
  ResearchMember,
  ResearchProcess,
  ResearchStatus,
  ResearchText,
  ResearchTitle,
  UpdateProcessRequest,
  UpdateResearchRequest,
  UpdateStatusRequest,
} from './generated/types.gen'

import { createResearch } from './generated/sdk.gen'
import type { CreateResearchRequest, ResearchMember } from './generated/types.gen'

/**
 * The generated multipart serializer serializes arrays as repeated parts and
 * drops null values. This contract requires one JSON `projectMembers` part and
 * an explicit `continuationOfId=null` part, so adapt only those wire values
 * while retaining the generated operation and its typed request shape.
 */
export function createResearchMultipart(body: CreateResearchRequest) {
  return createResearch({
    body: {
      ...body,
      continuationOfId: (body.continuationOfId === null ? 'null' : String(body.continuationOfId)) as unknown as number | null,
      projectMembers: JSON.stringify(body.projectMembers) as unknown as ResearchMember[],
    },
  })
}
