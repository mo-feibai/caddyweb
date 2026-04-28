<template>
    <el-container class="layout-container">
        <el-aside width="200px" class="sidebar">
            <div class="logo">
                <h3>🚀 CaddyWeb</h3>
            </div>
            <el-menu :default-active="activeMenu" router class="sidebar-menu" background-color="#304156"
                text-color="#bfcbd9" active-text-color="#409EFF">
                <el-menu-item index="/web/dashboard">
                    <el-icon>
                        <DataAnalysis />
                    </el-icon>
                    <span>仪表盘</span>
                </el-menu-item>
                <el-menu-item index="/web/domains">
                    <el-icon>
                        <Link />
                    </el-icon>
                    <span>域名管理</span>
                </el-menu-item>
                <el-menu-item index="/web/sites">
                    <el-icon>
                        <Folder />
                    </el-icon>
                    <span>子站点</span>
                </el-menu-item>
                <el-menu-item index="/web/tls">
                    <el-icon>
                        <Lock />
                    </el-icon>
                    <span>SSL 证书</span>
                </el-menu-item>
                <el-menu-item index="/web/logs">
                    <el-icon>
                        <Document />
                    </el-icon>
                    <span>访问日志</span>
                </el-menu-item>
                <el-divider />
                <el-menu-item index="/web/settings">
                    <el-icon>
                        <Setting />
                    </el-icon>
                    <span>系统设置</span>
                </el-menu-item>
            </el-menu>
        </el-aside>

        <el-container>
            <el-header class="header">
                <div class="header-left">
                    <el-breadcrumb separator="/">
                        <el-breadcrumb-item :to="{ path: '/web/dashboard' }">首页</el-breadcrumb-item>
                        <el-breadcrumb-item v-if="currentRoute">{{ currentRoute }}</el-breadcrumb-item>
                    </el-breadcrumb>
                </div>
                <div class="header-right">
                    <el-button v-if="settingsStore.settings.reloadMode === 'manual' && settingsStore.needsReload" size="small" type="warning" @click="reloadCaddy">
                        <el-icon>
                            <Refresh />
                        </el-icon>
                        <span style="margin-left: 4px">重载 Caddy</span>
                    </el-button>
                    <el-button size="small" @click="checkCaddyStatus" :type="caddyStatus === 'running' ? 'success' : 'danger'">
                        <el-icon>
                            <Refresh />
                        </el-icon>
                        <span style="margin-left: 4px">{{ caddyStatus === 'running' ? 'Caddy2 运行中' : 'Caddy2 已停止' }}</span>
                    </el-button>
                    <el-dropdown @command="handleCommand">
                        <el-button size="small">
                            <el-icon>
                                <User />
                            </el-icon>
                        </el-button>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item command="settings">设置</el-dropdown-item>
                                <el-dropdown-item command="logout" divided>退出</el-dropdown-item>
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
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { caddyAPI, settingsAPI } from '@/api'
import {
    DataAnalysis, Folder, Connection, Lock, Document,
    Setting, Refresh, User, Link
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const settingsStore = useSettingsStore()

const caddyStatus = ref<'running' | 'stopped'>('stopped')

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
        if (res.status === 'running') {
            ElMessage.success('Caddy2 运行正常')
        } else {
            ElMessage.error('Caddy2 未运行')
        }
    } catch (error) {
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

const reloadCaddy = async () => {
    try {
        await settingsAPI.reloadCaddy()
        settingsStore.clearNeedsReload()
        ElMessage.success('Caddy2 配置重载成功')
    } catch (error) {
        ElMessage.error('配置重载失败')
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
    background-color: #304156;

    .logo {
        height: 60px;
        display: flex;
        align-items: center;
        justify-content: center;
        background-color: #2b3a4b;

        h3 {
            color: #fff;
            margin: 0;
            font-size: 18px;
        }
    }

    .sidebar-menu {
        border-right: none;

        :deep(.el-menu-item) {
            &.is-active {
                background-color: #263445 !important;
            }

            &:hover {
                background-color: #263445 !important;
            }
        }
    }

    :deep(.el-divider) {
        margin: 10px 0;
        background-color: #4a5568;
    }
}

.header {
    background-color: #fff;
    box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20px;

    .header-left {
        display: flex;
        align-items: center;
    }

    .header-right {
        display: flex;
        align-items: center;
        gap: 15px;
    }
}

.main-content {
    background-color: #f0f2f5;
    padding: 20px;
    overflow-y: auto;
}
</style>
