<script setup lang="ts">
import { computed, watch, onMounted } from 'vue'

import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import { usePreferredDark } from '@vueuse/core'
import { useSettingsStore } from '@/stores/settings'

const settingsStore = useSettingsStore()
const prefersDark = usePreferredDark()

const isDark = computed(() => {
  const theme = settingsStore.settings.theme
  if (theme === 'auto') {
    return prefersDark.value
  }
  return theme === 'dark'
})

const currentThemeConfig = {
  locale: zhCn,
  size: 'default' as const
}

watch(isDark, (dark) => {
  if (dark) {
    document.documentElement.classList.add('dark')
    document.documentElement.classList.remove('light')
  } else {
    document.documentElement.classList.remove('dark')
    document.documentElement.classList.add('light')
  }
}, { immediate: true })

onMounted(() => {
  settingsStore.loadSettings()
})
</script>

<template>
  <el-config-provider :locale="currentThemeConfig.locale" :size="currentThemeConfig.size ">
    <router-view />
  </el-config-provider>
</template>

<style>
html, body, #app {
  height: 100%;
  margin: 0;
  padding: 0;
}
</style>