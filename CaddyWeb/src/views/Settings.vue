<template>
    <div class="settings-view">
        <div class="settings-bg-grid"></div>
        <div class="settings-bg-glow"></div>

        <div class="settings-header">
            <div class="header-icon">
                <el-icon :size="28"><i-ep-setting /></el-icon>
            </div>
            <div class="header-text">
                <h1 class="header-title">系统设置</h1>
                <p class="header-subtitle">Configure your CaddyWeb experience</p>
            </div>
        </div>

        <div class="settings-layout">
            <nav class="settings-nav">
                <button
                    v-for="(tab, index) in tabs"
                    :key="tab.name"
                    class="nav-item"
                    :class="{ active: activeTab === tab.name }"
                    @click="activeTab = tab.name"
                    :style="{ animationDelay: `${index * 60}ms` }"
                >
                    <span class="nav-indicator"></span>
                    <el-icon :size="18" class="nav-icon"><component :is="'i-ep-' + tab.icon.toLowerCase()" /></el-icon>
                    <span class="nav-label">{{ tab.label }}</span>
                    <span class="nav-accent"></span>
                </button>
            </nav>

            <main class="settings-content">
                <Transition name="tab-fade" mode="out-in">
                    <section v-if="activeTab === 'basic'" key="basic" class="settings-section">
                        <div class="section-header">
                            <div class="section-icon">
                                <el-icon><i-ep-connection /></el-icon>
                            </div>
                            <div class="section-titles">
                                <h2 class="section-title">连接配置</h2>
                                <p class="section-desc">Caddy2 Admin API 连接参数</p>
                            </div>
                        </div>

                        <div class="form-card">
                            <div class="form-row">
                                <div class="form-group">
                                    <label class="form-label">
                                        <span class="label-dot"></span>
                                        Unix Socket
                                    </label>
                                    <div class="input-wrapper">
                                        <el-icon class="input-prefix"><i-ep-document /></el-icon>
                                        <el-input
                                            v-model="localSettings.caddy.unixSocket"
                                            placeholder="/var/run/caddy/caddy.sock"
                                        />
                                        <div class="input-line"></div>
                                    </div>
                                    <p class="form-hint">Caddy2 Admin API 的 Unix Socket 路径</p>
                                </div>
                            </div>

                            <div class="form-row">
                                <div class="form-group">
                                    <label class="form-label">
                                        <span class="label-dot"></span>
                                        Admin 端口
                                    </label>
                                    <div class="input-wrapper">
                                        <el-icon class="input-prefix"><i-ep-connection /></el-icon>
                                        <el-input-number
                                            v-model="localSettings.caddy.adminPort"
                                            :min="1"
                                            :max="65535"
                                            controls-position="right"
                                        />
                                        <div class="input-line"></div>
                                    </div>
                                    <p class="form-hint">当不使用 Unix Socket 时使用的 TCP 端口</p>
                                </div>
                            </div>
                        </div>
                    </section>

                    <section v-else-if="activeTab === 'ui'" key="ui" class="settings-section">
                        <div class="section-header">
                            <div class="section-icon">
                                <el-icon><i-ep-magic-stick /></el-icon>
                            </div>
                            <div class="section-titles">
                                <h2 class="section-title">外观</h2>
                                <p class="section-desc">界面主题与语言偏好</p>
                            </div>
                        </div>

                        <div class="form-card">
                            <div class="form-row">
                                <div class="form-group">
                                    <label class="form-label">
                                        <span class="label-dot"></span>
                                        主题
                                    </label>
                                    <div class="theme-selector">
                                        <button
                                            v-for="theme in themeOptions"
                                            :key="theme.value"
                                            class="theme-option"
                                            :class="{ active: localSettings.theme === theme.value }"
                                            @click="localSettings.theme = theme.value"
                                        >
                                            <div class="theme-preview" :class="theme.value">
                                                <div class="preview-bar"></div>
                                                <div class="preview-bar short"></div>
                                            </div>
                                            <span class="theme-name">{{ theme.label }}</span>
                                            <div class="theme-check">
                                                <el-icon><i-ep-check /></el-icon>
                                            </div>
                                        </button>
                                    </div>
                                </div>
                            </div>

                            <div class="form-row">
                                <div class="form-group">
                                    <label class="form-label">
                                        <span class="label-dot"></span>
                                        语言
                                    </label>
                                    <div class="language-selector">
                                        <button
                                            v-for="lang in languageOptions"
                                            :key="lang.value"
                                            class="language-option"
                                            :class="{ active: localSettings.language === lang.value }"
                                            @click="localSettings.language = lang.value"
                                        >
                                            <span class="lang-flag">{{ lang.flag }}</span>
                                            <span class="lang-name">{{ lang.label }}</span>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </section>

                    <section v-else-if="activeTab === 'caddy'" key="caddy" class="settings-section">
                        <div class="section-header">
                            <div class="section-icon caddy-icon">
                                <el-icon><i-ep-box /></el-icon>
                            </div>
                            <div class="section-titles">
                                <h2 class="section-title">Caddy2 管理</h2>
                                <p class="section-desc">服务器状态与连接检测</p>
                            </div>
                        </div>

                        <div class="status-dashboard">
                            <div class="status-main-card">
                                <div class="status-ring-wrapper">
                                    <div class="status-ring" :class="caddyStatus">
                                        <div class="ring-track"></div>
                                        <div class="ring-progress"></div>
                                        <div class="ring-center">
                                            <el-icon :size="32" class="ring-icon">
                                                <component :is="caddyStatus === 'running' ? 'CircleCheck' : 'CircleClose'" />
                                            </el-icon>
                                        </div>
                                    </div>
                                    <div class="ring-glow" :class="caddyStatus"></div>
                                </div>
                                <div class="status-details">
                                    <div class="status-main-text" :class="caddyStatus">
                                        {{ caddyStatus === 'running' ? '运行中' : '已停止' }}
                                    </div>
                                    <div class="status-version">
                                        <span class="version-label" v-if="caddyVersion !== '未知' && !caddyVersion.startsWith('v')">v</span>
                                        <span class="version-value">{{ caddyVersion }}</span>
                                    </div>
                                    <div class="status-uptime" v-if="caddyStatus === 'running'">
                                        <el-icon><i-ep-timer /></el-icon>
                                        <span>实时监控</span>
                                    </div>
                                </div>
                            </div>

                            <div class="status-actions-row">
                                <button class="action-card" @click="checkCaddyStatus" :class="{ loading: checking }">
                                    <div class="action-card-icon">
                                        <el-icon><i-ep-refresh /></el-icon>
                                    </div>
                                    <div class="action-card-content">
                                        <span class="action-card-title">检测连接</span>
                                        <span class="action-card-desc">验证 Caddy2 连接状态</span>
                                    </div>
                                    <div class="action-card-arrow">
                                        <el-icon><i-ep-arrow-right /></el-icon>
                                    </div>
                                </button>
                            </div>
                        </div>
                    </section>

                    <section v-else-if="activeTab === 'about'" key="about" class="settings-section about-section">
                        <div class="about-hero">
                            <div class="about-logo">
                                <div class="logo-shape"></div>
                                <div class="logo-shape"></div>
                                <div class="logo-shape"></div>
                                <div class="logo-core">
                                    <el-icon :size="36"><i-ep-box /></el-icon>
                                </div>
                            </div>
                            <h2 class="about-title">CaddyWeb</h2>
                            <p class="about-version">v2.0.0</p>
                            <p class="about-tagline">基于 Vue3 + Go + Caddy2 的代理配置管理平台</p>
                        </div>

                        <div class="about-stack">
                            <div class="stack-item" v-for="(item, i) in stackItems" :key="item.label" :style="{ animationDelay: `${300 + i * 80}ms` }">
                                <div class="stack-icon">
                                    <el-icon><component :is="'i-ep-' + item.icon.toLowerCase()" /></el-icon>
                                </div>
                                <div class="stack-info">
                                    <span class="stack-category">{{ item.category }}</span>
                                    <span class="stack-label">{{ item.label }}</span>
                                </div>
                            </div>
                        </div>

                        <div class="about-footer">
                            <p class="footer-text">Crafted with precision for modern infrastructure</p>
                        </div>
                    </section>
                </Transition>
            </main>
        </div>

        <footer class="settings-footer">
            <div class="footer-saved" v-if="lastSaved">
                <el-icon><i-ep-check /></el-icon>
                <span>已保存 {{ lastSaved }}</span>
            </div>
            <div class="footer-actions">
                <button class="btn-reset" @click="resetSettings">
                    <el-icon><i-ep-refresh-left /></el-icon>
                    <span>重置</span>
                </button>
                <button class="btn-save" @click="saveSettings">
                    <el-icon><i-ep-check /></el-icon>
                    <span>保存设置</span>
                </button>
            </div>
        </footer>
    </div>
