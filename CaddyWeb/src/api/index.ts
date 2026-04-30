import axios from 'axios'

import { useSettingsStore } from '@/stores/settings'

const api = axios.create({
    baseURL: '/api',
    timeout: 30000,
})

// 请求拦截器 - 添加 baseURL
api.interceptors.request.use(
    (config) => {
        const settingsStore = useSettingsStore()
        if (settingsStore.apiBaseUrl) {
            config.baseURL = settingsStore.apiBaseUrl + '/api'
        }
        return config
    },
    (error) => Promise.reject(error)
)

// 响应拦截器
api.interceptors.response.use(
    (response) => {
        const res = response.data
        if (res.code !== 0) {
            ElMessage.error(res.message || '请求失败')
            return Promise.reject(res)
        }
        return res.data
    },
    (error) => {
        const message = error.response?.data?.message || error.message || '请求失败'
        ElMessage.error(message)
        return Promise.reject(error)
    }
)

// Caddy2 Admin API 封装
export const caddyAPI = {
    getConfig: () => api.get('/config'),
    putConfig: (data: any) => api.put('/config', data),
    getConfigPath: (path: string) => api.get(`/config/${path}`),
    putConfigPath: (path: string, data: any) => api.put(`/config/${path}`, data),
    patchConfigPath: (path: string, data: any) => api.patch('/config', { path, data }),
    deleteConfigPath: (path: string) => api.delete(`/config/${path}`),
    getDomains: () => api.get('/config/caddyfile/hosts'),
    validateConfig: (config: any) => api.post('/config.validate', config),
    getCertificates: () => api.get('/pki/certificates'),
    loadCertificate: (data: any) => api.post('/pki/load', data),
}

// 域名 API
export const domainAPI = {
    list: () => api.get('/domains'),
    get: (name: string) => api.get(`/domains/${name}`),
    create: (data: any) => api.post('/domains', data),
    update: (name: string, data: any) => api.put(`/domains/${name}`, data),
    delete: (server_id: string, domain_id: string) => api.delete(`/domains/${server_id}/domains/${domain_id}`),
    getSites: (server_id: string) => api.get(`/domain-sites/${server_id}`),
}

// 站点 API (子站点 - 属于域名)
export const siteAPI = {
    list: () => api.get('/sites'),
    get: (domain: string, name: string) => api.get(`/domains/${domain}/sites/${name}`),
    create: (domain: string, data: any) => api.post(`/domains/${domain}/sites`, data),
    update: (siteId: string, data: any) => api.put(`/sites/${siteId}`, data),
    delete: (domain: string, site_id: string) => api.delete(`/sites/${domain}/${site_id}`),
}

// 应用设置 API
export const settingsAPI = {
    get: () => api.get('/settings'),
    save: (data: any) => api.post('/settings', data),
    detectCaddy: () => api.get('/caddy/detect'),
    installCaddy: (installType: string) => api.post('/caddy/install', { installType }),
    getCaddyStatus: () => api.get('/caddy/status'),
    checkCaddyInstallStatus: () => api.get('/caddy/check-install'),
    initCaddy: (force: boolean = false) => api.post('/caddy/init', { force }),
    getServers: () => api.get('/caddy/servers'),
}

// 日志 API
export const logsAPI = {
    get: (params?: { limit?: number; offset?: number }) => api.get('/logs', { params }),
    stream: () => {
        const settingsStore = useSettingsStore()
        const wsUrl = settingsStore.wsBaseUrl
            ? settingsStore.wsBaseUrl.replace('http', 'ws') + '/api/logs/stream'
            : `ws://${window.location.hostname}:8081/api/logs/stream`
        return new WebSocket(wsUrl)
    },
}

// SSE API
export const sseAPI = {
    createConnection: () => {
        const wsBaseUrl = localStorage.getItem('wsBaseUrl') || `http://${window.location.hostname}:8081`
        const url = wsBaseUrl.includes('localhost')
            ? `http://${window.location.hostname}:8081/api/sse`
            : wsBaseUrl + '/api/sse'
        return new EventSource(url)
    },
}

export default api
