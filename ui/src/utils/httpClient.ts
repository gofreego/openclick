import { HttpClient } from '@gofreego/tsutils'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://api.bappaapp.com'

export const httpClient = new HttpClient({
  baseURL: API_BASE_URL,
  timeout: 30000,
})

export default httpClient
