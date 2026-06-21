import { httpClient } from '../utils/httpClient'
import { PersonResponse, ListPersonsResponse } from '../apis/proto/openclick/v1/person'

export const personService = {
  async list(projectId: string): Promise<ListPersonsResponse> {
    const response = await httpClient.get<ListPersonsResponse>(`/api/v1/projects/${projectId}/persons`)
    return response.data
  },

  async delete(projectId: string, distinctId: string): Promise<void> {
    await httpClient.delete(`/api/v1/projects/${projectId}/persons/${distinctId}`)
  },
}

export type Person = PersonResponse;