</template>

<script setup lang="ts">

import { useSettingsStore } from '@/stores/settings'
import { settingsAPI } from '@/api'

const settingsStore = useSettingsStore()

const activeTab = ref('basic')
const caddyStatus = ref<'running' | 'stopped'>('stopped')
const caddyVersion = ref('未知')
const checking = ref(false)
const lastSaved = ref('')

const tabs = [
    { name: 'basic', label: '基础设置', icon: 'connection' },
    { name: 'ui', label: '界面设置', icon: 'magic-stick' },
    { name: 'caddy', label: 'Caddy2 管理', icon: 'box' },
    { name: 'about', label: '关于', icon: 'info-filled' },
]

const themeOptions = [
    { value: 'light', label: '浅色' },
    { value: 'dark', label: '深色' },
    { value: 'auto', label: '跟随系统' },
]

const languageOptions = [
    { value: 'zh-CN', label: '简体中文', flag: '🇨🇳' },
    { value: 'en-US', label: 'English', flag: '🇺🇸' },
]

const stackItems = [
    { category: '前端框架', label: 'Vue 3.4 + TypeScript', icon: 'monitor' },
    { category: 'UI 组件', label: 'Element Plus', icon: 'grid' },
    { category: '后端框架', label: 'Go + Gin', icon: 'box' },
    { category: '代理服务器', label: 'Caddy2', icon: 'connection' },
]

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
        lastSaved.value = '刚刚'
        setTimeout(() => { lastSaved.value = '' }, 3000)
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
    checking.value = true
    try {
        const response = await settingsAPI.detectCaddy()
        if (response.success) {
            caddyStatus.value = 'running'
            caddyVersion.value = response.version || '未知'
            ElMessage.success('Caddy2 连接正常')
        } else {
            caddyStatus.value = 'stopped'
            caddyVersion.value = '未知'
            ElMessage.warning(response.message || 'Caddy2 未运行')
        }
    } catch (error: any) {
        caddyStatus.value = 'stopped'
        caddyVersion.value = '未知'
        ElMessage.error('无法连接到 Caddy2')
    } finally {
        checking.value = false
    }
}

