import httpClient from '../utils/httpClient'
import { AuthService, SessionManager } from '@gofreego/tsutils'

export const sessionManager = SessionManager.getInstance(httpClient)
export const authService = AuthService.getInstance(httpClient)
