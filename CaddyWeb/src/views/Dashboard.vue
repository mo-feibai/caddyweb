<template>
    <div class="dashboard-container">
        <el-row :gutter="20" class="stats-row">
            <el-col :span="6">
                <el-card shadow="hover" class="stat-card">
                    <div class="stat-icon sites"><el-icon>
                            <Folder />
                        </el-icon></div>
                    <div class="stat-content">
                        <div class="stat-value">{{ stats.totalSites }}</div>
                        <div class="stat-label">站点总数</div>
                    </div>
                </el-card>
            </el-col>
            <el-col :span="6">
                <el-card shadow="hover" class="stat-card">
                    <div class="stat-icon active"><el-icon>
                            <CircleCheck />
                        </el-icon></div>
                    <div class="stat-content">
                        <div class="stat-value">{{ stats.activeSites }}</div>
                        <div class="stat-label">运行中</div>
                    </div>
                </el-card>
            </el-col>
            <el-col :span="6">
                <el-card shadow="hover" class="stat-card">
                    <div class="stat-icon proxies"><el-icon>
                            <Connection />
                        </el-icon></div>
                    <div class="stat-content">
                        <div class="stat-value">{{ stats.totalProxies }}</div>
                        <div class="stat-label">代理数量</div>
                    </div>
                </el-card>
            </el-col>
            <el-col :span="6">
                <el-card shadow="hover" class="stat-card">
                    <div class="stat-icon certs"><el-icon>
                            <Lock />
                        </el-icon></div>
                    <div class="stat-content">
                        <div class="stat-value">{{ stats.sslCerts }}</div>
                        <div class="stat-label">SSL 证书</div>
                    </div>
                </el-card>
            </el-col>
        </el-row>

        <el-row :gutter="20">
            <el-col :span="16">
                <el-card class="main-card">
                    <template #header>
                        <div class="card-header">
                            <span>最近站点</span>
                            <el-button type="primary" size="small" @click="$router.push('/web/sites/add')">
                                <el-icon>
                                    <Plus />
                                </el-icon> 添加站点
                            </el-button>
                        </div>
                    </template>
                    <el-table :data="recentSites" stripe style="width: 100%">
                        <el-table-column prop="domain" label="域名" min-width="180" />
                        <el-table-column prop="type" label="类型" width="120">
                            <template #default="{ row }">
                                <el-tag :type="row.type === 'reverse_proxy' ? 'success' : 'info'">
                                    {{ row.type === 'reverse_proxy' ? '反向代理' : '静态站点' }}
                                </el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column prop="status" label="状态" width="100">
                            <template #default="{ row }">
                                <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
                                    {{ row.status === 'active' ? '运行中' : '已停止' }}
                                </el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column label="操作" width="150" align="center">
                            <template #default="{ row }">
                                <el-button type="primary" link size="small" @click="editSite(row)">编辑</el-button>
                                <el-button type="danger" link size="small" @click="deleteSite(row)">删除</el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                </el-card>
            </el-col>

            <el-col :span="8">
                <el-card class="main-card">
                    <template #header>
                        <span>Caddy2 状态</span>
                    </template>
                    <div class="caddy-status">
                        <div class="status-indicator" :class="caddyStatus">
                            <el-icon :size="48">
                                <CircleCheck v-if="caddyStatus === 'running'" />
                                <CircleClose v-else />
                            </el-icon>
                        </div>
                        <div class="status-info">
                            <p class="status-text">{{ caddyStatusText }}</p>
                            <p class="status-version">v{{ caddyVersion }}</p>
                        </div>
                        <div class="status-actions">
                            <el-button size="small" @click="reloadCaddy">重载配置</el-button>
                            <el-button size="small" type="primary"
                                @click="$router.push('/web/logs')">查看日志</el-button>
                        </div>
                    </div>
                </el-card>

                <el-card class="main-card mt-20">
                    <template #header>
                        <span>快捷操作</span>
                    </template>
                    <div class="quick-actions">
                        <el-button class="action-btn" @click="$router.push('/web/tls')">
                            <el-icon>
                                <Lock />
                            </el-icon>
                            <span>SSL 证书</span>
                        </el-button>
                        <el-button class="action-btn" @click="$router.push('/web/settings')">
                            <el-icon>
                                <Setting />
                            </el-icon>
                            <span>系统设置</span>
                        </el-button>
                    </div>
                </el-card>
            </el-col>
        </el-row>
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
    Plus, Setting
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
    } catch (error) {
        caddyStatus.value = 'stopped'
        caddyVersion.value = '未知'
    }
}

const reloadCaddy = async () => {
    try {
        await caddyAPI.getConfig()
        ElMessage.success('Caddy2 配置重载成功')
    } catch (error) {
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

    // 建立 SSE 连接
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
    padding: 20px;
}

.stats-row {
    margin-bottom: 20px;
}

.stat-card {
    display: flex;
    align-items: center;
    padding: 20px;

    .stat-icon {
        width: 60px;
        height: 60px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-right: 20px;
        font-size: 28px;
        color: #fff;

        &.sites {
            background: linear-gradient(135deg, #667eea, #764ba2);
        }

        &.active {
            background: linear-gradient(135deg, #67c23a, #85ce61);
        }

        &.proxies {
            background: linear-gradient(135deg, #409eff, #66b1ff);
        }

        &.certs {
            background: linear-gradient(135deg, #e6a23c, #ebb563);
        }
    }

    .stat-content {
        .stat-value {
            font-size: 28px;
            font-weight: bold;
            color: #303133;
        }

        .stat-label {
            font-size: 14px;
            color: #909399;
        }
    }
}

.main-card {
    .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
}

.mt-20 {
    margin-top: 20px;
}

.caddy-status {
    text-align: center;
    padding: 20px 0;

    .status-indicator {
        margin-bottom: 15px;
        color: #67c23a;

        &.stopped {
            color: #f56c6c;
        }
    }

    .status-info {
        .status-text {
            font-size: 18px;
            color: #303133;
            margin: 0 0 5px;
        }

        .status-version {
            color: #909399;
            font-size: 14px;
            margin: 0;
        }
    }

    .status-actions {
        margin-top: 20px;
        display: flex;
        gap: 10px;
        justify-content: center;
    }
}

.quick-actions {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 15px;

    .action-btn {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 20px 10px;
        height: auto;

        .el-icon {
            font-size: 24px;
            margin-bottom: 8px;
        }

        span {
            font-size: 13px;
        }
    }
}
</style>
