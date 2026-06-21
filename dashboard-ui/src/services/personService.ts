import { httpClient } from '../utils/httpClient'

export interface Person {
  id: string
  distinct_id: string
  properties: any
  created_at: string
}

export const personService = {
  async list(projectId: string): Promise<{ results: Person[], total: number }> {
    const response = await httpClient.get<{ results: Person[], total: number }>(`/v1/projects/${projectId}/persons`)
    return response.data
  },

  async delete(projectId: string, distinctId: string): Promise<void> {
    await httpClient.delete(`/v1/projects/${projectId}/persons/${distinctId}`)
  },
}
