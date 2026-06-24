import { httpClient } from '../utils/httpClient'
import {
  ProjectResponse,
  CreateProjectRequest,
  ListProjectsResponse,
  GetProjectResponse,
  AddMemberResponse,
} from '../apis/proto/openclick/v1/project'

const BASE_URL = '/openclick/api/v1/projects'

export const projectService = {
  async list(): Promise<ListProjectsResponse> {
    const response = await httpClient.get<ListProjectsResponse>(BASE_URL)
    return response.data
  },

  async getById(id: string): Promise<GetProjectResponse> {
    const response = await httpClient.get<GetProjectResponse>(`${BASE_URL}/${id}`)
    return response.data
  },

  async create(data: CreateProjectRequest): Promise<ProjectResponse> {
    const response = await httpClient.post<ProjectResponse>(BASE_URL, data)
    return response.data
  },

  async update(id: string, data: { name?: string; timezone?: string }): Promise<ProjectResponse> {
    const response = await httpClient.patch<ProjectResponse>(`${BASE_URL}/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await httpClient.delete(`${BASE_URL}/${id}`)
  },

  async addMember(projectId: string, userId: string, role: string): Promise<AddMemberResponse> {
    const response = await httpClient.post<AddMemberResponse>(`${BASE_URL}/${projectId}/members`, { userId, role })
    return response.data
  },

  async removeMember(projectId: string, userId: string): Promise<void> {
    await httpClient.delete(`${BASE_URL}/${projectId}/members/${userId}`)
  },
}

export type Project = ProjectResponse;
export type ProjectDetail = GetProjectResponse;
