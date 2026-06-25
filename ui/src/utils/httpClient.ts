import { HttpClient } from '@gofreego/tsutils'
import { SessionManager } from '@gofreego/tsutils'
import {  type HttpError } from '@gofreego/tsutils'
import {  LOGIN_URL, API_BASE_URL } from '../utils/envs'

export const sessionManager = SessionManager.getInstance()

const onUnauthorized = (err: HttpError) => {
  // log error
  console.log("Unauthorized response received, redirecting to -> ", LOGIN_URL, err)
  sessionManager.clear()
  window.location.href = LOGIN_URL
}

export const httpClient = new HttpClient({
  baseURL: API_BASE_URL,
  timeout: 30000,
  onUnauthorized: onUnauthorized
})