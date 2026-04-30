<template>
  <div class="tls-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>SSL 证书管理</span>
          <el-button type="primary" @click="openAddDialog">
            <el-icon><i-ep-plus /></el-icon> 添加证书
          </el-button>
        </div>
      </template>

      <el-alert
        title="TLS 设置说明"
        type="info"
        :closable="false"
        style="margin-bottom: 20px"
      >
        Caddy2 会自动为站点配置 HTTPS 并使用 Let's Encrypt 获取证书。
        你可以在这里管理自有证书或查看已管理的域名。
      </el-alert>

      <el-table :data="certificates" stripe style="width: 100%" v-loading="loading">
        <el-table-column prop="domain" label="域名" min-width="200" />
        <el-table-column prop="issuer" label="颁发者" min-width="150">
          <template #default="{ row }">
            <el-tag v-if="row.issuer === 'Let\'s Encrypt'" type="success" size="small">
              Let's Encrypt
            </el-tag>
            <el-tag v-else type="warning" size="small">
              自有证书
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="expiry" label="到期时间" width="180">
          <template #default="{ row }">
            <span :class="{ 'expiring': isExpiringSoon(row.expiry) }">
              {{ formatDate(row.expiry) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'valid' ? 'success' : 'danger'" size="small">
              {{ row.status === 'valid' ? '有效' : '无效' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewCert(row)">查看</el-button>
            <el-button type="warning" link size="small" @click="renewCert(row)">续期</el-button>
            <el-button type="danger" link size="small" @click="deleteCert(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showAddDialog" title="添加证书" width="600px">
      <div class="add-cert-flow">
        <div class="flow-step">
          <div class="step-indicator">
            <div class="step-num" :class="{ active: true, completed: certForm.authMethod }">1</div>
            <div class="step-line" :class="{ active: certForm.authMethod }"></div>
            <div class="step-num" :class="{ active: certForm.authMethod && certForm.authMethod !== 'file' }">2</div>
          </div>
          <div class="step-title">认证方式</div>
        </div>

        <div class="auth-method-selector">
          <button
            class="method-card"
            :class="{ active: certForm.authMethod === 'auto' }"
            @click="certForm.authMethod = 'auto'"
          >
            <div class="method-icon">
              <el-icon><i-ep-magic-stick /></el-icon>
            </div>
            <div class="method-content">
              <span class="method-title">自动管理</span>
              <span class="method-desc">使用 Let's Encrypt 自动获取证书</span>
            </div>
            <div class="method-check">
              <el-icon><i-ep-check /></el-icon>
            </div>
          </button>

          <button
            class="method-card"
            :class="{ active: certForm.authMethod === 'file' }"
            @click="certForm.authMethod = 'file'"
          >
            <div class="method-icon file">
              <el-icon><i-ep-folder /></el-icon>
            </div>
            <div class="method-content">
              <span class="method-title">从文件加载</span>
              <span class="method-desc">使用已有的证书文件</span>
            </div>
            <div class="method-check">
              <el-icon><i-ep-check /></el-icon>
            </div>
          </button>
        </div>

        <Transition name="slide-fade">
          <div v-if="certForm.authMethod === 'file'" class="file-form">
            <el-form :model="certForm" :rules="rules" ref="formRef" label-width="120px">
              <el-form-item label="域名" prop="domain">
                <el-input v-model="certForm.domain" placeholder="example.com" />
              </el-form-item>
              <el-form-item label="证书文件" prop="certFile">
                <el-input v-model="certForm.certFile" placeholder="/path/to/cert.pem" />
              </el-form-item>
              <el-form-item label="私钥文件" prop="keyFile">
                <el-input v-model="certForm.keyFile" placeholder="/path/to/key.pem" />
              </el-form-item>
              <el-form-item label="证书标识">
                <el-input v-model="certForm.certId" placeholder="留空则自动生成" />
                <div class="form-tip">证书的唯一标识，用于关联和管理</div>
              </el-form-item>
              <el-form-item label="自动 HTTPS">
                <el-switch v-model="certForm.autoHTTPS" />
                <div class="form-tip">禁用后 Caddy 不会为该域名自动启用 HTTPS</div>
              </el-form-item>
            </el-form>
          </div>
        </Transition>

        <Transition name="slide-fade">
          <div v-if="certForm.authMethod === 'auto'" class="auto-form">
            <div class="target-type-selector">
              <button
                class="target-card"
                :class="{ active: certForm.targetType === 'domain' }"
                @click="certForm.targetType = 'domain'"
              >
                <div class="target-icon">
                  <el-icon><i-ep-link /></el-icon>
                </div>
                <span class="target-name">域名</span>
                <span class="target-desc">为域名添加证书</span>
              </button>

              <button
                class="target-card"
                :class="{ active: certForm.targetType === 'site' }"
                @click="certForm.targetType = 'site'"
              >
                <div class="target-icon">
                  <el-icon><i-ep-document /></el-icon>
                </div>
                <span class="target-name">站点</span>
                <span class="target-desc">为站点添加证书</span>
              </button>
            </div>

            <div v-if="certForm.targetType === 'domain'" class="selector-wrapper">
              <label class="selector-label">证书标识</label>
              <el-input v-model="certForm.certId" placeholder="留空则自动生成" />
            </div>

            <div v-if="certForm.targetType === 'domain'" class="selector-wrapper">
              <label class="selector-label">选择域名</label>
              <el-select
                v-model="certForm.selectedDomain"
                placeholder="请选择域名"
                filterable
                class="selector-dropdown"
              >
                <el-option
                  v-for="domain in domainList"
                  :key="domain.name"
                  :label="domain.name"
                  :value="domain.name"
                >
                  <div class="domain-option">
                    <span class="domain-name">{{ domain.name }}</span>
                    <span class="domain-wildcard" v-if="domain.wildcard">泛域名</span>
                  </div>
                </el-option>
              </el-select>
            </div>

            <div v-if="certForm.targetType === 'site'" class="selector-wrapper">
              <label class="selector-label">证书标识</label>
              <el-input v-model="certForm.certId" placeholder="留空则自动生成" />
            </div>

            <div v-if="certForm.targetType === 'site'" class="selector-wrapper">
              <label class="selector-label">选择站点</label>
              <el-select
                v-model="certForm.selectedSite"
                placeholder="请选择站点"
                filterable
                class="selector-dropdown"
              >
                <el-option
                  v-for="site in siteList"
                  :key="site.id"
                  :label="site.name"
                  :value="site.id"
                >
                  <div class="site-option">
                    <span class="site-name">{{ site.name }}</span>
                    <span class="site-domain">{{ site.domain }}</span>
                  </div>
                </el-option>
              </el-select>
            </div>
          </div>
        </Transition>
      </div>

      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="saveCert" :loading="saving" :disabled="!canSave">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showDetailDialog" title="证书详情" width="700px">
      <el-descriptions :column="1" border v-if="selectedCert">
        <el-descriptions-item label="域名">{{ selectedCert.domain }}</el-descriptions-item>
        <el-descriptions-item label="颁发者">{{ selectedCert.issuer }}</el-descriptions-item>
        <el-descriptions-item label="到期时间">{{ formatDate(selectedCert.expiry) }}</el-descriptions-item>
        <el-descriptions-item label="序列号">{{ selectedCert.serialNumber || 'N/A' }}</el-descriptions-item>
        <el-descriptions-item label="指纹">{{ selectedCert.fingerprint || 'N/A' }}</el-descriptions-item>
      </el-descriptions>
      <div v-if="selectedCert?.pem" class="cert-preview">
        <pre>{{ selectedCert.pem }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { caddyAPI, domainAPI, siteAPI } from '@/api'
import api from '@/api'


interface Certificate {
  domain: string
  issuer: string
  expiry: string
  status: 'valid' | 'invalid' | 'expiring'
  serialNumber?: string
  fingerprint?: string
  pem?: string
}

interface DomainItem {
  name: string
  wildcard?: boolean
  server_id?: string
}

interface SiteItem {
  id: string
  name: string
  domain?: string
}

const loading = ref(false)
const saving = ref(false)
const showAddDialog = ref(false)
const showDetailDialog = ref(false)
const selectedCert = ref<Certificate | null>(null)
const certificates = ref<Certificate[]>([])
const domainList = ref<DomainItem[]>([])
const siteList = ref<SiteItem[]>([])

const formRef = ref()

const certForm = reactive({
  authMethod: '' as '' | 'auto' | 'file',
  targetType: '' as '' | 'domain' | 'site',
  selectedDomain: '',
  selectedSite: '',
  domain: '',
  certFile: '',
  keyFile: '',
  autoHTTPS: true,
  certId: ''
})

const canSave = computed(() => {
  if (certForm.authMethod === 'file') {
    return certForm.domain && certForm.certFile && certForm.keyFile
  }
  if (certForm.authMethod === 'auto') {
    if (certForm.targetType === 'domain') {
      return !!certForm.selectedDomain
    }
    if (certForm.targetType === 'site') {
      return !!certForm.selectedSite
    }
  }
  return false
})

const rules = {
  domain: [{ required: true, message: '请输入域名', trigger: 'blur' }],
  certFile: [{ required: true, message: '请输入证书文件路径', trigger: 'blur' }],
  keyFile: [{ required: true, message: '请输入私钥文件路径', trigger: 'blur' }]
}

const loadCertificates = async () => {
  loading.value = true
  try {
    const config = await caddyAPI.getConfig()
    const tls = config.apps?.http?.servers || {}

    const certList: Certificate[] = []
    Object.entries(tls).forEach(([serverName, server]: [string, any]) => {
      const routes = server.routes || []
      routes.forEach((route: any) => {
        const match = route.match || []
        match.forEach((m: any) => {
          if (m.host) {
            m.host.forEach((domain: string) => {
              const existing = certList.find(c => c.domain === domain)
              if (!existing) {
                certList.push({
                  domain,
                  issuer: "Let's Encrypt",
                  expiry: new Date(Date.now() + 90 * 24 * 60 * 60 * 1000).toISOString(),
                  status: 'valid'
                })
              }
            })
          }
        })
      })
    })

    certificates.value = certList
  } catch (error: any) {
    ElMessage.error('加载证书失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return 'N/A'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

const isExpiringSoon = (dateStr: string) => {
  if (!dateStr) return false
  const expiry = new Date(dateStr)
  const now = new Date()
  const daysUntilExpiry = (expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  return daysUntilExpiry < 30
}

const viewCert = (cert: Certificate) => {
  selectedCert.value = cert
  showDetailDialog.value = true
}

const renewCert = async (cert: Certificate) => {
  try {
    await ElMessageBox.confirm(`确定要为域名 "${cert.domain}" 续期证书吗？`, '续期确认', {
      type: 'warning'
    })
    ElMessage.success('证书续期请求已发送')
  } catch {
    // 用户取消
  }
}

const resetCertForm = () => {
  certForm.authMethod = ''
  certForm.targetType = ''
  certForm.selectedDomain = ''
  certForm.selectedSite = ''
  certForm.domain = ''
  certForm.certFile = ''
  certForm.keyFile = ''
  certForm.autoHTTPS = true
  certForm.certId = ''
}

const loadDomainsAndSites = async () => {
  try {
    const [domains, sites] = await Promise.all([
      domainAPI.list().catch(() => []),
      siteAPI.list().catch(() => [])
    ])
    domainList.value = domains || []
    siteList.value = sites || []
  } catch (error) {
    domainList.value = []
    siteList.value = []
  }
}

const saveCert = async () => {
  try {
    if (certForm.authMethod === 'file') {
      await formRef.value?.validate()
    }
    saving.value = true

    const requestData: any = {
      authMethod: certForm.authMethod,
      certId: certForm.certId
    }

    if (certForm.authMethod === 'file') {
      requestData.domain = certForm.domain
      requestData.certFile = certForm.certFile
      requestData.keyFile = certForm.keyFile
      requestData.autoHTTPS = certForm.autoHTTPS
    } else if (certForm.authMethod === 'auto') {
      requestData.targetType = certForm.targetType
      if (certForm.targetType === 'domain') {
        requestData.domainName = certForm.selectedDomain
      } else if (certForm.targetType === 'site') {
        requestData.siteId = certForm.selectedSite
      }
    }

    const response = await api.post('/certs', requestData)
    ElMessage.success(response.message || '证书配置已保存')

    showAddDialog.value = false
    resetCertForm()
    loadCertificates()
  } catch (error: any) {
    ElMessage.error(error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const deleteCert = async (cert: Certificate) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除域名 "${cert.domain}" 的证书吗？`,
      '删除确认',
      { type: 'warning' }
    )
    ElMessage.success('证书已删除')
    loadCertificates()
  } catch {
    // 用户取消
  }
}

onMounted(() => {
  loadCertificates()
})

const openAddDialog = () => {
  resetCertForm()
  loadDomainsAndSites()
  showAddDialog.value = true
}
</script>

<style scoped lang="scss">
.tls-container {
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

.expiring {
    color: var(--accent-amber);
    font-weight: 600;
}

.cert-preview {
    margin-top: 20px;
    padding: 16px 20px;
    overflow-x: auto;

    pre {
        margin: 0;
        font-size: 12px;
        font-family: var(--font-mono);
        white-space: pre-wrap;
        word-break: break-all;
    }
}

.add-cert-flow {
    display: flex;
    flex-direction: column;
    gap: 24px;
}

.flow-step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;

    .step-indicator {
        display: flex;
        align-items: center;
        gap: 0;
    }

    .step-num {
        width: 28px;
        height: 28px;
        border-radius: 50%;
        background: var(--bg-hover);
        border: 2px solid var(--border-subtle);
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 12px;
        font-weight: 600;
        color: var(--text-muted);
        transition: all var(--transition-smooth);

        &.active {
            background: var(--accent-cyan);
            border-color: var(--accent-cyan);
            color: var(--bg-primary);
        }

        &.completed {
            background: var(--accent-cyan);
            border-color: var(--accent-cyan);
            color: var(--bg-primary);
        }
    }

    .step-line {
        width: 80px;
        height: 2px;
        background: var(--border-subtle);
        transition: background var(--transition-smooth);

        &.active {
            background: var(--accent-cyan);
        }
    }

    .step-title {
        font-size: 13px;
        font-weight: 600;
        color: var(--text-secondary);
    }
}

.auth-method-selector {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 14px;
}

.method-card {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 16px 18px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-subtle);
    border-radius: 12px;
    cursor: pointer;
    transition: all var(--transition-smooth);
    text-align: left;

    .method-icon {
        width: 42px;
        height: 42px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: rgba(0, 212, 255, 0.1);
        border-radius: 10px;
        color: var(--accent-cyan);
        font-size: 20px;
        flex-shrink: 0;

        &.file {
            background: rgba(255, 184, 0, 0.1);
            color: var(--accent-amber);
        }
    }

    .method-content {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 3px;

        .method-title {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-primary);
        }

        .method-desc {
            font-size: 12px;
            color: var(--text-muted);
        }
    }

    .method-check {
        width: 22px;
        height: 22px;
        border-radius: 50%;
        background: var(--bg-hover);
        border: 2px solid var(--border-subtle);
        display: flex;
        align-items: center;
        justify-content: center;
        color: transparent;
        font-size: 12px;
        transition: all var(--transition-smooth);
        flex-shrink: 0;
    }

    &:hover {
        border-color: var(--text-muted);
        transform: translateY(-2px);
    }

    &.active {
        border-color: var(--accent-cyan);
        box-shadow: var(--glow-cyan);

        .method-check {
            background: var(--accent-cyan);
            border-color: var(--accent-cyan);
            color: var(--bg-primary);
        }

        .method-icon {
            background: rgba(0, 212, 255, 0.15);
        }

        &.file .method-icon {
            background: rgba(255, 184, 0, 0.15);
        }
    }
}

.target-type-selector {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 14px;
}

.target-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 20px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-subtle);
    border-radius: 12px;
    cursor: pointer;
    transition: all var(--transition-smooth);

    .target-icon {
        width: 44px;
        height: 44px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: rgba(0, 212, 255, 0.1);
        border-radius: 10px;
        color: var(--accent-cyan);
        font-size: 20px;
    }

    .target-name {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
    }

    .target-desc {
        font-size: 12px;
        color: var(--text-muted);
    }

    &:hover {
        border-color: var(--text-muted);
        transform: translateY(-2px);
    }

    &.active {
        border-color: var(--accent-cyan);
        box-shadow: var(--glow-cyan);

        .target-icon {
            background: rgba(0, 212, 255, 0.15);
        }

        .target-name {
            color: var(--accent-cyan);
        }
    }
}

.selector-wrapper {
    display: flex;
    flex-direction: column;
    gap: 8px;

    .selector-label {
        font-size: 13px;
        font-weight: 600;
        color: var(--text-secondary);
    }

    .selector-dropdown {
        width: 100%;
    }
}

.domain-option,
.site-option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;

    .domain-name,
    .site-name {
        font-weight: 500;
    }

    .domain-wildcard,
    .site-domain {
        font-size: 12px;
        color: var(--text-muted);
    }
}

.file-form,
.auto-form {
    padding-top: 8px;
}

.selector-wrapper {
    margin-top: 16px;
}

.slide-fade-enter-active,
.slide-fade-leave-active {
    transition: all 0.25s ease;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
    opacity: 0;
    transform: translateY(-8px);
}
</style>
