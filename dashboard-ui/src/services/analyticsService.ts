import { httpClient } from '../utils/httpClient'

export const analyticsService = {
  async queryTrends(projectId: string, data: any): Promise<{ results: any[] }> {
    const response = await httpClient.post<{ results: any[] }>(`/api/v1/projects/${projectId}/query/trends`, data)
    return response.data
  },

  async queryFunnel(projectId: string, data: any): Promise<{ result: any[] }> {
    const response = await httpClient.post<{ result: any[] }>(`/api/v1/projects/${projectId}/query/funnel`, data)
    return response.data
  },

  async queryRetention(projectId: string, data: any): Promise<{ result: any[] }> {
    const response = await httpClient.post<{ result: any[] }>(`/api/v1/projects/${projectId}/query/retention`, data)
    return response.data
  },
}
