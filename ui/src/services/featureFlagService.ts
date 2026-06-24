import { httpClient } from '../utils/httpClient'
import {
  FeatureFlagResponse,
  CreateFeatureFlagRequest,
  ListFeatureFlagsResponse,
} from '../apis/proto/openclick/v1/feature_flag'

export const featureFlagService = {
  async list(projectId: string): Promise<ListFeatureFlagsResponse> {
    const response = await httpClient.get<ListFeatureFlagsResponse>(`/openclick/api/v1/projects/${projectId}/feature-flags`)
    return response.data
  },

  async create(projectId: string, data: Partial<CreateFeatureFlagRequest>): Promise<FeatureFlagResponse> {
    const response = await httpClient.post<FeatureFlagResponse>(`/openclick/api/v1/projects/${projectId}/feature-flags`, data)
    return response.data
  },

  async update(projectId: string, flagId: string, data: Partial<CreateFeatureFlagRequest>): Promise<FeatureFlagResponse> {
    const response = await httpClient.patch<FeatureFlagResponse>(`/openclick/api/v1/projects/${projectId}/feature-flags/${flagId}`, data)
    return response.data
  },

  async delete(projectId: string, flagId: string): Promise<void> {
    await httpClient.delete(`/openclick/api/v1/projects/${projectId}/feature-flags/${flagId}`)
  },
}

export type FeatureFlag = FeatureFlagResponse;
