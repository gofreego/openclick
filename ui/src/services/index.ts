import { httpClient } from '../utils/httpClient'
import { AuthService } from '@gofreego/tsutils'

export const authService = AuthService.getInstance(httpClient)
