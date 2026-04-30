export interface ServerInfo {
    id: string
    listen: string[]
}

export interface DomainFormData {
    name: string
    server_id: string
    tls_enabled: boolean
    id: string
}

export interface SiteFormData {
    domain?: string
    name: string
    type: 'static' | 'reverse_proxy'
    id: string
    upstream: string
    root: string
    indexNames: string
    health_check: boolean
}

export interface Certificate {
    domain: string
    issuer: string
    expiry: string
    status: 'valid' | 'invalid' | 'expiring'
    serialNumber?: string
    fingerprint?: string
    pem?: string
}

export interface SiteItem {
    id: string
    name: string
    domain?: string
}

export interface CertFormData {
    authMethod: '' | 'auto' | 'file'
    targetType: '' | 'domain' | 'site'
    selectedDomain: string
    selectedSite: string
    domain: string
    certFile: string
    keyFile: string
    autoHTTPS: boolean
    certId: string
}

export interface LocalSettings {
    caddy: {
        unixSocket: string
        adminPort: number
    }
    theme: 'light' | 'dark' | 'auto'
    language: 'zh-CN' | 'en-US'
}

export interface LocalConfig {
    unixSocket: string
    adminPort: number
}

export interface CaddyInfo {
    installed: boolean
    running: boolean
    version: string
}

export interface HttpLogEntry {
    timestamp: string
    level: 'INFO' | 'WARN' | 'ERROR'
    host: string
    method: string
    path: string
    status: number
    duration: number
    size: number
    message?: string
}
