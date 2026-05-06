<template>
    <div class="sites-container">
        <el-card>
            <template #header>
                <div class="card-header">
                    <div class="header-left">
                        <span>子站点管理</span>
                        <el-select v-model="selectedDomain" placeholder="选择域名" style="width: 200px; margin-left: 20px" clearable>
                            <el-option v-for="d in domains" :key="d.name" :label="d.name" :value="d.name" />
                        </el-select>
                        <el-radio-group v-model="selectedType" style="margin-left: 20px">
                            <el-radio-button value="">全部</el-radio-button>
                            <el-radio-button value="static">静态站点</el-radio-button>
                            <el-radio-button value="reverse_proxy">反向代理</el-radio-button>
                        </el-radio-group>
                    </div>
                    <div class="header-right">
                        <el-button type="primary" @click="showAddDialog">
                            <el-icon><i-ep-plus /></el-icon> 添加子站点
                        </el-button>
                        <el-button type="danger" @click="batchDeleteSites" :disabled="selectedSites.length === 0">
                            <el-icon><i-ep-delete /></el-icon> 批量删除
                        </el-button>
                    </div>
                </div>
            </template>

            <el-table :data="filteredSites" stripe style="width: 100%" v-loading="loading" row-key="host" @selection-change="handleSelectionChange">
                <el-table-column type="selection" width="50" align="center" />
                <el-table-column prop="name" label="站点名称" width="150" align="center" />
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
                <el-table-column prop="health_check" label="健康检查" width="100" align="center">
                    <template #default="{ row }">
                        <el-icon v-if="row.health_check" color="#67C23A">
                            <i-ep-check />
                        </el-icon>
                        <el-icon v-else color="#909399">
                            <i-ep-close />
                        </el-icon>
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

        <!-- 添加/编辑子站点弹窗 -->
        <el-dialog v-model="dialogVisible" :title="isEditing ? '编辑子站点' : '添加子站点'" width="500px">
            <el-form ref="formRef" :model="siteForm" :rules="rules" label-width="120px">
                <el-form-item label="所属域名" prop="domain">
                    <el-select v-model="siteForm.domain" placeholder="选择域名" :disabled="isEditing" style="width: 100%">
                        <el-option v-for="d in domains" :key="d.name" :label="d.name" :value="d.name" />
                    </el-select>
                </el-form-item>

                <el-form-item label="站点名称" prop="name">
                    <el-input v-model="siteForm.name" :placeholder="isEditing ? '' : 'first'" />
                    <div class="form-tip" v-if="!isEditing">完整域名: {{ siteForm.name }}.{{ siteForm.domain }}</div>
                </el-form-item>

                <el-form-item label="站点类型" prop="type">
                    <el-radio-group v-model="siteForm.type">
                        <el-radio value="static">静态站点</el-radio>
                        <el-radio value="reverse_proxy">反向代理</el-radio>
                    </el-radio-group>
                </el-form-item>

                <template v-if="siteForm.type === 'static'">
                    <el-form-item label="站点目录" prop="root">
                        <el-input v-model="siteForm.root" placeholder="/var/www/html" />
                    </el-form-item>
                    <el-form-item label="默认文档">
                        <el-input v-model="siteForm.indexNames" placeholder="index.html" />
                    </el-form-item>
                </template>

                <template v-else>
                    <el-form-item label="上游服务器" prop="upstream">
                        <el-input v-model="siteForm.upstream" placeholder="127.0.0.1:8080" />
                    </el-form-item>
                    <el-form-item label="健康检查">
                        <el-switch v-model="siteForm.health_check" />
                    </el-form-item>
                </template>
            </el-form>

            <template #footer>
                <el-button @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" :loading="saving" @click="submitSite">{{ isEditing ? '更新' : '创建' }}</el-button>
            </template>
        </el-dialog>

        <!-- 批量删除二次确认弹窗 -->
        <el-dialog v-model="confirmDialogVisible" title="二次确认" width="450px" :close-on-click-modal="false">
            <div class="confirm-content">
                <p class="confirm-text">请输入 <span class="delete-code" @click="copyDeleteCode">DELETE</span> 确认删除 {{ selectedSites.length }} 个子站点</p>
                <el-input v-model="confirmInput" placeholder="请输入 DELETE" clearable>
                    <template #suffix>
                        <el-icon v-if="confirmInput === 'DELETE'" color="#67C23A"><i-ep-check /></el-icon>
                        <el-icon v-else-if="confirmInput && confirmInput !== 'DELETE'" color="#F56C6C"><i-ep-close /></el-icon>
                    </template>
                </el-input>
            </div>
            <template #footer>
                <el-button @click="confirmDialogVisible = false">取消</el-button>
                <el-button type="danger" :disabled="confirmInput !== 'DELETE'" :loading="deleting" @click="executeBatchDelete">确认删除</el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import type { FormInstance } from 'element-plus'
import type { Site, Domain } from '@/types'
import { domainAPI, siteAPI } from '@/api'
import { useClipboard } from '@vueuse/core'
import type { SiteFormData, SiteType } from '@/types'

const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const selectedSites = shallowRef<Site[]>([])
const confirmDialogVisible = ref(false)
const confirmInput = ref('')
const { copy } = useClipboard({ legacy: true })

const domains = shallowRef<Domain[]>([])
const sites = shallowRef<Site[]>([])
const selectedDomain = ref('')
const selectedType = ref('')

const dialogVisible = ref(false)
const isEditing = ref(false)
const formRef = ref<FormInstance>()

const siteForm = reactive<SiteFormData>({
    domain: '',
    id: '',
    name: '',
    type: 'reverse_proxy',
    upstream: '',
    root: '/var/www/html',
    indexNames: 'index.html',
    health_check: false
})