onMounted(() => {
    loadSettings()
    checkCaddyStatus()
})
</script>

<style scoped lang="scss">
.settings-view {
    min-height: 100vh;
    padding: 32px 40px;
    position: relative;
    overflow: hidden;
}

.settings-bg-grid {
    position: fixed;
    inset: 0;
    background-image:
        linear-gradient(var(--border-subtle) 1px, transparent 1px),
        linear-gradient(90deg, var(--border-subtle) 1px, transparent 1px);
    background-size: 40px 40px;
    opacity: 0.4;
    pointer-events: none;
}

.settings-bg-glow {
    position: fixed;
    top: -200px;
    right: -200px;
    width: 600px;
    height: 600px;
    background: radial-gradient(circle, rgba(0, 212, 255, 0.08) 0%, transparent 70%);
    pointer-events: none;
}

.settings-header {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 40px;
    animation: fadeSlideDown 0.5s ease-out;

    .header-icon {
        width: 56px;
        height: 56px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--bg-card);
        border: 1px solid var(--border-subtle);
        border-radius: 14px;
        color: var(--accent-cyan);
        box-shadow: var(--glow-cyan);
    }

    .header-title {
        font-family: var(--font-display);
        font-size: 26px;
        font-weight: 700;
        color: var(--text-primary);
        margin: 0;
        letter-spacing: -0.5px;
    }

    .header-subtitle {
        font-family: var(--font-mono);
        font-size: 12px;
        color: var(--text-muted);
        margin: 4px 0 0;
        letter-spacing: 0.5px;
    }
}

