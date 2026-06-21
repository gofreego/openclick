import { httpClient } from '../utils/httpClient'
import {
  QueryEventsRequest,
  QueryEventsResponse,
  EventResult
} from '../apis/proto/openclick/v1/analytics'

export const eventService = {
  async queryEvents(projectId: string, data: Partial<QueryEventsRequest>): Promise<QueryEventsResponse> {
    const response = await httpClient.post<QueryEventsResponse>(`/api/v1/projects/${projectId}/query/events`, {
      ...data,
      limit: data.limit || 50,
      offset: data.offset || 0,
    })
    return response.data
  },
}

export type Event = EventResult;