const rules = {
    domain: [{ required: true, message: '请选择域名', trigger: 'change' }],
    name: [
        { required: true, message: '请输入站点名称', trigger: 'blur' },
        { pattern: /^[a-zA-Z0-9\-]+$/, message: '只能包含字母、数字和连字符', trigger: 'blur' }
    ],
    upstream: [{ required: true, message: '请输入上游服务器', trigger: 'blur' }],
    root: [{ required: true, message: '请输入站点目录', trigger: 'blur' }]
}

const filteredSites = computed(() => {
    let result = sites.value
    if (selectedDomain.value) {
        result = result.filter(s => s.host.endsWith('.' + selectedDomain.value))
    }
    if (selectedType.value) {
        result = result.filter(s => s.type === selectedType.value)
    }
    return result
})

const loadDomains = async () => {
    try {
        const data = await domainAPI.list()
        domains.value = data || []
    } catch (error: any) {
        ElMessage.error('加载域名失败: ' + error.message)
    }
}

const loadAllSites = async () => {
    loading.value = true
    try {
        const data = await siteAPI.list()
        sites.value = data || []
    } catch (error: any) {
        ElMessage.error('加载子站点失败: ' + error.message)
    } finally {
        loading.value = false
    }
}

const showAddDialog = () => {
    isEditing.value = false
    siteForm.domain = selectedDomain.value || ''
    siteForm.id = ''
    siteForm.name = ''
    siteForm.type = 'reverse_proxy'
    siteForm.upstream = ''
    siteForm.root = '/var/www/html'
    siteForm.indexNames = 'index.html'
    siteForm.health_check = false
    dialogVisible.value = true
}

const editSite = (site: Site) => {
    isEditing.value = true
    const hostParts = site.host.split('.')
    const domainName = hostParts.slice(1).join('.')
    siteForm.domain = domainName
    siteForm.id = site.id
    siteForm.name = site.name
    siteForm.type = site.type as SiteType
    siteForm.upstream = site.upstream || ''
    siteForm.root = site.root || ''
    siteForm.indexNames = site.index_names || ''
    siteForm.health_check = site.health_check
    dialogVisible.value = true
}

const submitSite = async () => {
    if (!siteForm.domain) {
        ElMessage.warning('请选择域名')
        return
    }

    try {
        await formRef.value?.validate()
        saving.value = true

        const domain = domains.value.find(d => d.name === siteForm.domain)
        const data = {
            name: siteForm.name,
            type: siteForm.type,
            upstream: siteForm.upstream,
            root: siteForm.root,
            indexNames: siteForm.indexNames,
            health_check: siteForm.health_check,
            domain: siteForm.domain
        }

        if (isEditing.value) {
            await siteAPI.update(siteForm.id, data)
            ElMessage.success('子站点更新成功')
        } else {
            await siteAPI.create(domain!.id, data)
            ElMessage.success('子站点创建成功')
        }

        dialogVisible.value = false
        loadAllSites()
    } catch (error: any) {
        if (error !== false) {
            ElMessage.error((isEditing.value ? '更新' : '创建') + '失败: ' + error.message)
        }
    } finally {
        saving.value = false
    }
}

const deleteSite = async (site: Site) => {
    if (!site.server_id) {
        ElMessage.warning('请先选择域名')
        return
    }

    try {
        await ElMessageBox.confirm(
            `确定要删除子站点 "${site.name}" 吗？`,
            '删除确认',
            { type: 'warning' }
        )
        await siteAPI.delete(site.server_id, site.id!)
        ElMessage.success('子站点已删除')
        loadAllSites()
    } catch (error: any) {
        if (error !== 'cancel') {
            ElMessage.error('删除失败: ' + error.message)
        }
    }
}

const handleSelectionChange = (selection: Site[]) => {
    selectedSites.value = selection
}

const batchDeleteSites = async () => {
    if (selectedSites.value.length === 0) {
        ElMessage.warning('请先选择要删除的子站点')
        return
    }

    try {
        await ElMessageBox.confirm(
            `确定要删除选中的 ${selectedSites.value.length} 个子站点吗？`,
            '批量删除确认',
            {
                type: 'warning',
                confirmButtonText: '删除',
                cancelButtonText: '取消'
            }
        )

        confirmInput.value = ''
        confirmDialogVisible.value = true
    } catch (error: any) {
        if (error !== 'cancel') {
            ElMessage.error('批量删除失败: ' + error.message)
        }
    }
}

const copyDeleteCode = () => {
    copy('DELETE')
    ElMessage.success('已复制 DELETE')
}

const executeBatchDelete = async () => {
    if (confirmInput.value !== 'DELETE') {
        ElMessage.error('请输入正确的确认码')
        return
    }

    confirmDialogVisible.value = false
    try {
        deleting.value = true
        for (const site of selectedSites.value) {
            if (site.server_id && site.id) {
                await siteAPI.delete(site.server_id, site.id)
            }
        }
        ElMessage.success('批量删除成功')
        selectedSites.value = []
        loadAllSites()
    } catch (error: any) {
        ElMessage.error('批量删除失败: ' + error.message)
    } finally {
        deleting.value = false
    }
}

onMounted(() => {
    loadDomains()
    loadAllSites()
})
</script>

<style scoped lang="scss">
.sites-container {
    padding: 24px;
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .header-right {
        display: flex;
        gap: 12px;
    }
}

.form-tip {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 8px;
}

.delete-code {
    display: inline-block;
    padding: 4px 12px;
    margin: 0 6px;
    color: var(--accent-magenta);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
        opacity: 0.8;
    }
}
</style>