.settings-layout {
    display: grid;
    grid-template-columns: 220px 1fr;
    gap: 28px;
    animation: fadeSlideUp 0.5s ease-out 0.1s both;
}

.settings-nav {
    display: flex;
    flex-direction: column;
    gap: 6px;
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: 16px;
    padding: 12px;
    height: fit-content;
    position: sticky;
    top: 32px;
}

.nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    border: none;
    background: transparent;
    border-radius: 10px;
    cursor: pointer;
    transition: all var(--transition-smooth);
    position: relative;
    overflow: hidden;
    animation: fadeSlideRight 0.4s ease-out both;

    .nav-indicator {
        position: absolute;
        left: 0;
        top: 50%;
        transform: translateY(-50%) scaleY(0);
        width: 3px;
        height: 60%;
        background: var(--accent-cyan);
        border-radius: 0 2px 2px 0;
        transition: transform var(--transition-smooth);
    }

    .nav-icon {
        color: var(--text-muted);
        transition: color var(--transition-fast);
    }

    .nav-label {
        font-family: var(--font-display);
        font-size: 14px;
        font-weight: 500;
        color: var(--text-secondary);
        transition: color var(--transition-fast);
    }

    .nav-accent {
        position: absolute;
        right: 12px;
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: var(--accent-cyan);
        opacity: 0;
        transform: scale(0);
        transition: all var(--transition-smooth);
    }

    &:hover {
        background: var(--bg-hover);

        .nav-icon { color: var(--text-primary); }
        .nav-label { color: var(--text-primary); }
    }

    &.active {
        background: var(--bg-active);

        .nav-indicator {
            transform: translateY(-50%) scaleY(1);
        }

        .nav-icon { color: var(--accent-cyan); }
        .nav-label { color: var(--text-primary); font-weight: 600; }

        .nav-accent {
            opacity: 1;
            transform: scale(1);
            animation: pulse-dot 2s infinite;
        }
    }
}

.settings-content {
    min-height: 500px;
}

.settings-section {
    animation: fadeSlideUp 0.4s ease-out;
}

.section-header {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--border-subtle);

    .section-icon {
        width: 44px;
        height: 44px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--bg-card);
        border: 1px solid var(--border-subtle);
        border-radius: 12px;
        color: var(--accent-cyan);
        font-size: 20px;
    }

    .section-titles {
        .section-title {
            font-family: var(--font-display);
            font-size: 18px;
            font-weight: 700;
            color: var(--text-primary);
            margin: 0;
        }

        .section-desc {
            font-size: 13px;
            color: var(--text-muted);
            margin: 4px 0 0;
        }
    }
}

.form-card {
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: 16px;
    padding: 28px;
    display: flex;
    flex-direction: column;
    gap: 28px;
}

.form-row {
    display: flex;
    gap: 24px;

    @media (max-width: 768px) {
        flex-direction: column;
    }
}

.form-group {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.form-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-display);
    font-size: 13px;
    font-weight: 600;
    color: var(--text-secondary);
    letter-spacing: 0.3px;

    .label-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: var(--accent-cyan);
    }
}

.input-wrapper {
    position: relative;
    display: flex;
    align-items: center;

    .input-prefix {
        position: absolute;
        left: 14px;
        color: var(--text-muted);
        font-size: 16px;
        z-index: 1;
        pointer-events: none;
    }

    .el-input {
        --el-input-bg-color: var(--bg-secondary);
        --el-input-border-color: var(--border-subtle);
        --el-input-text-color: var(--text-primary);
        --el-input-placeholder-color: var(--text-muted);
        width: 100%;

        :deep(.el-input__wrapper) {
            padding-left: 42px;
            padding-right: 14px;
            background: var(--bg-secondary);
            border: 1px solid var(--border-subtle);
            border-radius: 10px;
            box-shadow: none;
            transition: all var(--transition-smooth);
            position: relative;

            &:hover {
                border-color: var(--text-muted);
            }

            &.is-focus {
                border-color: var(--accent-cyan);
                box-shadow: 0 0 0 3px rgba(0, 212, 255, 0.1);
            }
        }
    }

    .el-input-number {
        width: auto;
        min-width: 140px;

        :deep(.el-input__wrapper) {
            padding-left: 42px;
        }

        :deep(.el-input-number__decrease),
        :deep(.el-input-number__increase) {
            background: var(--bg-hover);
            border-color: var(--border-subtle);
            color: var(--text-secondary);

            &:hover {
                color: var(--accent-cyan);
            }
        }
    }

    .input-line {
        position: absolute;
        bottom: 0;
        left: 50%;
        width: 0;
        height: 2px;
        background: var(--accent-cyan);
        transition: all var(--transition-smooth);
        transform: translateX(-50%);
        border-radius: 2px;
    }
}

