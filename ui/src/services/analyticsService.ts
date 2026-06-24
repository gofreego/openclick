import { httpClient } from '../utils/httpClient'
import {
  QueryTrendsRequest,
  QueryTrendsResponse,
  QueryFunnelRequest,
  QueryFunnelResponse,
  QueryRetentionRequest,
  QueryRetentionResponse,
  QueryPathsRequest,
  QueryPathsResponse,
  QueryEventsRequest,
  QueryEventsResponse,
  ListEventNamesResponse,
} from '../apis/proto/openclick/v1/analytics'

export const analyticsService = {
  async queryTrends(projectId: string, data: Partial<QueryTrendsRequest>): Promise<QueryTrendsResponse> {
    const response = await httpClient.post<QueryTrendsResponse>(`/openclick/api/v1/projects/${projectId}/query/trends`, data)
    return response.data
  },

  async queryFunnel(projectId: string, data: Partial<QueryFunnelRequest>): Promise<QueryFunnelResponse> {
    const response = await httpClient.post<QueryFunnelResponse>(`/openclick/api/v1/projects/${projectId}/query/funnel`, data)
    return response.data
  },

  async queryRetention(projectId: string, data: Partial<QueryRetentionRequest>): Promise<QueryRetentionResponse> {
    const response = await httpClient.post<QueryRetentionResponse>(`/openclick/api/v1/projects/${projectId}/query/retention`, data)
    return response.data
  },

  async queryPaths(projectId: string, data: Partial<QueryPathsRequest>): Promise<QueryPathsResponse> {
    const response = await httpClient.post<QueryPathsResponse>(`/openclick/api/v1/projects/${projectId}/query/paths`, data)
    return response.data
  },

  async queryEvents(projectId: string, data: Partial<QueryEventsRequest>): Promise<QueryEventsResponse> {
    const response = await httpClient.post<QueryEventsResponse>(`/openclick/api/v1/projects/${projectId}/query/events`, {
      ...data,
      limit: data.limit || 50,
      offset: data.offset || 0,
    })
    return response.data
  },

  async listEventNames(projectId: string): Promise<string[]> {
    const response = await httpClient.get<ListEventNamesResponse>(`/openclick/api/v1/projects/${projectId}/event-names`)
    return response.data.eventNames || []
  },
}
