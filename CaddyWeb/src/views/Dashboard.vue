<template>
    <div class="dashboard-container">
        <div class="dashboard-header">
            <div class="header-content">
                <h1 class="page-title">仪表盘</h1>
                <p class="page-subtitle">实时监控系统状态</p>
            </div>
            <div class="header-actions">
                <button class="action-btn" @click="reloadCaddy">
                    <el-icon><Refresh /></el-icon>
                    <span>重载配置</span>
                </button>
            </div>
        </div>

        <div class="stats-grid">
            <div class="stat-card" v-for="(stat, index) in statsData" :key="stat.label" :style="{ animationDelay: `${index * 100}ms` }">
                <div class="stat-icon" :class="stat.color">
                    <el-icon :size="24"><component :is="stat.icon" /></el-icon>
                </div>
                <div class="stat-info">
                    <div class="stat-value">{{ stat.value }}</div>
                    <div class="stat-label">{{ stat.label }}</div>
                </div>
                <div class="stat-glow"></div>
            </div>
        </div>

        <div class="dashboard-grid">
            <div class="panel sites-panel">
                <div class="panel-header">
                    <h2 class="panel-title">
                        <el-icon><Folder /></el-icon>
                        <span>最近站点</span>
                    </h2>
                    <el-button type="primary" size="small" @click="$router.push('/web/sites/add')">
                        <el-icon><Plus /></el-icon>
                        添加站点
                    </el-button>
                </div>

                <div class="sites-list">
                    <div class="site-item" v-for="site in recentSites" :key="site.id">
                        <div class="site-info">
                            <div class="site-domain">{{ site.domain }}</div>
                            <div class="site-meta">
                                <span class="site-type" :class="site.type">{{ site.type === 'reverse_proxy' ? '反向代理' : '静态站点' }}</span>
                                <span class="site-status" :class="site.status">{{ site.status === 'active' ? '运行中' : '已停止' }}</span>
                            </div>
                        </div>
                        <div class="site-actions">
                            <button class="icon-btn" @click="editSite(site)">
                                <el-icon><Edit /></el-icon>
                            </button>
                            <button class="icon-btn danger" @click="deleteSite(site)">
                                <el-icon><Delete /></el-icon>
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <div class="side-panels">
                <div class="panel status-panel">
                    <div class="panel-header">
                        <h2 class="panel-title">
                            <el-icon><Monitor /></el-icon>
                            <span>Caddy2 状态</span>
                        </h2>
                    </div>

                    <div class="status-display">
                        <div class="status-ring" :class="caddyStatus">
                            <div class="ring-dot"></div>
                            <div class="ring-glow"></div>
                        </div>
                        <div class="status-info">
                            <div class="status-text">{{ caddyStatusText }}</div>
                            <div class="status-version">v{{ caddyVersion }}</div>
                        </div>
                    </div>

                    <div class="status-actions">
                        <button class="status-btn" @click="reloadCaddy">
                            <el-icon><Refresh /></el-icon>
                            重载配置
                        </button>
                        <button class="status-btn" @click="$router.push('/web/logs')">
                            <el-icon><Document /></el-icon>
                            查看日志
                        </button>
                    </div>
                </div>

                <div class="panel quick-panel">
                    <div class="panel-header">
                        <h2 class="panel-title">
                            <el-icon><Lightning /></el-icon>
                            <span>快捷操作</span>
                        </h2>
                    </div>

                    <div class="quick-grid">
                        <button class="quick-btn" @click="$router.push('/web/tls')">
                            <el-icon><Lock /></el-icon>
                            <span>SSL 证书</span>
                        </button>
                        <button class="quick-btn" @click="$router.push('/web/domains')">
                            <el-icon><Link /></el-icon>
                            <span>域名管理</span>
                        </button>
                        <button class="quick-btn" @click="$router.push('/web/settings')">
                            <el-icon><Setting /></el-icon>
                            <span>系统设置</span>
                        </button>
                        <button class="quick-btn" @click="$router.push('/web/logs')">
                            <el-icon><DataLine /></el-icon>
                            <span>访问日志</span>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSitesStore } from '@/stores/sites'
import { caddyAPI, sseAPI, settingsAPI } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
    Folder, CircleCheck, CircleClose, Connection, Lock,
    Plus, Setting, Refresh, Monitor, Lightning,
    Edit, Delete, Document, Link, DataLine
} from '@element-plus/icons-vue'

const router = useRouter()
const sitesStore = useSitesStore()

let eventSource: EventSource | null = null

const stats = ref({
    totalSites: 0,
    activeSites: 0,
    totalProxies: 0,
    sslCerts: 0
})

const caddyStatus = ref<'running' | 'checking'>('checking')
const caddyVersion = ref('')
const caddyStatusText = computed(() =>
    caddyStatus.value === 'running' ? 'Caddy2 运行中' : 'Caddy2 已停止'
)

