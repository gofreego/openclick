import { httpClient } from '../utils/httpClient'

export interface Project {
  id: string
  name: string
  api_key: string
  secret_key?: string
  timezone: string
  created_at: string
}

export interface CreateProjectRequest {
  name: string
  timezone?: string
}

const BASE_URL = '/v1/projects'

export const projectService = {
  async list(): Promise<{ results: Project[] }> {
    const response = await httpClient.get<{ results: Project[] }>(BASE_URL)
    return response.data
  },

  async getById(id: string): Promise<Project> {
    const response = await httpClient.get<Project>(`${BASE_URL}/${id}`)
    return response.data
  },

  async create(data: CreateProjectRequest): Promise<Project> {
    const response = await httpClient.post<Project>(BASE_URL, data)
    return response.data
  },

  async update(id: string, data: Partial<CreateProjectRequest>): Promise<Project> {
    const response = await httpClient.patch<Project>(`${BASE_URL}/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await httpClient.delete(`${BASE_URL}/${id}`)
  },
}
