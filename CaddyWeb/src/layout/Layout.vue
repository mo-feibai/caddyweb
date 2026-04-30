<template>
    <el-container class="layout-container">
        <el-aside width="240px" class="sidebar">
            <div class="logo">
                <div class="logo-icon">
                    <svg width="28" height="28" viewBox="0 0 24 24" fill="none">
                        <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="var(--accent-cyan)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        <path d="M2 17L12 22L22 17" stroke="var(--accent-cyan)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        <path d="M2 12L12 17L22 12" stroke="var(--accent-cyan)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                </div>
                <div class="logo-text">
                    <span class="logo-name">CaddyWeb</span>
                    <span class="logo-version">v2.0</span>
                </div>
            </div>

            <el-menu :default-active="activeMenu" router class="sidebar-menu">
                <el-menu-item index="/web/dashboard">
                    <el-icon><i-ep-data-analysis /></el-icon>
                    <span>仪表盘</span>
                </el-menu-item>
                <el-menu-item index="/web/domains">
                    <el-icon><i-ep-link /></el-icon>
                    <span>域名管理</span>
                </el-menu-item>
                <el-menu-item index="/web/sites">
                    <el-icon><i-ep-folder /></el-icon>
                    <span>子站点</span>
                </el-menu-item>
                <el-menu-item index="/web/tls">
                    <el-icon><i-ep-lock /></el-icon>
                    <span>SSL 证书</span>
                </el-menu-item>
                <el-menu-item index="/web/logs">
                    <el-icon><i-ep-document /></el-icon>
                    <span>访问日志</span>
                </el-menu-item>

                <div class="menu-divider"></div>

                <el-menu-item index="/web/settings">
                    <el-icon><i-ep-setting /></el-icon>
                    <span>系统设置</span>
                </el-menu-item>
            </el-menu>

            <div class="sidebar-footer">
                <div class="status-dot" :class="caddyStatus"></div>
                <span class="status-label">{{ caddyStatus === 'running' ? 'Caddy2 运行中' : 'Caddy2 已停止' }}</span>
            </div>
        </el-aside>

        <el-container class="main-wrapper">
            <el-header class="header">
                <div class="header-left">
                    <div class="breadcrumb-wrapper">
                        <el-breadcrumb separator="/">
                            <el-breadcrumb-item :to="{ path: '/web/dashboard' }">
                                <span class="breadcrumb-home">首页</span>
                            </el-breadcrumb-item>
                            <el-breadcrumb-item v-if="currentRoute">{{ currentRoute }}</el-breadcrumb-item>
                        </el-breadcrumb>
                    </div>
                </div>

                <div class="header-right">
                    <div class="caddy-indicator" :class="caddyStatus" @click="checkCaddyStatus">
                        <div class="indicator-dot"></div>
                        <span class="indicator-text">{{ caddyStatus === 'running' ? '运行中' : '已停止' }}</span>
                    </div>

                    <el-dropdown @command="handleCommand" trigger="click">
                        <button class="user-btn">
                            <div class="user-avatar">
                                <el-icon><i-ep-user /></el-icon>
                            </div>
                        </button>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item command="settings">
                                    <el-icon><i-ep-setting /></el-icon>
                                    <span>设置</span>
                                </el-dropdown-item>
                                <el-dropdown-item command="logout" divided>
                                    <el-icon><i-ep-switch-button /></el-icon>
                                    <span>退出</span>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                </div>
            </el-header>

            <el-main class="main-content">
                <router-view />
            </el-main>
        </el-container>
    </el-container>
</template>

<script setup lang="ts">
import type { CaddyStatus } from '@/types'
import { useSettingsStore } from '@/stores/settings'
import { settingsAPI } from '@/api'

const route = useRoute()
const router = useRouter()
const settingsStore = useSettingsStore()

const caddyStatus = shallowRef<CaddyStatus>('stopped')

const activeMenu = computed(() => route.path)

