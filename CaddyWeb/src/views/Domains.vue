<template>
    <div class="domains-container">
        <el-card>
            <template #header>
                <div class="card-header">
                    <span>域名管理</span>
                    <el-button type="primary" @click="showAddDialog">
                        <el-icon>
                            <i-ep-plus />
                        </el-icon> 添加域名
                    </el-button>
                </div>
            </template>

            <el-table :data="domains" stripe style="width: 100%" v-loading="loading" row-key="name">
                <el-table-column prop="id" label="ID" min-width="200" align="center">
                    <template #default="{ row }">
                        <el-text size="small" type="info">{{ row.id || '-' }}</el-text>
                    </template>
                </el-table-column>
                <el-table-column prop="name" label="基础域名" min-width="180" align="center">
                    <template #default="{ row }">
                        <el-link type="primary" @click="viewDomainDetail(row)">{{ row.name }}</el-link>
                    </template>
                </el-table-column>
                <el-table-column prop="wildcard" label="泛域名" min-width="180" align="center">
                    <template #default="{ row }">
                        {{ formatWildcard(row.wildcard) }}
                    </template>
                </el-table-column>
                <el-table-column prop="server_id" label="服务器" width="200" align="center">
                    <template #default="{ row }">
                        <div class="server-info">
                            <span class="server-id">{{ row.server_id || '-' }}</span>
                            <el-tag size="small" type="info">{{ row.listen?.[0]?.replace(/^:/, '') || '-' }}</el-tag>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column label="TLS" width="80" align="center">
                    <template #default="{ row }">
                        <el-icon v-if="row.tls_enabled" color="#67C23A">
                            <i-ep-lock />
                        </el-icon>
                        <el-icon v-else color="#909399">
                            <i-ep-unlock />
                        </el-icon>
                    </template>
                </el-table-column>
                <el-table-column label="操作" width="160" align="center">
                    <template #default="{ row }">
                        <el-button type="primary" link size="small" @click="viewDomainSites(row)">子站点</el-button>
                        <el-button type="danger" link size="small" @click="deleteDomain(row)">删除</el-button>
                    </template>
                </el-table-column>
            </el-table>
        </el-card>

        <!-- 添加/编辑域名弹窗 -->
        <el-dialog v-model="dialogVisible" title="添加域名" width="500px">
            <el-form ref="formRef" :model="domainForm" :rules="rules" label-width="120px">
                <el-form-item label="域名" prop="name">
                    <el-input v-model="domainForm.name" placeholder="example.com" />
                    <div class="form-tip">基础域名，将自动匹配 *.domain</div>
                </el-form-item>

                <el-form-item label="服务器">
                    <el-select v-model="domainForm.server_id" placeholder="选择服务器">
                        <el-option v-for="s in servers" :key="s.id" :label="s.id + ' (' + s.listen.join(', ') + ')'"
                            :value="s.id" />
                    </el-select>
                </el-form-item>

                <el-form-item label="启用 TLS">
                    <el-switch v-model="domainForm.tls_enabled" />
                    <div class="form-tip">启用后自动申请 Let's Encrypt 证书</div>
                </el-form-item>

                <el-form-item label="标识 ID">
                    <el-input v-model="domainForm.id" placeholder="留空则自动生成 UUID" />
                    <div class="form-tip">用于 API 引用的可选标识符</div>
                </el-form-item>
            </el-form>

            <template #footer>
                <el-button @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" :loading="saving" @click="submitDomain">
                    创建
                </el-button>
            </template>

        </el-dialog>

        <!-- 域名详情弹窗 -->
        <el-dialog v-model="detailVisible" :title="currentDomain?.name + ' 详情'" width="600px">
            <div v-if="currentDomain" class="domain-detail">
                <el-descriptions :column="2" border>
                    <el-descriptions-item label="ID">{{ currentDomain?.id || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="基础域名">{{ currentDomain.name }}</el-descriptions-item>
                    <el-descriptions-item label="泛域名">{{ formatWildcard(currentDomain.wildcard)
                    }}</el-descriptions-item>
                    <el-descriptions-item label="服务器">{{ currentDomain.server_id }}</el-descriptions-item>
                    <el-descriptions-item label="TLS">{{ currentDomain.tls_enabled ? '已启用' : '未启用'
                        }}</el-descriptions-item>
                </el-descriptions>
            </div>
        </el-dialog>

        <!-- 子站点列表弹窗 -->
        <el-dialog v-model="sitesVisible" :title="currentDomain?.name + ' - 子站点'" width="800px">
            <div class="sites-header">
                <el-button type="primary" size="small" @click="showAddSiteDialog">
                    <el-icon>
                        <i-ep-plus />
                    </el-icon> 添加子站点
                </el-button>
            </div>

            <el-table :data="domainSites" stripe style="width: 100%; margin-top: 15px">
                <el-table-column prop="id" label="ID" min-width="180" align="center">
                    <template #default="{ row }">
                        <el-text size="small" type="info">{{ row.id || '-' }}</el-text>
                    </template>
                </el-table-column>
                <el-table-column prop="name" label="名称" width="120" align="center" />
                <el-table-column prop="host" label="完整域名" min-width="180" align="center" />
                <el-table-column prop="type" label="类型" width="100" align="center">
                    <template #default="{ row }">
                        <el-tag :type="row.type === 'reverse_proxy' ? 'success' : 'info'" size="small">
                            {{ row.type === 'reverse_proxy' ? '反向代理' : '静态站点' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column label="目标" min-width="150" align="center">
                    <template #default="{ row }">
                        {{ row.type === 'static' ? row.root : row.upstream }}
                    </template>
                </el-table-column>
                <el-table-column label="操作" width="160" align="center">
                    <template #default="{ row }">
                        <el-button type="primary" link size="small"
                            @click="$router.push('/web/sites/' + row.id)">查看</el-button>
                        <el-button type="primary" link size="small" @click="editSite(row)">编辑</el-button>
                        <el-button type="danger" link size="small" @click="deleteSite(row)">删除</el-button>
                    </template>
                </el-table-column>
            </el-table>
        </el-dialog>

        <!-- 添加/编辑子站点弹窗 -->
        <el-dialog class="site-dialog" v-model="siteDialogVisible" :title="isEditingSite ? '编辑子站点' : '添加子站点'"
            width="900px">
            <div class="site-dialog-content">
                <el-form ref="siteFormRef" :model="siteForm" :rules="siteRules" label-width="120px">
                    <el-form-item label="标识 ID">
                        <el-input v-model="siteForm.id" :disabled="isEditingSite" :placeholder="isEditingSite ? '' : '留空则自动生成 UUID'" />
                        <div class="form-tip">{{ isEditingSite ? '站点标识不可修改' : '用于 API 引用的可选标识符' }}</div>
                    </el-form-item>

                    <el-form-item label="子站点名称" prop="name">
                        <el-input v-model="siteForm.name" placeholder="first" />
                        <div class="form-tip">完整域名: {{ siteForm.name }}.{{ currentDomain?.name }}</div>
                    </el-form-item>

                    <el-form-item label="站点类型" prop="type">
                        <el-radio-group v-model="siteForm.type">
                            <el-radio value="static">静态站点</el-radio>
                            <el-radio value="reverse_proxy">反向代理</el-radio>
                        </el-radio-group>
                    </el-form-item>

                    <el-form-item label="站点目录" prop="root" v-if="siteForm.type === 'static'">
                        <el-input v-model="siteForm.root" placeholder="/var/www/html" />
                    </el-form-item>
                    <el-form-item label="默认文档" v-if="siteForm.type === 'static'">
                        <el-input v-model="siteForm.indexNames" placeholder="index.html" />
                    </el-form-item>
                    <el-form-item label="上游服务器" prop="upstream" v-if="siteForm.type === 'reverse_proxy'">
                        <el-input v-model="siteForm.upstream" placeholder="127.0.0.1:8080" />
                    </el-form-item>
                    <el-form-item label="健康检查" v-if="siteForm.type === 'reverse_proxy'">
                        <el-switch v-model="siteForm.health_check" />
                    </el-form-item>
                </el-form>
            </div>

            <template #footer>
                <el-button @click="siteDialogVisible = false">取消</el-button>
                <el-button type="primary" :loading="siteSaving" @click="submitSite">
                    {{ isEditingSite ? '更新' : '创建' }}
                </el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import type { FormInstance } from 'element-plus'
import type { Domain, Site, CreateSiteRequest, UpdateSiteRequest } from '@/types'
import { domainAPI, settingsAPI, siteAPI } from '@/api'
import type { ServerInfo, DomainFormData, SiteFormData } from '@/types'

const loading = ref(false)
const saving = ref(false)
const siteSaving = ref(false)

const domains = ref<Domain[]>([])
const domainSites = ref<Site[]>([])
const servers = ref<ServerInfo[]>([])

const dialogVisible = ref(false)
const detailVisible = ref(false)
const sitesVisible = ref(false)
const siteDialogVisible = ref(false)

const isEditingSite = ref(false)

const currentDomain = ref<Domain | null>(null)

const formRef = ref<FormInstance>()
const siteFormRef = ref<FormInstance>()

const domainForm = reactive<DomainFormData>({
    name: '',
    server_id: '',
    tls_enabled: true,
    id: ''
})

const siteForm = reactive<SiteFormData>({
    name: '',
    type: 'reverse_proxy',
    id: '',
    upstream: '',
    root: '/var/www/html',
    indexNames: 'index.html',
    health_check: false
})

const rules = {
    name: [
        { required: true, message: '请输入域名', trigger: 'blur' },
        { pattern: /^[\w\-\.]+\.[\w\-\.]+$/, message: '域名格式不正确', trigger: 'blur' }
    ]
}

const formatWildcard = (wildcard: string) => {
    if (!wildcard) return '-'
    if (wildcard.startsWith('*.')) {
        return wildcard
    }
    return '*.' + wildcard
}

const siteRules = {
    name: [
        { required: true, message: '请输入子站点名称', trigger: 'blur' },
        { pattern: /^[a-zA-Z0-9\-]+$/, message: '只能包含字母、数字和连字符', trigger: 'blur' }
    ],
    upstream: [{ required: true, message: '请输入上游服务器', trigger: 'blur' }],
    root: [{ required: true, message: '请输入站点目录', trigger: 'blur' }]
}

const loadDomains = async () => {
    loading.value = true
    try {
        const data = await domainAPI.list()
        domains.value = data || []
    } catch (error) {
        ElMessage.error('加载域名失败')
    } finally {
        loading.value = false
    }
}

const loadServers = async () => {
    try {
        const data = await settingsAPI.getServers()
        servers.value = data || []
    } catch (error) {
        console.error('加载服务器失败:', error)
    }
}

const showAddDialog = async () => {
    domainForm.name = ''
    domainForm.server_id = servers.value[0]?.id || ''
    domainForm.tls_enabled = true
    domainForm.id = ''
    dialogVisible.value = true
    await loadServers()
}

const submitDomain = async () => {
    try {
        await formRef.value?.validate()
        saving.value = true

        const payload: { name: string; server_id: string; tls: boolean; id?: string } = {
            name: domainForm.name,
            server_id: domainForm.server_id,
            tls: domainForm.tls_enabled
        }
        if (domainForm.id) {
            payload.id = domainForm.id
        }

        await domainAPI.create(payload)
        ElMessage.success('域名创建成功')

        dialogVisible.value = false
        loadDomains()
    } catch (error) {
        if (error !== false) {
            ElMessage.error('创建失败')
        }
    } finally {
        saving.value = false
    }
}

const deleteDomain = async (domain: Domain) => {
    try {
        await ElMessageBox.confirm(
            `确定要删除域名 "${domain.name}" 吗？`,
            '删除确认',
            { type: 'warning' }
        )
        await domainAPI.delete(domain.server_id, domain.id)
        ElMessage.success('域名已删除')
        loadDomains()
    } catch (error) {
        if (error !== 'cancel') {
            if ((error as { response?: { data?: { message?: string } } })?.response?.data?.message?.includes('sub-sites')) {
                ElMessageBox.alert('无法删除域名：该域名下存在子站点，请先删除所有子站点', '删除失败', {
                    type: 'warning'
                })
            } else {
                ElMessage.error('删除失败')
            }
        }
    }
}

const viewDomainDetail = (domain: Domain) => {
    currentDomain.value = domain
    detailVisible.value = true
}

const viewDomainSites = async (domain: Domain) => {
    currentDomain.value = domain
    loading.value = true
    sitesVisible.value = true
    try {
        const data = await domainAPI.getSites(domain.id)
        domainSites.value = data || []
    } catch (error) {
        ElMessage.error('加载子站点失败')
    } finally {
        loading.value = false
    }
}

const showAddSiteDialog = () => {
    isEditingSite.value = false
    siteForm.name = ''
    siteForm.id = ''
    siteForm.type = 'reverse_proxy'
    siteForm.upstream = ''
    siteForm.root = '/var/www/html'
    siteForm.indexNames = 'index.html'
    siteForm.health_check = false
    siteDialogVisible.value = true
}

const editSite = (site: Site) => {
    isEditingSite.value = true
    siteForm.name = site.name
    siteForm.id = site.id
    siteForm.type = site.type
    siteForm.upstream = site.upstream || ''
    siteForm.root = site.root || ''
    siteForm.indexNames = site.index_names || ''
    siteForm.health_check = site.health_check
    siteDialogVisible.value = true
}

const submitSite = async () => {
    try {
        await siteFormRef.value?.validate()
        siteSaving.value = true

        if (isEditingSite.value && currentDomain.value) {
            const payload: UpdateSiteRequest = {
                name: siteForm.name,
                type: siteForm.type,
                upstream: siteForm.upstream,
                root: siteForm.root,
                index_names: siteForm.indexNames,
                health_check: siteForm.health_check
            }
            await siteAPI.update(siteForm.id, payload)
            ElMessage.success('子站点更新成功')
        } else if (currentDomain.value) {
            const payload: CreateSiteRequest = {
                name: siteForm.name,
                type: siteForm.type,
                upstream: siteForm.upstream,
                root: siteForm.root,
                index_names: siteForm.indexNames,
                health_check: siteForm.health_check,
                domain: currentDomain.value.name
            }
            await siteAPI.create(currentDomain.value.id, payload)
            ElMessage.success('子站点创建成功')
        }

        siteDialogVisible.value = false
        viewDomainSites(currentDomain.value!)
    } catch (error) {
        if (error !== false) {
            ElMessage.error((isEditingSite.value ? '更新' : '创建') + '失败')
        }
    } finally {
        siteSaving.value = false
    }
}

const deleteSite = async (site: Site) => {
    if (!site.id) return

    try {
        await ElMessageBox.confirm(
            `确定要删除子站点 "${site.name}" 吗？`,
            '删除确认',
            { type: 'warning' }
        )
        await siteAPI.delete(currentDomain.value!.id, site.id)
        ElMessage.success('子站点已删除')
        viewDomainSites(currentDomain.value!)
    } catch (error) {
        if (error !== 'cancel') {
            ElMessage.error('删除失败')
        }
    }
}

onMounted(() => {
    loadDomains()
})
</script>

<style scoped lang="scss">
.domains-container {
    padding: 24px;
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.form-tip {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 8px;
}

.sites-header {
    display: flex;
    justify-content: flex-start;
}

.server-info {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;

    .server-id {
        font-family: var(--font-mono);
        font-weight: 500;
    }
}
</style>
