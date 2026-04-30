export type SiteType = 'static' | 'reverse_proxy'

export type DeployedMode = 'local' | 'remote'

export type Theme = 'light' | 'dark' | 'auto'

export type Language = 'zh-CN' | 'en-US'

export type CaddyStatus = 'running' | 'stopped'

export type CaddyStatusWithChecking = 'running' | 'checking' | 'stopped'

export type CheckStatus = 'checking' | 'installed' | 'not_installed'

export type DetectStatus = 'checking' | 'success' | 'failed'

export type InitStatus = 'none' | 'checking' | 'success' | 'warning' | 'failed'

export type LogLevel = 'INFO' | 'WARN' | 'ERROR'

export type CertStatus = 'valid' | 'invalid' | 'expiring'

export type AuthMethod = '' | 'auto' | 'file'

export type TargetType = '' | 'domain' | 'site'

export const SITE_TYPES = ['static', 'reverse_proxy'] as const

export const THEMES = ['light', 'dark', 'auto'] as const

export const LANGUAGES = ['zh-CN', 'en-US'] as const

export const LOG_LEVELS = ['INFO', 'WARN', 'ERROR'] as const

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
    type: SiteType
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
    status: CertStatus
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
    authMethod: AuthMethod
    targetType: TargetType
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
    theme: Theme
    language: Language
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
    level: LogLevel
    host: string
    method: string
    path: string
    status: number
    duration: number
    size: number
    message?: string
}
