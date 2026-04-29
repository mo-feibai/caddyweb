<template>
    <div class="settings-container">
        <el-card>
            <template #header>
                <span>系统设置</span>
            </template>

            <el-tabs v-model="activeTab" tab-position="left" class="settings-tabs">
                <el-tab-pane label="基础设置" name="basic">
                    <el-form label-width="140px" class="settings-form">
                        <el-divider content-position="left">连接配置</el-divider>

                        <el-form-item label="Unix Socket">
                            <el-input v-model="localSettings.caddy.unixSocket" placeholder="/var/run/caddy/caddy.sock" />
                            <div class="form-tip">Caddy2 Admin API 的 Unix Socket 路径</div>
                        </el-form-item>

                        <el-form-item label="Admin 端口">
                            <el-input-number v-model="localSettings.caddy.adminPort" :min="1" :max="65535" />
                            <div class="form-tip">当不使用 Unix Socket 时使用的 TCP 端口</div>
                        </el-form-item>
                    </el-form>
                </el-tab-pane>

                <el-tab-pane label="界面设置" name="ui">
                    <el-form label-width="100px" class="settings-form">
                        <el-divider content-position="left">外观</el-divider>

                        <el-form-item label="主题">
                            <el-radio-group v-model="localSettings.theme">
                                <el-radio value="light">浅色</el-radio>
                                <el-radio value="dark">深色</el-radio>
                                <el-radio value="auto">跟随系统</el-radio>
                            </el-radio-group>
                        </el-form-item>

                        <el-form-item label="语言">
                            <el-select v-model="localSettings.language">
                                <el-option label="简体中文" value="zh-CN" />
                                <el-option label="English" value="en-US" />
                            </el-select>
                        </el-form-item>
                    </el-form>
                </el-tab-pane>

                <el-tab-pane label="Caddy2 管理" name="caddy">
                    <el-form label-width="140px" class="settings-form">
                        <el-divider content-position="left">Caddy2 状态</el-divider>

                        <el-form-item label="运行状态">
                            <el-tag :type="caddyStatus === 'running' ? 'success' : 'danger'">
                                {{ caddyStatus === 'running' ? '运行中' : '已停止' }}
                            </el-tag>
                        </el-form-item>

                        <el-form-item label="版本">
                            <span>{{ caddyVersion }}</span>
                        </el-form-item>

                        <el-divider content-position="left">操作</el-divider>

                        <el-form-item>
                            <el-button @click="checkCaddyStatus">检测连接</el-button>
                        </el-form-item>
                    </el-form>
                </el-tab-pane>

                <el-tab-pane label="关于" name="about">
                    <div class="about-content">
                        <el-card shadow="never">
                            <div class="about-header">
                                <h2>CaddyWeb</h2>
                                <p class="version">v2.0.0</p>
                            </div>
                            <p class="description">
                                基于 Vue3 + Go + Caddy2 的代理配置管理平台
                            </p>
                            <el-descriptions :column="1" border class="about-info">
                                <el-descriptions-item label="前端框架">Vue 3.4 + TypeScript</el-descriptions-item>
                                <el-descriptions-item label="UI 组件">Element Plus</el-descriptions-item>
                                <el-descriptions-item label="后端框架">Go + Gin</el-descriptions-item>
                                <el-descriptions-item label="代理服务器">Caddy2</el-descriptions-item>
                            </el-descriptions>
                        </el-card>
                    </div>
                </el-tab-pane>
            </el-tabs>

            <div class="settings-footer">
                <el-button type="primary" @click="saveSettings">保存设置</el-button>
                <el-button @click="resetSettings">重置</el-button>
            </div>
        </el-card>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { settingsAPI } from '@/api'
import { ElMessage } from 'element-plus'

const settingsStore = useSettingsStore()

const activeTab = ref('basic')
const caddyStatus = ref<'running' | 'stopped'>('stopped')
const caddyVersion = ref('未知')

const localSettings = reactive({
    caddy: {
        unixSocket: '/var/run/caddy/caddy.sock',
        adminPort: 2019
    },
    theme: 'auto' as 'light' | 'dark' | 'auto',
    language: 'zh-CN' as 'zh-CN' | 'en-US'
})

watch(() => settingsStore.settings, (newSettings) => {
    localSettings.caddy.unixSocket = newSettings.caddy.unixSocket
    localSettings.caddy.adminPort = newSettings.caddy.adminPort
    localSettings.theme = newSettings.theme
    localSettings.language = newSettings.language
}, { immediate: true, deep: true })

const loadSettings = async () => {
    try {
        const data = await settingsAPI.get()
        Object.assign(localSettings, {
            caddy: {
                unixSocket: data.caddy?.unixSocket || '/var/run/caddy/caddy.sock',
                adminPort: data.caddy?.adminPort || 2019
            },
            theme: data.theme || 'auto',
            language: data.language || 'zh-CN'
        })
        settingsStore.updateSettings(localSettings)
    } catch (error) {
        console.error('Failed to load settings from backend:', error)
    }
}

const saveSettings = async () => {
    try {
        await settingsAPI.save(localSettings)
        settingsStore.updateSettings(localSettings)
        ElMessage.success('设置已保存')
    } catch (error) {
        ElMessage.error('保存设置失败')
    }
}

const resetSettings = async () => {
    try {
        await settingsAPI.reset?.()
        settingsStore.resetSettings()
        loadSettings()
        ElMessage.info('设置已重置')
    } catch (error) {
        settingsStore.resetSettings()
        loadSettings()
        ElMessage.info('设置已重置')
    }
}

const checkCaddyStatus = async () => {
    try {
        const response = await settingsAPI.detectCaddy()
        if (response.success) {
            caddyStatus.value = 'running'
            ElMessage.success('Caddy2 连接正常')
        } else {
            caddyStatus.value = 'stopped'
            ElMessage.warning(response.message || 'Caddy2 未运行')
        }
    } catch (error: any) {
        caddyStatus.value = 'stopped'
        ElMessage.error('无法连接到 Caddy2')
    }
}

onMounted(() => {
    loadSettings()
    checkCaddyStatus()
})
</script>

<style scoped lang="scss">
.settings-container {
    padding: 24px;
}

.settings-footer {
    margin-top: 30px;
    padding-top: 20px;
    border-top: 1px solid var(--border-subtle);
    display: flex;
    gap: 15px;
}

.about-content {
    max-width: 500px;

    .about-header {
        text-align: center;
        margin-bottom: 24px;

        h2 {
            margin: 0 0 10px;
            color: var(--text-primary);
            font-size: 24px;
        }

        .version {
            font-family: var(--font-mono);
            color: var(--el-color-primary);
            font-size: 14px;
            margin: 0;
        }
    }

    .description {
        text-align: center;
        color: var(--text-secondary);
        margin-bottom: 30px;
        line-height: 1.6;
    }
}
</style>