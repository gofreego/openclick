import { httpClient } from '../utils/httpClient'
import { PersonResponse, ListPersonsResponse, GetPersonResponse } from '../apis/proto/openclick/v1/person'

export const personService = {
  async list(projectId: string, params?: { search?: string; limit?: number; offset?: number }): Promise<ListPersonsResponse> {
    const query = new URLSearchParams()
    if (params?.search) query.set('search', params.search)
    if (params?.limit !== undefined) query.set('limit', String(params.limit))
    if (params?.offset !== undefined) query.set('offset', String(params.offset))
    const qs = query.toString()
    const response = await httpClient.get<ListPersonsResponse>(`/openclick/api/v1/projects/${projectId}/persons${qs ? `?${qs}` : ''}`)
    return response.data
  },

  async get(projectId: string, distinctId: string): Promise<GetPersonResponse> {
    const response = await httpClient.get<GetPersonResponse>(`/openclick/api/v1/projects/${projectId}/persons/${encodeURIComponent(distinctId)}`)
    return response.data
  },

  async delete(projectId: string, distinctId: string): Promise<void> {
    await httpClient.delete(`/openclick/api/v1/projects/${projectId}/persons/${encodeURIComponent(distinctId)}`)
  },
}

export type Person = PersonResponse;
export type PersonDetail = GetPersonResponse;