.form-hint {
    font-size: 12px;
    color: var(--text-muted);
    margin: 0;
    font-family: var(--font-mono);
}

.theme-selector {
    display: flex;
    gap: 12px;
}

.theme-option {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 14px 20px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-subtle);
    border-radius: 12px;
    cursor: pointer;
    transition: all var(--transition-smooth);
    position: relative;

    .theme-preview {
        width: 48px;
        height: 36px;
        border-radius: 6px;
        padding: 4px;
        display: flex;
        flex-direction: column;
        gap: 3px;

        &.light {
            background: #f0f2f5;
            border: 1px solid #e8ecf1;

            .preview-bar {
                background: #d0d5dd;
                border-radius: 2px;
                height: 4px;
                width: 100%;

                &.short { width: 60%; }
            }
        }

        &.dark {
            background: #12121a;
            border: 1px solid #222230;

            .preview-bar {
                background: #3a3a4a;
                border-radius: 2px;
                height: 4px;
                width: 100%;

                &.short { width: 60%; }
            }
        }

        &.auto {
            background: linear-gradient(135deg, #f0f2f5 50%, #12121a 50%);
            border: 1px solid #d0d5dd;

            .preview-bar {
                background: #909399;
                border-radius: 2px;
                height: 4px;
                width: 100%;

                &.short { width: 60%; }
            }
        }
    }

    .theme-name {
        font-size: 12px;
        font-weight: 500;
        color: var(--text-secondary);
    }

    .theme-check {
        position: absolute;
        top: 8px;
        right: 8px;
        width: 18px;
        height: 18px;
        border-radius: 50%;
        background: var(--accent-cyan);
        display: flex;
        align-items: center;
        justify-content: center;
        opacity: 0;
        transform: scale(0);
        transition: all var(--transition-smooth);
        color: var(--bg-primary);
        font-size: 10px;
    }

    &:hover {
        border-color: var(--text-muted);
        transform: translateY(-2px);
    }

    &.active {
        border-color: var(--accent-cyan);
        box-shadow: var(--glow-cyan);

        .theme-check {
            opacity: 1;
            transform: scale(1);
        }

        .theme-name {
            color: var(--accent-cyan);
        }
    }
}

.language-selector {
    display: flex;
    gap: 12px;
}

.language-option {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 20px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-subtle);
    border-radius: 10px;
    cursor: pointer;
    transition: all var(--transition-smooth);

    .lang-flag {
        font-size: 18px;
    }

    .lang-name {
        font-size: 14px;
        font-weight: 500;
        color: var(--text-secondary);
    }

    &:hover {
        border-color: var(--text-muted);
        transform: translateY(-2px);
    }

    &.active {
        border-color: var(--accent-cyan);
        background: rgba(0, 212, 255, 0.05);

        .lang-name {
            color: var(--accent-cyan);
        }
    }
}

