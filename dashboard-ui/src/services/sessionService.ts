import { httpClient } from '../utils/httpClient'
import {
  SessionResponse,
  ListSessionsResponse,
  GetSessionChunksResponse,
} from '../apis/proto/openclick/v1/analytics'

export const sessionService = {
  async list(
    projectId: string,
    params?: {
      dateFrom?: string
      dateTo?: string
      distinctId?: string
      minDurationMs?: number
      search?: string
      limit?: number
      offset?: number
    }
  ): Promise<ListSessionsResponse> {
    const query = new URLSearchParams()
    if (params?.dateFrom) query.set('date_from', params.dateFrom)
    if (params?.dateTo) query.set('date_to', params.dateTo)
    if (params?.distinctId) query.set('distinct_id', params.distinctId)
    if (params?.minDurationMs !== undefined) query.set('min_duration_ms', String(params.minDurationMs))
    if (params?.search) query.set('search', params.search)
    if (params?.limit !== undefined) query.set('limit', String(params.limit))
    if (params?.offset !== undefined) query.set('offset', String(params.offset))
    const qs = query.toString()
    const response = await httpClient.get<ListSessionsResponse>(
      `/api/v1/projects/${projectId}/sessions${qs ? `?${qs}` : ''}`
    )
    return response.data
  },

  async get(projectId: string, sessionId: string): Promise<SessionResponse> {
    const response = await httpClient.get<SessionResponse>(
      `/api/v1/projects/${projectId}/sessions/${sessionId}`
    )
    return response.data
  },

  async getChunks(projectId: string, sessionId: string, fromChunk?: number): Promise<GetSessionChunksResponse> {
    const query = new URLSearchParams()
    if (fromChunk !== undefined) query.set('from_chunk', String(fromChunk))
    const qs = query.toString()
    const response = await httpClient.get<GetSessionChunksResponse>(
      `/api/v1/projects/${projectId}/sessions/${sessionId}/chunks${qs ? `?${qs}` : ''}`
    )
    return response.data
  },

  async delete(projectId: string, sessionId: string): Promise<void> {
    await httpClient.delete(`/api/v1/projects/${projectId}/sessions/${sessionId}`)
  },
}

export type Session = SessionResponse;
