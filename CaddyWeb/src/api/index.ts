import axios, { AxiosInstance } from 'axios'
import { useSettingsStore } from '@/stores/settings'

interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

interface CaddyConfig {
  apps?: {
    http?: {
      servers?: Record<string, CaddyServer>
    }
    tls?: {
      automation?: {
        policies?: unknown[]
      }
    }
  }
}

interface CaddyServer {
  listen: string[]
  routes?: CaddyRoute[]
  tls_connection_policies?: unknown[]
  errors?: unknown
}

interface CaddyRoute {
  '@id'?: string
  match?: CaddyMatch[]
  handle?: CaddyHandle[]
  terminal?: boolean
}

interface CaddyMatch {
  host?: string[]
  path?: string[]
}

interface CaddyHandle {
  handler: string
  '@id'?: string
  root?: string
  index_names?: string[]
  upstreams?: CaddyUpstream[]
  routes?: CaddyRoute[]
  health_checks?: CaddyHealthChecks
}

interface CaddyUpstream {
  dial: string
}

interface CaddyHealthChecks {
  active?: {
    path?: string
    interval?: string
    timeout?: string
  }
}

interface Domain {
  name: string
  wildcard: string
  server_id: string
  listen: string[]
  tls_enabled: boolean
  id: string
}

interface Site {
  name: string
  host: string
  type: 'static' | 'reverse_proxy'
  upstream?: string
  root?: string
  index_names?: string
  health_check: boolean
  id: string
  server_id: string
}

interface CreateSiteRequest {
  name: string
  type: 'static' | 'reverse_proxy'
  domain: string
  upstream?: string
  root?: string
  index_names?: string
  health_check?: boolean
}

interface UpdateSiteRequest {
  name?: string
  type: 'static' | 'reverse_proxy'
  domain?: string
  upstream?: string
  root?: string
  index_names?: string
  health_check?: boolean
}

interface AppSettings {
  deployed_mode: 'local' | 'remote'
  api_base_url: string
  ws_base_url: string
  caddy: CaddySettings
  theme: 'light' | 'dark' | 'auto'
  language: 'zh-CN' | 'en-US'
  first_launch: boolean
}

interface CaddySettings {
  api_url: string
  unix_socket: string
  admin_port: number
}

interface CaddyStatus {
  status: 'running' | 'stopped'
  version: string
}

interface CaddyInstallStatus {
  installed: boolean
  version: string
  running: boolean
  unix_socket: string
  admin_port: number
}

interface InitResult {
  success: boolean
  warning?: boolean
  existing_count?: number
  version?: string
  server_id?: string
  ports?: string[]
}

interface LogEntry {
  timestamp: string
  level: string
  message: string
}

interface LogParams {
  limit?: number
  offset?: number
}

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

api.interceptors.request.use(
  (config) => {
    // const settingsStore = useSettingsStore()
    // if (settingsStore.settings.apiBaseUrl) {
    //   config.baseURL = settingsStore.settings.apiBaseUrl + '/api'
    // }
    return config
  },
  (error) => Promise.reject(error)
)

// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(api.interceptors.response.use as any)(
  (response: any) => {
    const res = response.data as ApiResponse
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(res)
    }
    return res.data
  },
  (error: any) => {
    const message =
      error.response?.data?.message || error.message || '请求失败'
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

async function get<T>(url: string, config?: Parameters<typeof api.get>[1]): Promise<T> {
  return api.get<T>(url, config) as Promise<T>
}

async function post<T>(url: string, data?: unknown): Promise<T> {
  return api.post<T>(url, data) as Promise<T>
}

async function put<T>(url: string, data?: unknown): Promise<T> {
  return api.put<T>(url, data) as Promise<T>
}

async function patch<T>(url: string, data?: unknown): Promise<T> {
  return api.patch<T>(url, data) as Promise<T>
}

async function del<T>(url: string): Promise<T> {
  return api.delete<T>(url) as Promise<T>
}

export const caddyAPI = {
  getConfig: () => get<CaddyConfig>('/config'),

  putConfig: (data: CaddyConfig) => put<CaddyConfig>('/config', data),

  getConfigPath: <T = unknown>(path: string) => get<T>(`/config/${path}`),

  putConfigPath: <T = unknown>(path: string, data: T) =>
    put<T>(`/config/${path}`, data),

  patchConfigPath: (_path: string, data: { path: string; data: unknown }) =>
    patch('/config', data),

  deleteConfigPath: (path: string) => del(`/config/${path}`),

  getDomains: () => get<Domain[]>('/config/caddyfile/hosts'),

  validateConfig: (config: CaddyConfig) =>
    post<CaddyConfig>('/config.validate', config),

  getCertificates: () => get('/pki/certificates'),

  loadCertificate: (data: unknown) => post('/pki/load', data),
}

export const domainAPI = {
  list: () => get<Domain[]>('/domains'),

  get: (name: string) => get<Domain>(`/domains/${name}`),

  create: (data: { name: string; server_id?: string; tls?: boolean }) =>
    post<{ name: string; server_id: string; id: string }>('/domains', data),

  update: (name: string, data: Partial<Domain>) =>
    put<Domain>(`/domains/${name}`, data),

  delete: (serverId: string, domainId: string) =>
    del(`/domains/${serverId}/domains/${domainId}`),

  getSites: (serverId: string) => get<Site[]>(`/domain-sites/${serverId}`),
}

export const siteAPI = {
  list: () => get<Site[]>('/sites'),

  get: (domain: string, name: string) =>
    get<Site>(`/domains/${domain}/sites/${name}`),

  create: (domain: string, data: CreateSiteRequest) =>
    post<{ domain: string; name: string; id: string }>(
      `/domains/${domain}/sites`,
      data
    ),

  update: (siteId: string, data: UpdateSiteRequest) =>
    put<Site>(`/sites/${siteId}`, data),

  delete: (domain: string, siteId: string) =>
    del(`/sites/${domain}/${siteId}`),
}

export const settingsAPI = {
  get: () => get<AppSettings>('/settings'),

  save: (data: AppSettings) => post('/settings', data),

  detectCaddy: () => get<{ success: boolean; version?: string }>('/caddy/detect'),

  installCaddy: (installType: string) =>
    post<{ success: boolean; version?: string; action?: string }>(
      '/caddy/install',
      { installType }
    ),

  getCaddyStatus: () => get<CaddyStatus>('/caddy/status'),

  checkCaddyInstallStatus: () => get<CaddyInstallStatus>('/caddy/check-install'),

  initCaddy: (force = false) => post<InitResult>('/caddy/init', { force }),

  getServers: () => get<{ id: string; listen: string[] }[]>('/caddy/servers'),
}

export const logsAPI = {
  get: (params?: LogParams) => get<LogEntry[]>('/logs', { params }),

  stream: () => {
    const settingsStore = useSettingsStore()
    const wsUrl = settingsStore.settings.wsBaseUrl
      ? settingsStore.settings.wsBaseUrl.replace('http', 'ws') +
        '/api/logs/stream'
      : `ws://${window.location.hostname}:8081/api/logs/stream`
    return new WebSocket(wsUrl)
  },
}

export const sseAPI = {
  createConnection: () => {
    const wsBaseUrl =
      localStorage.getItem('wsBaseUrl') ||
      `http://${window.location.hostname}:8081`
    const url = wsBaseUrl.includes('localhost')
      ? `http://${window.location.hostname}:8081/api/sse`
      : wsBaseUrl + '/api/sse'
    return new EventSource(url)
  },
}

export type {
  ApiResponse,
  CaddyConfig,
  CaddyServer,
  CaddyRoute,
  CaddyMatch,
  CaddyHandle,
  CaddyUpstream,
  CaddyHealthChecks,
  Domain,
  Site,
  CreateSiteRequest,
  UpdateSiteRequest,
  AppSettings,
  CaddySettings,
  CaddyStatus,
  CaddyInstallStatus,
  InitResult,
  LogEntry,
  LogParams,
}

export default api