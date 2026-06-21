import { httpClient } from '../utils/httpClient'

export interface FeatureFlag {
  id: string
  key: string
  name: string
  active: boolean
  rollout_pct: number
  filters: any
  created_at: string
}

export interface CreateFeatureFlagRequest {
  key: string
  name: string
  active: boolean
  rollout_pct: number
  filters?: any
}

export const featureFlagService = {
  async list(projectId: string): Promise<{ results: FeatureFlag[] }> {
    const response = await httpClient.get<{ results: FeatureFlag[] }>(`/v1/projects/${projectId}/feature-flags`)
    return response.data
  },

  async create(projectId: string, data: CreateFeatureFlagRequest): Promise<FeatureFlag> {
    const response = await httpClient.post<FeatureFlag>(`/v1/projects/${projectId}/feature-flags`, data)
    return response.data
  },

  async update(projectId: string, flagId: string, data: Partial<CreateFeatureFlagRequest>): Promise<FeatureFlag> {
    const response = await httpClient.patch<FeatureFlag>(`/v1/projects/${projectId}/feature-flags/${flagId}`, data)
    return response.data
  },

  async delete(projectId: string, flagId: string): Promise<void> {
    await httpClient.delete(`/v1/projects/${projectId}/feature-flags/${flagId}`)
  },
}
