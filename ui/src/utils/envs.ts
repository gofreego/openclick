// define all env flags here

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://api.bappaapp.com'
export const LOGIN_URL = import.meta.env.VITE_LOGIN_URL as string || 'https://admin.bappaapp.com/openauth/admin/v2/login'