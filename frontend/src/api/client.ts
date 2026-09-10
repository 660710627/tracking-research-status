import { client } from './generated/client.gen'
import {
  createResearch as createResearchGenerated,
  deleteResearch as deleteResearchGenerated,
  getHealth as getHealthGenerated,
  listResearches as listResearchesGenerated,
  updateResearch as updateResearchGenerated,
  updateResearchProcess as updateResearchProcessGenerated,
  updateResearchStatus as updateResearchStatusGenerated,
} from './generated/sdk.gen'
import type {
  CreateResearchRequest,
  ResearchId,
  ResearchProcess,
  ResearchStatus,
  UpdateResearchRequest,
} from './generated/types.gen'

const projectMembersPart = (
  members: CreateResearchRequest['projectMembers'],
) =>
  new Blob([JSON.stringify(members)], {
    type: 'application/json',
  })

const appendEditableFields = (
  formData: FormData,
  body: CreateResearchRequest | UpdateResearchRequest,
) => {
  formData.append('title', body.title)
  formData.append('isSubsidized', String(body.isSubsidized))
  formData.append(
    'projectMembers',
    projectMembersPart(body.projectMembers),
    'project-members.json',
  )
  formData.append('fundingType', body.fundingType)
  formData.append('fundingSourceName', body.fundingSourceName)
  formData.append('contractNumber', body.contractNumber)

  if (body.contractFile !== undefined) {
    formData.append('contractFile', body.contractFile)
  }

  formData.append('projectType', body.projectType)
  formData.append('responsibleProjectUnit', body.responsibleProjectUnit)
  formData.append('responsibleBudgetUnit', body.responsibleBudgetUnit)
  formData.append('startDate', body.startDate)
  formData.append('endDate', body.endDate)
  formData.append('budgetAmount', String(body.budgetAmount))
  formData.append('thaiAbstract', body.thaiAbstract)
  formData.append('englishAbstract', body.englishAbstract)
  formData.append('objectives', body.objectives)
  formData.append('keywords', body.keywords)
}

const serializeCreateResearch = (body: CreateResearchRequest) => {
  const formData = new FormData()

  appendEditableFields(formData, body)
  formData.append(
    'continuationOfId',
    body.continuationOfId === null ? 'null' : String(body.continuationOfId),
  )
  formData.append('researchKind', body.researchKind)

  return formData
}

const serializeUpdateResearch = (body: UpdateResearchRequest) => {
  const formData = new FormData()
  appendEditableFields(formData, body)
  return formData
}

export const configureApiClient = (baseUrl: string) => {
  client.setConfig({ baseUrl })
}

export const api = {
  getHealth: () => getHealthGenerated({ client }),
  listResearches: () => listResearchesGenerated({ client }),
  createResearch: (body: CreateResearchRequest) =>
    createResearchGenerated({
      body,
      bodySerializer: serializeCreateResearch,
      client,
    }),
  deleteResearch: (id: ResearchId) =>
    deleteResearchGenerated({ client, path: { id } }),
  updateResearch: (id: ResearchId, body: UpdateResearchRequest) =>
    updateResearchGenerated({
      body,
      bodySerializer: serializeUpdateResearch,
      client,
      path: { id },
    }),
  updateResearchStatus: (id: ResearchId, status: ResearchStatus) =>
    updateResearchStatusGenerated({
      body: { status },
      client,
      path: { id },
    }),
  updateResearchProcess: (id: ResearchId, process: ResearchProcess) =>
    updateResearchProcessGenerated({
      body: { process },
      client,
      path: { id },
    }),
}
