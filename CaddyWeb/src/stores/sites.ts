import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { siteAPI } from '@/api'

export interface Site {
  id: string
  label?: string
  status: 'active' | 'inactive' | 'error'
  createdAt: string
  updatedAt: string
}

export interface SiteConfig {
  match: string[]
  handle: Handle[]
}

export interface Handle {
  handler: string
  [key: string]: any
}

export interface Upstream {
  dial: string
}

export const useSitesStore = defineStore('sites', () => {
  const sites = ref<Site[]>([])
  const currentSite = ref<Site | null>(null)
  const currentConfig = ref<SiteConfig | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activeSites = computed(() => sites.value.filter(s => s.status === 'active'))
  const inactiveSites = computed(() => sites.value.filter(s => s.status === 'inactive'))

  // 加载站点列表
  const fetchSites = async () => {
    loading.value = true
    error.value = null
    try {
      const data = await siteAPI.list()
      sites.value = data
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  // 获取站点详情
  const fetchSite = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      const data = await siteAPI.get(id)
      currentSite.value = data.site
      currentConfig.value = data.config
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  // 创建站点
  const createSite = async (siteData: any) => {
    loading.value = true
    error.value = null
    try {
      const data = await siteAPI.create(siteData)
      sites.value.push(data)
      return data
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 更新站点
  const updateSite = async (id: string, siteData: any) => {
    loading.value = true
    error.value = null
    try {
      const data = await siteAPI.update(id, siteData)
      const index = sites.value.findIndex(s => s.id === id)
      if (index !== -1) {
        sites.value[index] = data
      }
      return data
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 删除站点
  const deleteSite = async (domain: string, id: string) => {
    loading.value = true
    error.value = null
    try {
      await siteAPI.delete(domain, id)
      sites.value = sites.value.filter(s => s.id !== id)
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 更新站点配置
  const updateSiteConfig = async (id: string, config: SiteConfig) => {
    loading.value = true
    error.value = null
    try {
      const data = await siteAPI.update(id, { config })
      currentConfig.value = config
      return data
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 添加反向代理处理
  const addReverseProxy = async (siteId: string, upstream: string, pathRewrite?: string) => {
    loading.value = true
    error.value = null
    try {
      const handle: Handle = {
        handler: 'reverse_proxy',
        upstreams: [{ dial: upstream }]
      }
      if (pathRewrite) {
        handle.handle_response = [{ expression: '{http.vars.token}', status_code: 401 }]
      }
      await siteAPI.update(siteId, { handle })
      await fetchSite(siteId)
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 清空当前站点
  const clearCurrentSite = () => {
    currentSite.value = null
    currentConfig.value = null
  }

  return {
    sites,
    currentSite,
    currentConfig,
    loading,
    error,
    activeSites,
    inactiveSites,
    fetchSites,
    fetchSite,
    createSite,
    updateSite,
    deleteSite,
    updateSiteConfig,
    addReverseProxy,
    clearCurrentSite
  }
})
