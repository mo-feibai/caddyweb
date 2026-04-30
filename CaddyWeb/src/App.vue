<template>
  <el-config-provider :locale="zhCn" :theme="currentThemeConfig">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">

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

const currentThemeConfig = computed(() => ({
  locale: zhCn,
  size: 'default'
}))

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

<style>
html, body, #app {
  height: 100%;
  margin: 0;
  padding: 0;
}
</style>