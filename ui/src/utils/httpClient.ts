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
  onUnauthorized: onUnauthorized,
  // comment this headers these are only for testing purpose, in production we will get the user permissions from the backend
  headers:{"x-user-perms": "oc.projects.read,oc.projects.write,oc.projects.delete,oc.dashboards.read,oc.dashboards.write,oc.dashboards.delete,oc.members.write,oc.analytics.read,oc.events.read,oc.replay.read,oc.replay.delete,oc.persons.read,oc.persons.delete,oc.flags.read,oc.flags.write,oc.flags.delete", "x-user-id":"1"}
})