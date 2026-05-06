import { get, post, put, patch, del } from './axios'
import { useSettingsStore } from '@/stores/settings'
import type {
  CaddyConfig,
  Domain,
  Site,
  CreateSiteRequest,
  UpdateSiteRequest,
  AppSettings,
  CaddySettings,
  CaddyStatusResponse,
  CaddyInstallStatus,
  InitResult,
  LogEntry,
  LogParams,
  Certificate,
} from '@/types'

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

  getCertificates: () => get<Certificate[]>('/certs'),

  addCertificate: (data: unknown) => post('/certs', data),
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

  getCaddyStatus: () => get<CaddyStatusResponse>('/caddy/status'),

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
  CaddyConfig,
  Domain,
  Site,
  CreateSiteRequest,
  UpdateSiteRequest,
  AppSettings,
  CaddySettings,
  CaddyStatusResponse,
  CaddyInstallStatus,
  InitResult,
  LogEntry,
  LogParams,
}
