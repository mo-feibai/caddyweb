<template>
    <div class="logs-container">
        <el-card>
            <template #header>
                <div class="card-header">
                    <span>访问日志</span>
                    <div class="header-actions">
                        <el-select v-model="logLevel" placeholder="日志级别" style="width: 120px; margin-right: 10px"
                            @change="loadLogs">
                            <el-option label="全部" value="all" />
                            <el-option label="INFO" value="info" />
                            <el-option label="WARN" value="warn" />
                            <el-option label="ERROR" value="error" />
                        </el-select>
                        <el-button @click="loadLogs" :loading="loading">
                            <el-icon>
                                <i-ep-refresh />
                            </el-icon> 刷新
                        </el-button>
                        <el-button @click="clearLogs">
                            <el-icon>
                                <i-ep-delete />
                            </el-icon> 清空
                        </el-button>
                    </div>
                </div>
            </template>

            <div class="log-stats">
                <el-row :gutter="20">
                    <el-col :span="6">
                        <div class="stat-item">
                            <span class="stat-label">总请求数</span>
                            <span class="stat-value">{{ stats.total }}</span>
                        </div>
                    </el-col>
                    <el-col :span="6">
                        <div class="stat-item">
                            <span class="stat-label">成功 (2xx)</span>
                            <span class="stat-value success">{{ stats.success }}</span>
                        </div>
                    </el-col>
                    <el-col :span="6">
                        <div class="stat-item">
                            <span class="stat-label">重定向 (3xx)</span>
                            <span class="stat-value warning">{{ stats.redirect }}</span>
                        </div>
                    </el-col>
                    <el-col :span="6">
                        <div class="stat-item">
                            <span class="stat-label">错误 (4xx/5xx)</span>
                            <span class="stat-value danger">{{ stats.error }}</span>
                        </div>
                    </el-col>
                </el-row>
            </div>

            <el-input v-model="searchKeyword" placeholder="搜索日志内容..." style="margin: 15px 0" clearable>
                <template #prefix>
                    <el-icon>
                        <i-ep-search />
                    </el-icon>
                </template>
            </el-input>

            <el-table :data="filteredLogs" stripe style="width: 100%" v-loading="loading" max-height="500">
                <el-table-column prop="timestamp" label="时间" width="180">
                    <template #default="{ row }">
                        {{ formatTime(row.timestamp) }}
                    </template>
                </el-table-column>
                <el-table-column prop="level" label="级别" width="80">
                    <template #default="{ row }">
                        <el-tag :type="getLevelType(row.level)" size="small">
                            {{ row.level }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="host" label="主机" width="150" />
                <el-table-column prop="method" label="方法" width="80">
                    <template #default="{ row }">
                        <el-tag :type="getMethodType(row.method)" size="small">
                            {{ row.method }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
                <el-table-column prop="status" label="状态" width="80">
                    <template #default="{ row }">
                        <span :class="getStatusClass(row.status)">
                            {{ row.status }}
                        </span>
                    </template>
                </el-table-column>
                <el-table-column prop="duration" label="耗时" width="100">
                    <template #default="{ row }">
                        {{ row.duration }}ms
                    </template>
                </el-table-column>
                <el-table-column prop="size" label="大小" width="100">
                    <template #default="{ row }">
                        {{ formatSize(row.size) }}
                    </template>
                </el-table-column>
            </el-table>

            <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :total="filteredLogs.length"
                :page-sizes="[50, 100, 200, 500]" layout="total, sizes, prev, pager, next"
                style="margin-top: 20px; justify-content: flex-end" />
        </el-card>
    </div>
</template>

<script setup lang="ts">
import { logsAPI } from '@/api'

interface HttpLogEntry {
    timestamp: string
    level: 'INFO' | 'WARN' | 'ERROR'
    host: string
    method: string
    path: string
    status: number
    duration: number
    size: number
    message?: string
}

const loading = ref(false)
const logLevel = ref('all')
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(100)
const logs = shallowRef<HttpLogEntry[]>([])

const stats = computed(() => {
    const total = logs.value.length
    const success = logs.value.filter(l => l.status >= 200 && l.status < 300).length
    const redirect = logs.value.filter(l => l.status >= 300 && l.status < 400).length
    const error = logs.value.filter(l => l.status >= 400).length
    return { total, success, redirect, error }
})

const filteredLogs = computed(() => {
    let result = logs.value

    if (logLevel.value !== 'all') {
        result = result.filter(log => log.level.toLowerCase() === logLevel.value)
    }

    if (searchKeyword.value) {
        const keyword = searchKeyword.value.toLowerCase()
        result = result.filter(log =>
            log.path.toLowerCase().includes(keyword) ||
            log.host.toLowerCase().includes(keyword) ||
            log.message?.toLowerCase().includes(keyword)
        )
    }

    return result
})

const formatTime = (timestamp: string) => {
    const date = new Date(timestamp)
    return date.toLocaleString('zh-CN')
}

const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

const getLevelType = (level: string) => {
    switch (level) {
        case 'ERROR': return 'danger'
        case 'WARN': return 'warning'
        default: return 'info'
    }
}

const getMethodType = (method: string) => {
    switch (method) {
        case 'GET': return 'primary'
        case 'POST': return 'success'
        case 'PUT': return 'warning'
        case 'DELETE': return 'danger'
        default: return 'info'
    }
}

const getStatusClass = (status: number) => {
    if (status >= 200 && status < 300) return 'status-success'
    if (status >= 300 && status < 400) return 'status-redirect'
    if (status >= 400 && status < 500) return 'status-client-error'
    return 'status-server-error'
}

const loadLogs = async () => {
    loading.value = true
    try {
        const data = await logsAPI.get({ limit: pageSize.value })
        logs.value = (data as unknown as HttpLogEntry[]) || mockLogs()
    } catch {
        ElMessage.error('加载日志失败')
        logs.value = mockLogs()
    } finally {
        loading.value = false
    }
}

const clearLogs = () => {
    logs.value = []
    ElMessage.success('日志已清空')
}

const mockLogs = (): HttpLogEntry[] => {
    const methods = ['GET', 'POST', 'PUT', 'DELETE']
    const levels: Array<'INFO' | 'WARN' | 'ERROR'> = ['INFO', 'INFO', 'INFO', 'WARN', 'ERROR']
    const paths = ['/', '/api/users', '/api/products', '/static/js/app.js', '/api/orders']
    const hosts = ['192.168.1.100', '10.0.0.50', '172.16.0.25']

    return Array.from({ length: 50 }, (_, i) => ({
        timestamp: new Date(Date.now() - i * 60000).toISOString(),
        level: levels[Math.floor(Math.random() * levels.length)],
        host: hosts[Math.floor(Math.random() * hosts.length)],
        method: methods[Math.floor(Math.random() * methods.length)],
        path: paths[Math.floor(Math.random() * paths.length)],
        status: [200, 200, 200, 301, 302, 400, 404, 500][Math.floor(Math.random() * 8)],
        duration: Math.floor(Math.random() * 500) + 10,
        size: Math.floor(Math.random() * 10000) + 100
    }))
}

onMounted(() => {
    loadLogs()
})
</script>

<style scoped lang="scss">
.logs-container {
    padding: 24px;
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.header-actions {
    display: flex;
    align-items: center;
}
</style>
