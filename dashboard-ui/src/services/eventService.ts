import { httpClient } from '../utils/httpClient'

export interface Event {
  uuid: string
  event: string
  distinct_id: string
  timestamp: string
  properties: any
}

export interface QueryEventsRequest {
  event?: string
  date_from?: string
  date_to?: string
  distinct_id?: string
  limit?: number
  offset?: number
}

export const eventService = {
  async queryEvents(projectId: string, data: QueryEventsRequest): Promise<{ results: Event[], total: number }> {
    const response = await httpClient.post<{ results: Event[], total: number }>(`/v1/projects/${projectId}/query/events`, {
      ...data,
      limit: data.limit || 50,
      offset: data.offset || 0,
    })
    return response.data
  },
}
