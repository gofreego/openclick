import { httpClient } from '../utils/httpClient'
import type { DeviceResponse, ListDevicesResponse, GetDeviceResponse, GetDeviceStatsResponse } from '../apis/proto/openclick/v1/person'

export const deviceService = {
  async list(projectId: string, params?: { deviceId?: string; limit?: number; offset?: number }): Promise<ListDevicesResponse> {
    const query = new URLSearchParams()
    if (params?.deviceId) query.set('device_id', params.deviceId)
    if (params?.limit !== undefined) query.set('limit', String(params.limit))
    if (params?.offset !== undefined) query.set('offset', String(params.offset))
    const qs = query.toString()
    const response = await httpClient.get<ListDevicesResponse>(`/openclick/api/v1/projects/${projectId}/devices${qs ? `?${qs}` : ''}`)
    return response.data
  },

  async get(projectId: string, deviceId: string): Promise<GetDeviceResponse> {
    const response = await httpClient.get<GetDeviceResponse>(`/openclick/api/v1/projects/${projectId}/devices/${encodeURIComponent(deviceId)}`)
    return response.data
  },

  async getStats(projectId: string): Promise<GetDeviceStatsResponse> {
    const response = await httpClient.get<GetDeviceStatsResponse>(`/openclick/api/v1/projects/${projectId}/device-stats`)
    return response.data
  },
}

export type Device = DeviceResponse
