import { HttpClient } from '@gofreego/tsutils'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

export const httpClient = new HttpClient({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: { 'x-user-id': '1', 'x-user-perms': 'projects:read, projects:write, projects:delete, dashboards:read, dashboards:write, dashboards:delete, members:write, analytics:read, events:read, replay:read, replay:delete, persons:read, persons:delete, flags:read, flags:write, flags:delete' },

})

export const loginClient = new HttpClient({
  baseURL: "https://api.bappaapp.com",
  timeout: 30000,
  headers: { 'x-user-id': '1', 'x-user-perms': 'projects:read, projects:write, projects:delete, dashboards:read, dashboards:write, dashboards:delete, members:write, analytics:read, events:read, replay:read, replay:delete, persons:read, persons:delete, flags:read, flags:write, flags:delete' },
})

export default httpClient