const currentRoute = computed(() => {
    const name = route.name as string
    const routeMap: Record<string, string> = {
        'Dashboard': '仪表盘',
        'Domains': '域名管理',
        'Sites': '子站点',
        'TLS': 'SSL 证书',
        'Logs': '访问日志',
        'Settings': '系统设置'
    }
    return routeMap[name] || ''
})

const checkCaddyStatus = async () => {
    try {
        const res = await settingsAPI.getCaddyStatus()
        caddyStatus.value = res.status === 'running' ? 'running' : 'stopped'
        ElMessage.success('Caddy2 运行正常')
    } catch {
        caddyStatus.value = 'stopped'
        ElMessage.error('无法连接到 Caddy2')
    }
}

const handleCommand = (command: string) => {
    switch (command) {
        case 'settings':
            router.push('/web/settings')
            break
        case 'logout':
            router.push('/login')
            break
    }
}

onMounted(() => {
    settingsStore.loadSettings()
    checkCaddyStatus()
})
</script>

<style scoped lang="scss">
.layout-container {
    height: 100vh;
}

.sidebar {
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
}

.logo {
    padding: 24px 20px;
    display: flex;
    align-items: center;
    gap: 14px;
    border-bottom: 1px solid var(--border-subtle);
    background: linear-gradient(180deg, var(--bg-card) 0%, var(--bg-secondary) 100%);

    .logo-icon {
        width: 44px;
        height: 44px;
        background: var(--bg-card);
        border: 1px solid var(--border-subtle);
        border-radius: var(--radius-md);
        display: flex;
        align-items: center;
        justify-content: center;
        box-shadow: var(--glow-cyan);
    }

    .logo-text {
        display: flex;
        flex-direction: column;

        .logo-name {
            font-size: 18px;
            font-weight: 700;
            color: var(--text-primary);
            letter-spacing: -0.5px;
        }

        .logo-version {
            font-family: var(--font-mono);
            font-size: 10px;
            color: var(--accent-cyan);
            text-transform: uppercase;
            letter-spacing: 1px;
        }
    }
}

.sidebar-menu {
    flex: 1;
    padding: 16px 12px;
}

.menu-divider {
    height: 1px;
    background: var(--border-subtle);
    margin: 12px 8px;
}

.sidebar-footer {
    padding: 16px 20px;
    border-top: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    gap: 10px;
    background: var(--bg-secondary);

    .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: var(--accent-magenta);

        &.running {
            background: var(--accent-cyan);
            box-shadow: 0 0 8px var(--accent-cyan);
        }
    }

    .status-label {
        font-family: var(--font-mono);
        font-size: 11px;
        color: var(--text-secondary);
    }
}

.main-wrapper {
    display: flex;
    flex-direction: column;
}

.header {
    background: var(--bg-card);
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 28px;
    height: 64px;
}

.breadcrumb-wrapper {
    :deep(.el-breadcrumb__inner) {
        font-weight: 500;
    }
}

.header-right {
    display: flex;
    align-items: center;
    gap: 20px;
}

.caddy-indicator {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 14px;
    background: var(--bg-hover);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
        border-color: var(--border-active);
        background: var(--bg-active);
    }

    .indicator-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: var(--accent-magenta);
        transition: all 0.2s;
    }

    &.running .indicator-dot {
        background: var(--accent-cyan);
        box-shadow: 0 0 8px var(--accent-cyan);
    }

    .indicator-text {
        font-family: var(--font-mono);
        font-size: 12px;
        color: var(--text-secondary);
    }
}

.user-btn {
    background: var(--bg-hover);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    width: 40px;
    height: 40px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;

    &:hover {
        border-color: var(--accent-cyan);
        background: var(--bg-active);
    }

    .user-avatar {
        color: var(--text-secondary);
    }
}

.main-content {
    background: var(--bg-primary);
    padding: 28px;
    overflow-y: auto;
    flex: 1;
}
</style>