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
                            <el-input v-model="settings.caddy.unixSocket" placeholder="/var/run/caddy/caddy.sock" />
                            <div class="form-tip">Caddy2 Admin API 的 Unix Socket 路径</div>
                        </el-form-item>

                        <el-form-item label="Admin 端口">
                            <el-input-number v-model="settings.caddy.adminPort" :min="1" :max="65535" />
                            <div class="form-tip">当不使用 Unix Socket 时使用的 TCP 端口</div>
                        </el-form-item>
                    </el-form>
                </el-tab-pane>

                <el-tab-pane label="界面设置" name="ui">
                    <el-form label-width="100px" class="settings-form">
                        <el-divider content-position="left">外观</el-divider>

                        <el-form-item label="主题">
                            <el-radio-group v-model="settings.theme">
                                <el-radio value="light">浅色</el-radio>
                                <el-radio value="dark">深色</el-radio>
                            </el-radio-group>
                        </el-form-item>

                        <el-form-item label="语言">
                            <el-select v-model="settings.language">
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

                        <el-divider content-position="left">重载方式</el-divider>

                        <el-form-item label="重载模式">
                            <el-radio-group v-model="settings.reloadMode">
                                <el-radio value="auto">自动重载</el-radio>
                                <el-radio value="manual">手动重载</el-radio>
                            </el-radio-group>
                            <div class="form-tip">自动：配置变更后自动重载；手动：配置变更后需手动点击重载按钮</div>
                        </el-form-item>

                        <el-divider content-position="left">操作</el-divider>

                        <el-form-item>
                            <el-button type="primary" @click="reloadCaddy">重载配置</el-button>
                            <el-button @click="checkCaddyStatus">检测连接</el-button>
                        </el-form-item>
                    </el-form>
                </el-tab-pane>

                <el-tab-pane label="关于" name="about">
                    <div class="about-content">
                        <el-card shadow="never">
                            <div class="about-header">
                                <h2>🚀 CaddyWeb</h2>
                                <p class="version">v1.0.0</p>
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
import { ref, reactive, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { settingsAPI, caddyAPI } from '@/api'
import { ElMessage } from 'element-plus'

const settingsStore = useSettingsStore()

const activeTab = ref('basic')
const caddyStatus = ref<'running' | 'stopped'>('stopped')
const caddyVersion = ref('未知')

const settings = reactive({
    caddy: {
        unixSocket: '/var/run/caddy/caddy.sock',
        adminPort: 2019
    },
    theme: 'light' as 'light' | 'dark',
    language: 'zh-CN' as 'zh-CN' | 'en-US',
    reloadMode: 'auto' as 'auto' | 'manual'
})

const loadSettings = async () => {
    try {
        const data = await settingsAPI.get()
        Object.assign(settings, {
            caddy: {
                unixSocket: data.caddy?.unixSocket || '/var/run/caddy/caddy.sock',
                adminPort: data.caddy?.adminPort || 2019
            },
            theme: data.theme || 'light',
            language: data.language || 'zh-CN',
            reloadMode: data.reloadMode || 'auto'
        })
        settingsStore.updateSettings(settings)
    } catch (error) {
        console.error('Failed to load settings from backend:', error)
    }
}

const saveSettings = async () => {
    try {
        await settingsAPI.save(settings)
        settingsStore.updateSettings(settings)
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

const reloadCaddy = async () => {
    try {
        await settingsAPI.reloadCaddy()
        ElMessage.success('Caddy2 配置重载成功')
    } catch (error) {
        ElMessage.error('配置重载失败')
    }
}

onMounted(() => {
    loadSettings()
    checkCaddyStatus()
})
</script>

<style scoped lang="scss">
.settings-container {
    padding: 20px;
}

.settings-tabs {
    min-height: 500px;

    :deep(.el-tab-pane) {
        padding-left: 20px;
    }
}

.settings-form {
    max-width: 600px;
    margin-top: 20px;
}

.form-tip {
    font-size: 12px;
    color: #909399;
    margin-top: 5px;
    line-height: 1.4;
}

.settings-footer {
    margin-top: 30px;
    padding-top: 20px;
    border-top: 1px solid #eee;
    display: flex;
    gap: 15px;
}

.about-content {
    max-width: 500px;

    .about-header {
        text-align: center;
        margin-bottom: 20px;

        h2 {
            margin: 0 0 10px;
            color: #303133;
        }

        .version {
            color: #909399;
            margin: 0;
        }
    }

    .description {
        text-align: center;
        color: #606266;
        margin-bottom: 30px;
    }
}
</style>
