import { httpClient } from '../utils/httpClient'
import {
  ProjectResponse,
  CreateProjectRequest,
  ListProjectsResponse,
} from '../apis/proto/openclick/v1/project'

const BASE_URL = '/api/v1/projects'

export const projectService = {
  async list(): Promise<ListProjectsResponse> {
    const response = await httpClient.get<ListProjectsResponse>(BASE_URL)
    return response.data
  },

  async getById(id: string): Promise<ProjectResponse> {
    const response = await httpClient.get<ProjectResponse>(`${BASE_URL}/${id}`)
    return response.data
  },

  async create(data: CreateProjectRequest): Promise<ProjectResponse> {
    const response = await httpClient.post<ProjectResponse>(BASE_URL, data)
    return response.data
  },

  async update(id: string, data: Partial<CreateProjectRequest>): Promise<ProjectResponse> {
    const response = await httpClient.patch<ProjectResponse>(`${BASE_URL}/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await httpClient.delete(`${BASE_URL}/${id}`)
  },
}

export type Project = ProjectResponse;
