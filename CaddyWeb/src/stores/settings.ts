import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { settingsAPI } from '@/api'

interface CaddySettings {
  apiUrl: string
  unixSocket: string
  adminPort: number
}

interface AppSettings {
  deployedMode: 'local' | 'remote'
  caddy: CaddySettings
  apiBaseUrl: string
  wsBaseUrl: string
  theme: 'light' | 'dark'
  language: 'zh-CN' | 'en-US'
  firstLaunch: boolean
  reloadMode: 'auto' | 'manual'
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<AppSettings>({
    deployedMode: 'local',
    caddy: {
      apiUrl: '',
      unixSocket: '/var/run/caddy/caddy.sock',
      adminPort: 2019
    },
    apiBaseUrl: 'http://localhost:8081',
    wsBaseUrl: 'http://localhost:8081',
    theme: 'light',
    language: 'zh-CN',
    firstLaunch: true,
    reloadMode: 'auto'
  })

  const needsReload = ref(false)

  const isLocalMode = computed(() => settings.value.deployedMode === 'local')
  const isRemoteMode = computed(() => settings.value.deployedMode === 'remote')
  const isAutoReload = computed(() => settings.value.reloadMode === 'auto')

  const autoReload = async () => {
    if (settings.value.reloadMode === 'auto') {
      try {
        await settingsAPI.reloadCaddy()
      } catch (error) {
        console.error('Auto reload failed:', error)
      }
    }
  }

  // 从本地存储加载设置
  const loadSettings = () => {
    const saved = localStorage.getItem('caddyweb-settings')
    if (saved) {
      try {
        const parsed = JSON.parse(saved)
        settings.value = { ...settings.value, ...parsed }
      } catch (e) {
        console.error('Failed to load settings:', e)
      }
    }
  }

  // 保存设置到本地存储
  const saveSettings = () => {
    localStorage.setItem('caddyweb-settings', JSON.stringify(settings.value))
  }

  // 更新设置
  const updateSettings = (newSettings: Partial<AppSettings>) => {
    settings.value = { ...settings.value, ...newSettings }
    saveSettings()
  }

  // 更新 Caddy 配置
  const updateCaddySettings = (caddySettings: Partial<CaddySettings>) => {
    settings.value.caddy = { ...settings.value.caddy, ...caddySettings }
    saveSettings()
  }

  // 设置部署模式
  const setDeployedMode = (mode: 'local' | 'remote') => {
    settings.value.deployedMode = mode
    if (mode === 'local') {
      settings.value.apiBaseUrl = 'http://localhost:8081'
      settings.value.wsBaseUrl = 'http://localhost:8081'
    }
    saveSettings()
  }

  // 完成初始设置向导
  const completeSetup = () => {
    settings.value.firstLaunch = false
    saveSettings()
  }

  // 标记需要重载
  const markNeedsReload = () => {
    needsReload.value = true
  }

  // 重置重载标记
  const clearNeedsReload = () => {
    needsReload.value = false
  }

  // 重置设置
  const resetSettings = () => {
    settings.value = {
      deployedMode: 'local',
      caddy: {
        apiUrl: '',
        unixSocket: '/var/run/caddy/caddy.sock',
        adminPort: 2019
      },
      apiBaseUrl: 'http://localhost:8081',
      wsBaseUrl: 'http://localhost:8081',
      theme: 'light',
      language: 'zh-CN',
      firstLaunch: true,
      reloadMode: 'auto'
    }
    needsReload.value = false
    localStorage.removeItem('caddyweb-settings')
  }

  return {
    settings,
    needsReload,
    isLocalMode,
    isRemoteMode,
    isAutoReload,
    loadSettings,
    saveSettings,
    updateSettings,
    updateCaddySettings,
    setDeployedMode,
    completeSetup,
    markNeedsReload,
    clearNeedsReload,
    resetSettings,
    autoReload
  }
})