.status-dashboard {
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.status-main-card {
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: 20px;
    padding: 40px;
    display: flex;
    align-items: center;
    gap: 40px;
}

.status-ring-wrapper {
    position: relative;
    width: 120px;
    height: 120px;
    flex-shrink: 0;
}

.status-ring {
    width: 120px;
    height: 120px;
    border-radius: 50%;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;

    .ring-track {
        position: absolute;
        inset: 0;
        border-radius: 50%;
        border: 3px solid var(--border-subtle);
    }

    .ring-progress {
        position: absolute;
        inset: 0;
        border-radius: 50%;
        border: 3px solid transparent;
        border-top-color: var(--text-muted);
        transition: all 0.6s ease;
    }

    .ring-center {
        width: 90px;
        height: 90px;
        border-radius: 50%;
        background: var(--bg-secondary);
        display: flex;
        align-items: center;
        justify-content: center;
        border: 1px solid var(--border-subtle);
    }

    &.running {
        .ring-progress {
            border-top-color: var(--accent-cyan);
            animation: spin-ring 2s linear infinite;
        }

        .ring-center {
            color: var(--accent-cyan);
        }
    }

    &.stopped {
        .ring-progress {
            border-top-color: var(--accent-magenta);
            animation: spin-ring 3s linear infinite;
        }

        .ring-center {
            color: var(--accent-magenta);
        }
    }
}

.ring-glow {
    position: absolute;
    inset: -10px;
    border-radius: 50%;
    opacity: 0;
    transition: opacity 0.6s ease;

    &.running {
        background: radial-gradient(circle, rgba(0, 212, 255, 0.15) 0%, transparent 70%);
        opacity: 1;
    }

    &.stopped {
        background: radial-gradient(circle, rgba(255, 45, 106, 0.1) 0%, transparent 70%);
        opacity: 1;
    }
}

.status-details {
    display: flex;
    flex-direction: column;
    gap: 8px;

    .status-main-text {
        font-family: var(--font-display);
        font-size: 28px;
        font-weight: 700;
        letter-spacing: -0.5px;

        &.running { color: var(--accent-cyan); }
        &.stopped { color: var(--accent-magenta); }
    }

    .status-version {
        font-family: var(--font-mono);
        font-size: 14px;
        color: var(--text-muted);

        .version-label {
            opacity: 0.6;
        }

        .version-value {
            color: var(--text-secondary);
        }
    }

    .status-uptime {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 12px;
        color: var(--text-muted);
        margin-top: 4px;
    }
}

.status-actions-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 14px;
}

.action-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 18px 20px;
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: 14px;
    cursor: pointer;
    transition: all var(--transition-smooth);
    text-align: left;

    .action-card-icon {
        width: 44px;
        height: 44px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--bg-secondary);
        border-radius: 10px;
        color: var(--accent-cyan);
        font-size: 20px;
        transition: all var(--transition-smooth);
    }

    .action-card-content {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 3px;

        .action-card-title {
            font-family: var(--font-display);
            font-size: 15px;
            font-weight: 600;
            color: var(--text-primary);
        }

        .action-card-desc {
            font-size: 12px;
            color: var(--text-muted);
        }
    }

    .action-card-arrow {
        color: var(--text-muted);
        transition: all var(--transition-smooth);
    }

    &:hover {
        border-color: var(--accent-cyan);
        transform: translateX(4px);
        box-shadow: var(--glow-cyan);

        .action-card-icon {
            background: rgba(0, 212, 255, 0.1);
        }

        .action-card-arrow {
            color: var(--accent-cyan);
            transform: translateX(4px);
        }
    }

    &.loading .action-card-icon {
        animation: spin 1s linear infinite;
    }
}

.about-section {
    display: flex;
    flex-direction: column;
    gap: 32px;
}

.about-hero {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 48px 24px;
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: 20px;
    text-align: center;
    position: relative;
    overflow: hidden;

    &::before {
        content: '';
        position: absolute;
        top: -100px;
        left: 50%;
        transform: translateX(-50%);
        width: 300px;
        height: 300px;
        background: radial-gradient(circle, rgba(0, 212, 255, 0.06) 0%, transparent 70%);
        pointer-events: none;
    }
}

.about-logo {
    width: 80px;
    height: 80px;
    position: relative;
    margin-bottom: 24px;

    .logo-shape {
        position: absolute;
        inset: 0;
        border: 2px solid var(--accent-cyan);
        border-radius: 20px;
        opacity: 0.2;
        animation: float-shape 4s ease-in-out infinite;

        &:nth-child(1) { animation-delay: 0s; transform: rotate(0deg); }
        &:nth-child(2) { animation-delay: 0.5s; transform: rotate(45deg) scale(0.85); }
        &:nth-child(3) { animation-delay: 1s; transform: rotate(90deg) scale(0.7); }
    }

    .logo-core {
        position: absolute;
        inset: 16px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--bg-secondary);
        border: 1px solid var(--border-subtle);
        border-radius: 14px;
        color: var(--accent-cyan);
    }
}

