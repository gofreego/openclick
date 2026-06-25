import { HttpClient, type HttpError } from '@gofreego/tsutils'
import { API_BASE_URL, LOGIN_URL } from './envs'

const onUnauthorized = (err: HttpError) => {
  // log error
  console.log("Unauthorized response received, redirecting to -> ", LOGIN_URL, err)
  window.location.href = LOGIN_URL
}

export const httpClient = new HttpClient({
  baseURL: API_BASE_URL,
  timeout: 30000,
  onUnauthorized,
})

export default httpClient
