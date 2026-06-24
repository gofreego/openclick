import { httpClient } from '../utils/httpClient'
import { CohortResponse, ListCohortsResponse } from '../apis/proto/openclick/v1/person'

export const cohortService = {
  async list(projectId: string): Promise<ListCohortsResponse> {
    const response = await httpClient.get<ListCohortsResponse>(`/openclick/api/v1/projects/${projectId}/cohorts`)
    return response.data
  },

  async create(projectId: string, data: { name: string; filters: Record<string, any> }): Promise<CohortResponse> {
    const response = await httpClient.post<CohortResponse>(`/openclick/api/v1/projects/${projectId}/cohorts`, data)
    return response.data
  },

  async delete(projectId: string, cohortId: string): Promise<void> {
    await httpClient.delete(`/openclick/api/v1/projects/${projectId}/cohorts/${cohortId}`)
  },
}

export type Cohort = CohortResponse;