.about-title {
    font-family: var(--font-display);
    font-size: 32px;
    font-weight: 700;
    color: var(--text-primary);
    margin: 0 0 8px;
    letter-spacing: -1px;
}

.about-version {
    font-family: var(--font-mono);
    font-size: 14px;
    color: var(--accent-cyan);
    margin: 0 0 16px;
    padding: 4px 12px;
    background: rgba(0, 212, 255, 0.08);
    border-radius: 20px;
    display: inline-block;
}

.about-tagline {
    font-size: 14px;
    color: var(--text-secondary);
    margin: 0;
    line-height: 1.6;
    max-width: 380px;
}

.about-stack {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 14px;

    @media (max-width: 600px) {
        grid-template-columns: 1fr;
    }
}

.stack-item {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 18px 20px;
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: 14px;
    animation: fadeSlideUp 0.4s ease-out both;
    transition: all var(--transition-smooth);

    &:hover {
        border-color: var(--border-active);
        transform: translateY(-2px);
        box-shadow: var(--glow-cyan);
    }

    .stack-icon {
        width: 40px;
        height: 40px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--bg-secondary);
        border-radius: 10px;
        color: var(--accent-cyan);
        font-size: 18px;
    }

    .stack-info {
        display: flex;
        flex-direction: column;
        gap: 2px;

        .stack-category {
            font-size: 11px;
            font-weight: 500;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .stack-label {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-primary);
        }
    }
}

.about-footer {
    text-align: center;
    padding: 16px;

    .footer-text {
        font-family: var(--font-mono);
        font-size: 12px;
        color: var(--text-muted);
        margin: 0;
        letter-spacing: 0.3px;
    }
}

.settings-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 40px;
    padding-top: 24px;
    border-top: 1px solid var(--border-subtle);
    animation: fadeSlideUp 0.5s ease-out 0.2s both;

    .footer-saved {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 13px;
        color: var(--accent-cyan);
        font-family: var(--font-mono);
        animation: fadeIn 0.3s ease-out;
    }

    .footer-actions {
        display: flex;
        gap: 12px;
    }
}

.btn-reset {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 20px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-subtle);
    border-radius: 10px;
    font-family: var(--font-display);
    font-size: 14px;
    font-weight: 500;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all var(--transition-smooth);

    &:hover {
        border-color: var(--text-muted);
        color: var(--text-primary);
        transform: translateY(-1px);
    }
}

.btn-save {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 24px;
    background: var(--accent-cyan);
    border: none;
    border-radius: 10px;
    font-family: var(--font-display);
    font-size: 14px;
    font-weight: 600;
    color: var(--bg-primary);
    cursor: pointer;
    transition: all var(--transition-smooth);

    &:hover {
        background: var(--accent-cyan-dim);
        transform: translateY(-1px);
        box-shadow: var(--glow-cyan);
    }
}

.tab-fade-enter-active,
.tab-fade-leave-active {
    transition: all 0.25s ease;
}

.tab-fade-enter-from {
    opacity: 0;
    transform: translateX(12px);
}

.tab-fade-leave-to {
    opacity: 0;
    transform: translateX(-12px);
}

@keyframes fadeSlideDown {
    from { opacity: 0; transform: translateY(-16px); }
    to { opacity: 1; transform: translateY(0); }
}

@keyframes fadeSlideUp {
    from { opacity: 0; transform: translateY(16px); }
    to { opacity: 1; transform: translateY(0); }
}

@keyframes fadeSlideRight {
    from { opacity: 0; transform: translateX(-12px); }
    to { opacity: 1; transform: translateX(0); }
}

@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}

@keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}

@keyframes spin-ring {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}

@keyframes pulse-dot {
    0%, 100% { opacity: 1; transform: scale(1); }
    50% { opacity: 0.5; transform: scale(0.7); }
}

@keyframes float-shape {
    0%, 100% { opacity: 0.2; }
    50% { opacity: 0.4; }
}
</style>
