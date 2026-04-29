import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

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
    firstLaunch: true
  })

  const isLocalMode = computed(() => settings.value.deployedMode === 'local')
  const isRemoteMode = computed(() => settings.value.deployedMode === 'remote')

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

  const saveSettings = () => {
    localStorage.setItem('caddyweb-settings', JSON.stringify(settings.value))
  }

  const updateSettings = (newSettings: Partial<AppSettings>) => {
    settings.value = { ...settings.value, ...newSettings }
    saveSettings()
  }

  const updateCaddySettings = (caddySettings: Partial<CaddySettings>) => {
    settings.value.caddy = { ...settings.value.caddy, ...caddySettings }
    saveSettings()
  }

  const setDeployedMode = (mode: 'local' | 'remote') => {
    settings.value.deployedMode = mode
    if (mode === 'local') {
      settings.value.apiBaseUrl = 'http://localhost:8081'
      settings.value.wsBaseUrl = 'http://localhost:8081'
    }
    saveSettings()
  }

  const completeSetup = () => {
    settings.value.firstLaunch = false
    saveSettings()
  }

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
      firstLaunch: true
    }
    localStorage.removeItem('caddyweb-settings')
  }

  return {
    settings,
    isLocalMode,
    isRemoteMode,
    loadSettings,
    saveSettings,
    updateSettings,
    updateCaddySettings,
    setDeployedMode,
    completeSetup,
    resetSettings
  }
})