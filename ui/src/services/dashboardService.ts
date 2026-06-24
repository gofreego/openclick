import { httpClient } from '../utils/httpClient'
import {
  DashboardResponse,
  ListDashboardsResponse,
  GetDashboardResponse,
  DashboardItemResponse,
} from '../apis/proto/openclick/v1/dashboard'

export const dashboardService = {
  async list(projectId: string): Promise<ListDashboardsResponse> {
    const response = await httpClient.get<ListDashboardsResponse>(`/openclick/api/v1/projects/${projectId}/dashboards`)
    return response.data
  },

  async create(projectId: string, name: string): Promise<DashboardResponse> {
    const response = await httpClient.post<DashboardResponse>(`/openclick/api/v1/projects/${projectId}/dashboards`, { name })
    return response.data
  },

  async get(projectId: string, dashboardId: string): Promise<GetDashboardResponse> {
    const response = await httpClient.get<GetDashboardResponse>(`/openclick/api/v1/projects/${projectId}/dashboards/${dashboardId}`)
    return response.data
  },

  async delete(projectId: string, dashboardId: string): Promise<void> {
    await httpClient.delete(`/openclick/api/v1/projects/${projectId}/dashboards/${dashboardId}`)
  },

  async createItem(projectId: string, dashboardId: string, data: {
    name: string
    type: string
    query?: Record<string, any>
    position?: Record<string, any>
  }): Promise<DashboardItemResponse> {
    const response = await httpClient.post<DashboardItemResponse>(
      `/openclick/api/v1/projects/${projectId}/dashboards/${dashboardId}/items`,
      data
    )
    return response.data
  },

  async updateItem(projectId: string, dashboardId: string, itemId: string, data: {
    name?: string
    type?: string
    query?: Record<string, any>
    position?: Record<string, any>
  }): Promise<DashboardItemResponse> {
    const response = await httpClient.patch<DashboardItemResponse>(
      `/openclick/api/v1/projects/${projectId}/dashboards/${dashboardId}/items/${itemId}`,
      data
    )
    return response.data
  },

  async deleteItem(projectId: string, dashboardId: string, itemId: string): Promise<void> {
    await httpClient.delete(`/openclick/api/v1/projects/${projectId}/dashboards/${dashboardId}/items/${itemId}`)
  },
}

export type Dashboard = DashboardResponse;
export type DashboardDetail = GetDashboardResponse;
export type DashboardItem = DashboardItemResponse;