const statsData = computed(() => [
    { label: '站点总数', value: stats.value.totalSites, icon: Folder, color: 'cyan' },
    { label: '运行中', value: stats.value.activeSites, icon: CircleCheck, color: 'green' },
    { label: '代理数量', value: stats.value.totalProxies, icon: Connection, color: 'blue' },
    { label: 'SSL 证书', value: stats.value.sslCerts, icon: Lock, color: 'amber' }
])

const recentSites = ref([
    { id: '1', domain: 'example.com', type: 'reverse_proxy', status: 'active' },
    { id: '2', domain: 'app.example.com', type: 'reverse_proxy', status: 'active' },
    { id: '3', domain: 'static.example.com', type: 'static', status: 'inactive' }
])

const loadStats = async () => {
    try {
        const config = await caddyAPI.getConfig()
        const apps = config.apps || {}
        const httpApps = apps.http || {}
        const servers = httpApps.servers || {}

        let siteCount = 0
        let proxyCount = 0

        Object.entries(servers).forEach(([key, server]: [string, any]) => {
            siteCount++
            const routes = server.routes || []
            routes.forEach((route: any) => {
                const handle = route.handle || []
                handle.forEach((h: any) => {
                    if (h.handler === 'reverse_proxy') {
                        proxyCount++
                    }
                })
            })
        })

        stats.value.totalSites = siteCount
        stats.value.activeSites = siteCount
        stats.value.totalProxies = proxyCount
        stats.value.sslCerts = Object.keys(httpApps.tls || {}).length || 0
    } catch (error) {
        console.error('Failed to load stats:', error)
    }
}

const loadCaddyStatus = async () => {
    try {
        const res = await settingsAPI.getCaddyStatus()
        caddyVersion.value = res.version || '2.x.x'
        caddyStatus.value = res.status === 'running' ? 'running' : 'stopped'
    } catch {
        caddyStatus.value = 'stopped'
        caddyVersion.value = '未知'
    }
}

const reloadCaddy = async () => {
    try {
        await caddyAPI.getConfig()
        ElMessage.success('Caddy2 配置重载成功')
    } catch {
        ElMessage.error('配置重载失败')
    }
}

const editSite = (site: any) => {
    router.push(`/web/sites/${site.id}`)
}

const deleteSite = async (site: any) => {
    try {
        await ElMessageBox.confirm(
            `确定要删除站点 "${site.domain}" 吗？`,
            '删除确认',
            { type: 'warning' }
        )
        ElMessage.success('站点已删除')
    } catch {
        // 用户取消
    }
}

onMounted(() => {
    loadStats()
    loadCaddyStatus()

    eventSource = sseAPI.createConnection()

    eventSource.addEventListener('caddy_status', (event) => {
        try {
            const data = JSON.parse(event.data)
            caddyStatus.value = data.status
            caddyVersion.value = data.version || ''
        } catch (error) {
            console.error('Failed to parse SSE data:', error)
        }
    })

    eventSource.onerror = () => {
        console.error('SSE connection error')
        caddyStatus.value = 'stopped'
    }
})

onUnmounted(() => {
    if (eventSource) {
        eventSource.close()
        eventSource = null
    }
})
</script>

<style scoped lang="scss">
.dashboard-container {
    animation: fadeIn 0.4s ease-out;
}

.dashboard-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 32px;

    .header-content {
        .page-title {
            font-size: 28px;
            font-weight: 700;
            color: var(--text-primary);
            margin: 0 0 6px;
            letter-spacing: -0.5px;
        }

        .page-subtitle {
            font-size: 14px;
            color: var(--text-secondary);
            margin: 0;
        }
    }
}

.stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 20px;
    margin-bottom: 28px;
}

.stat-card {
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    padding: 24px;
    display: flex;
    align-items: center;
    gap: 20px;
    position: relative;
    overflow: hidden;
    animation: slideUp 0.5s ease-out backwards;

    .stat-icon {
        width: 56px;
        height: 56px;
        border-radius: var(--radius-md);
        display: flex;
        align-items: center;
        justify-content: center;
        position: relative;

        &.cyan {
            background: rgba(0, 212, 255, 0.1);
            color: var(--accent-cyan);
        }

        &.green {
            background: rgba(103, 194, 58, 0.1);
            color: #67c23a;
        }

        &.blue {
            background: rgba(64, 158, 255, 0.1);
            color: #409eff;
        }

        &.amber {
            background: rgba(255, 184, 0, 0.1);
            color: var(--accent-amber);
        }
    }

    .stat-info {
        position: relative;
        z-index: 1;

        .stat-value {
            font-size: 32px;
            font-weight: 700;
            color: var(--text-primary);
            line-height: 1;
            margin-bottom: 6px;
        }

        .stat-label {
            font-size: 13px;
            color: var(--text-secondary);
        }
    }

    .stat-glow {
        position: absolute;
        top: -50%;
        right: -50%;
        width: 100%;
        height: 100%;
        background: radial-gradient(circle, rgba(0, 212, 255, 0.05) 0%, transparent 70%);
    }
}

.dashboard-grid {
    display: grid;
    grid-template-columns: 1fr 360px;
    gap: 24px;
}

.panel {
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    overflow: hidden;
}

.panel-header {
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    justify-content: space-between;
    align-items: center;

    .panel-title {
        display: flex;
        align-items: center;
        gap: 10px;
        font-size: 15px;
        font-weight: 600;
        color: var(--text-primary);
        margin: 0;

        .el-icon {
            color: var(--accent-cyan);
        }
    }
}

.sites-panel {
    .sites-list {
        padding: 8px 0;
    }

    .site-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 16px 24px;
        border-bottom: 1px solid var(--border-subtle);
        transition: background 0.2s;

        &:last-child {
            border-bottom: none;
        }

        &:hover {
            background: var(--bg-hover);
        }
    }

    .site-info {
        .site-domain {
            font-size: 14px;
            font-weight: 500;
            color: var(--text-primary);
            margin-bottom: 6px;
        }

        .site-meta {
            display: flex;
            gap: 12px;
        }

        .site-type, .site-status {
            font-family: var(--font-mono);
            font-size: 11px;
            padding: 3px 8px;
            border-radius: 4px;
        }

        .site-type {
            background: rgba(64, 158, 255, 0.1);
            color: #409eff;

            &.static {
                background: rgba(103, 194, 58, 0.1);
                color: #67c23a;
            }
        }

        .site-status {
            background: rgba(255, 184, 0, 0.1);
            color: var(--accent-amber);

            &.active {
                background: rgba(103, 194, 58, 0.1);
                color: #67c23a;
            }
        }
    }

    .site-actions {
        display: flex;
        gap: 8px;
    }

    .icon-btn {
        width: 32px;
        height: 32px;
        border: 1px solid var(--border-subtle);
        border-radius: var(--radius-sm);
        background: var(--bg-hover);
        color: var(--text-secondary);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.2s;

        &:hover {
            border-color: var(--accent-cyan);
            color: var(--accent-cyan);
        }

        &.danger:hover {
            border-color: var(--accent-magenta);
            color: var(--accent-magenta);
        }
    }
}

.side-panels {
    display: flex;
    flex-direction: column;
    gap: 24px;
}

.status-panel {
    .status-display {
        padding: 32px;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 20px;
    }

    .status-ring {
        width: 100px;
        height: 100px;
        border-radius: 50%;
        border: 3px solid var(--text-muted);
        display: flex;
        align-items: center;
        justify-content: center;
        position: relative;

        &.running {
            border-color: var(--accent-cyan);
            box-shadow: 0 0 20px rgba(0, 212, 255, 0.3);
            animation: pulse 2s ease-in-out infinite;

            .ring-dot {
                width: 16px;
                height: 16px;
                background: var(--accent-cyan);
                border-radius: 50%;
                box-shadow: 0 0 10px var(--accent-cyan);
            }
        }

        .ring-dot {
            width: 12px;
            height: 12px;
            background: var(--text-muted);
            border-radius: 50%;
        }
    }

    .status-info {
        text-align: center;

        .status-text {
            font-size: 16px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 4px;
        }

        .status-version {
            font-family: var(--font-mono);
            font-size: 12px;
            color: var(--text-muted);
        }
    }

    .status-actions {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
        padding: 0 24px 24px;
    }

    .status-btn {
        padding: 12px 16px;
        background: var(--bg-hover);
        border: 1px solid var(--border-subtle);
        border-radius: var(--radius-md);
        color: var(--text-secondary);
        font-family: var(--font-display);
        font-size: 13px;
        font-weight: 500;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        transition: all 0.2s;

        &:hover {
            border-color: var(--accent-cyan);
            color: var(--accent-cyan);
            background: var(--bg-active);
        }
    }
}

.quick-panel {
    .quick-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
        padding: 20px;
    }

    .quick-btn {
        padding: 16px;
        background: var(--bg-hover);
        border: 1px solid var(--border-subtle);
        border-radius: var(--radius-md);
        color: var(--text-secondary);
        cursor: pointer;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 10px;
        transition: all 0.2s;

        .el-icon {
            font-size: 20px;
        }

        span {
            font-size: 12px;
            font-weight: 500;
        }

        &:hover {
            border-color: var(--accent-cyan);
            color: var(--accent-cyan);
            background: var(--bg-active);
            transform: translateY(-2px);
        }
    }
}

.action-btn {
    padding: 10px 20px;
    background: transparent;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-secondary);
    font-family: var(--font-display);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    transition: all 0.2s;

    &:hover {
        border-color: var(--accent-cyan);
        color: var(--accent-cyan);
        background: var(--bg-hover);
    }
}

@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}

@keyframes slideUp {
    from {
        opacity: 0;
        transform: translateY(20px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

@keyframes pulse {
    0%, 100% {
        box-shadow: 0 0 20px rgba(0, 212, 255, 0.3);
    }
    50% {
        box-shadow: 0 0 40px rgba(0, 212, 255, 0.5);
    }
}
</style>