<template>
    <div class="setup-container">
        <el-card class="setup-card">
            <template #header>
                <div class="card-header">
                    <h2>🚀 CaddyWeb 初始化设置</h2>
                    <p class="subtitle">在开始之前，我们需要检测并配置 Caddy2</p>
                </div>
            </template>

            <el-steps :active="currentStep" finish-status="success" align-center class="setup-steps">
                <el-step title="检测安装" />
                <el-step title="配置连接" />
                <el-step title="环境检测" />
                <el-step title="初始化" />
                <el-step title="完成" />
            </el-steps>

            <div class="step-content">
                <!-- Step 0: 检测 Caddy 安装状态 -->
                <div v-if="currentStep === 0" class="step-panel">
                    <h3>检测 Caddy2 安装状态</h3>

                    <div class="check-status">
                        <el-icon v-if="checkStatus === 'checking'" class="is-loading" :size="48">
                            <i-ep-loading />
                        </el-icon>
                        <el-icon v-else-if="checkStatus === 'installed'" :size="48" color="#67C23A">
                            <i-ep-circle-check />
                        </el-icon>
                        <el-icon v-else :size="48" color="#F56C6C">
                            <i-ep-circle-close />
                        </el-icon>
                    </div>

                    <p class="check-message">{{ checkMessage }}</p>

                    <div v-if="checkStatus !== 'checking'" class="install-options">
                        <!-- 已安装 Caddy -->
                        <template v-if="caddyInfo.installed && !caddyInfo.running">
                            <p class="info-text">检测到 Caddy 已安装但未运行</p>
                            <div class="option-buttons">
                                <el-button type="primary" size="large" @click="useExistingCaddy">
                                    ▶️ 启动已安装的 Caddy
                                </el-button>
                                <el-button size="large" @click="startReinstall">
                                    🔄 重新安装
                                </el-button>
                            </div>
                            <p class="version-info" v-if="caddyInfo.version">版本: {{ caddyInfo.version }}</p>
                        </template>

                        <!-- Caddy 正在运行 -->
                        <template v-else-if="caddyInfo.running">
                            <p class="info-text success">Caddy2 正在运行中</p>
                            <div class="option-buttons">
                                <el-button type="primary" size="large" @click="skipInstall">
                                    ✅ 使用当前 Caddy
                                </el-button>
                                <el-button size="large" @click="startReinstall">
                                    🔄 重新安装
                                </el-button>
                            </div>
                            <p class="version-info" v-if="caddyInfo.version">版本: {{ caddyInfo.version }}</p>
                        </template>

                        <!-- 未安装 Caddy -->
                        <template v-else>
                            <p class="info-text">未检测到 Caddy2 安装</p>
                            <div class="option-buttons">
                                <el-button type="primary" size="large" @click="autoInstallCaddy">
                                    📥 自动安装 Caddy2
                                </el-button>
                            </div>
                            <p class="help-text">
                                或参考 <a href="https://caddyserver.com/docs/install" target="_blank">官方文档</a> 手动安装
                            </p>
                        </template>

                        <!-- 安装进度 -->
                        <div v-if="installing" class="install-progress">
                            <el-progress :percentage="installProgress" :stroke-width="10" />
                            <p>{{ installMessage }}</p>
                        </div>
                    </div>
                </div>

                <!-- Step 1: 配置连接 -->
                <div v-if="currentStep === 1" class="step-panel">
                    <h3>Caddy2 连接配置</h3>

                    <el-form :model="localConfig" label-width="160px" class="config-form">
                        <el-form-item label="连接方式">
                            <el-radio-group v-model="connectionType">
                                <el-radio value="socket">Unix Socket</el-radio>
                                <el-radio value="tcp">TCP 端口</el-radio>
                            </el-radio-group>
                        </el-form-item>

                        <el-form-item v-if="connectionType === 'socket'" label="Unix Socket 路径">
                            <el-input v-model="localConfig.unixSocket" placeholder="/var/run/caddy/caddy.sock" />
                            <div class="form-tip">Caddy2 Admin API 的 Unix Socket 文件路径</div>
                        </el-form-item>

                        <el-form-item v-else label="TCP 端口">
                            <el-input-number v-model="localConfig.adminPort" :min="1" :max="65535" />
                            <div class="form-tip">默认端口: 2019</div>
                        </el-form-item>
                    </el-form>
                </div>

                <!-- Step 2: 检测连接 -->
                <div v-if="currentStep === 2" class="step-panel">
                    <h3>正在检测环境...</h3>

                    <div class="detect-status">
                        <el-icon v-if="detectStatus === 'checking'" class="is-loading" :size="48">
                            <i-ep-loading />
                        </el-icon>
                        <el-icon v-else-if="detectStatus === 'success'" :size="48" color="#67C23A">
                            <i-ep-circle-check />
                        </el-icon>
                        <el-icon v-else :size="48" color="#F56C6C">
                            <i-ep-circle-close />
                        </el-icon>
                    </div>

                    <p class="detect-message">{{ detectMessage }}</p>

                    <div v-if="detectStatus === 'failed'" class="retry-section">
                        <el-button @click="detectConnection">重试</el-button>
                        <el-button @click="currentStep = 1">修改配置</el-button>
                    </div>
                </div>

                <!-- Step 3: 初始化 -->
                <div v-if="currentStep === 3" class="step-panel">
                    <h3>初始化 Caddy</h3>

                    <div class="init-status">
                        <el-icon v-if="initStatus === 'checking'" class="is-loading" :size="48">
                            <i-ep-loading />
                        </el-icon>
                        <el-icon v-else-if="initStatus === 'success'" :size="48" color="#67C23A">
                            <i-ep-circle-check />
                        </el-icon>
                        <el-icon v-else-if="initStatus === 'warning'" :size="48" color="#E6A23C">
                            <i-ep-warning />
                        </el-icon>
                        <el-icon v-else :size="48" color="#F56C6C">
                            <i-ep-circle-close />
                        </el-icon>
                    </div>

                    <p class="init-message">{{ initMessage }}</p>

                    <div v-if="initStatus === 'warning'" class="warning-section">
                        <el-alert type="warning" :closable="false" show-icon>
                            检测到已存在的 Caddy 配置，这将会覆盖现有的服务器配置。
                        </el-alert>
                        <div class="warning-buttons">
                            <el-button @click="skipInit">跳过</el-button>
                            <el-button type="warning" @click="forceInitCaddy">强制初始化</el-button>
                        </div>
                    </div>

                    <div v-if="initStatus === 'failed'" class="retry-section">
                        <el-button @click="initCaddy">重试</el-button>
                        <el-button @click="currentStep = 2">返回</el-button>
                    </div>
                </div>

                <!-- Step 4: 完成 -->
                <div v-if="currentStep === 4" class="step-panel">
                    <div class="success-panel">
                        <el-icon :size="80" color="#67C23A">
                            <i-ep-circle-check />
                        </el-icon>
                        <h3>设置完成！</h3>
                        <p>点击下方按钮进入管理界面</p>
                    </div>
                </div>
            </div>

            <template #footer>
                <div class="card-footer">
                    <el-button v-if="currentStep > 0 && currentStep < 4" @click="prevStep">上一步</el-button>
                    <el-button v-if="currentStep === 1" type="primary" @click="nextStep">检测连接</el-button>
                    <el-button v-if="currentStep === 2 && detectStatus === 'success'" type="primary" @click="nextStep">
                        下一步
                    </el-button>
                    <el-button v-if="currentStep === 3 && initStatus === 'success'" type="primary" @click="nextStep">
                        完成
                    </el-button>
                    <el-button v-if="currentStep === 3 && initStatus === 'none'" type="primary" @click="initCaddy">
                        初始化 Caddy
                    </el-button>
                    <el-button v-if="currentStep === 4" type="success" @click="finishSetup">
                        进入管理界面
                    </el-button>
                </div>
            </template>
        </el-card>
    </div>
</template>

<script setup lang="ts">
import { useSettingsStore } from '@/stores/settings'
import { settingsAPI } from '@/api'

const router = useRouter()
const settingsStore = useSettingsStore()

const currentStep = ref(0)
const checkStatus = ref<'checking' | 'installed' | 'not_installed'>('checking')
const checkMessage = ref('正在检测...')
const detectStatus = ref<'checking' | 'success' | 'failed'>('checking')
const detectMessage = ref('正在检测连接...')
const initStatus = ref<'none' | 'checking' | 'success' | 'warning' | 'failed'>('none')
const initMessage = ref('正在初始化...')
const installing = ref(false)
const installProgress = ref(0)
const installMessage = ref('')
const connectionType = ref('socket')

const caddyInfo = ref({
    installed: false,
    running: false,
    version: ''
})

const localConfig = ref({
    unixSocket: '/var/run/caddy/caddy.sock',
    adminPort: 2019
})

const checkCaddyInstallStatus = async () => {
    checkStatus.value = 'checking'
    checkMessage.value = '正在检测 Caddy2 安装状态...'

    try {
        const data: any = await settingsAPI.checkCaddyInstallStatus()
        caddyInfo.value = {
            installed: data.installed,
            running: data.running,
            version: data.version || ''
        }

        if (data.installed) {
            if (data.running) {
                checkStatus.value = 'installed'
                checkMessage.value = '✅ Caddy2 已安装并正在运行'
            } else {
                checkStatus.value = 'installed'
                checkMessage.value = '✅ Caddy2 已安装但未运行'
            }
        } else {
            checkStatus.value = 'not_installed'
            checkMessage.value = '❌ 未检测到 Caddy2 安装'
        }
    } catch (error: any) {
        checkStatus.value = 'not_installed'
        checkMessage.value = '检测失败: ' + (error.message || '未知错误')
    }
}

const useExistingCaddy = () => {
    currentStep.value = 1
}

const skipInstall = () => {
    currentStep.value = 1
}

const startReinstall = () => {
    autoInstallCaddy()
}

const autoInstallCaddy = async () => {
    installing.value = true
    installProgress.value = 0
    installMessage.value = '正在安装 Caddy2...'

    try {
        // 模拟安装进度
        const interval = setInterval(() => {
            if (installProgress.value < 85) {
                installProgress.value += Math.random() * 10
                installMessage.value = '正在下载并配置...'
            }
        }, 500)

        const response: any = await settingsAPI.installCaddy('auto')

        clearInterval(interval)
        installMessage.value = '安装完成！'

        if (response.success) {
            installProgress.value = 100
            caddyInfo.value.installed = true
            caddyInfo.value.version = response.version || ''

            setTimeout(() => {
                installing.value = false
                currentStep.value = 1
            }, 500)
        } else {
            ElMessage.error(response.message || '安装失败')
            installing.value = false
        }
    } catch (error: any) {
        ElMessage.error('安装失败: ' + (error.message || '未知错误'))
        installing.value = false
    }
}

const nextStep = async () => {
    if (currentStep.value === 1) {
        // 保存配置
        settingsStore.updateCaddySettings({
            unixSocket: localConfig.value.unixSocket,
            adminPort: localConfig.value.adminPort
        })
    }

    if (currentStep.value === 2) {
        currentStep.value++
        return
    }

    if (currentStep.value === 3 && initStatus.value === 'success') {
        currentStep.value++
        return
    }

    currentStep.value++

    if (currentStep.value === 2) {
        await detectConnection()
    }
}

const prevStep = () => {
    if (currentStep.value === 4 && initStatus.value === 'none') {
        currentStep.value = 3
        return
    }
    if (currentStep.value > 0) {
        currentStep.value--
    }
}

const detectConnection = async () => {
    detectStatus.value = 'checking'
    detectMessage.value = '正在检测 Caddy2 连接...'

    try {
        const response: any = await settingsAPI.detectCaddy()
        if (response.success) {
            detectStatus.value = 'success'
            detectMessage.value = '✅ Caddy2 连接成功！'
            caddyInfo.value.running = true
            caddyInfo.value.version = response.version || ''
        } else {
            detectStatus.value = 'failed'
            detectMessage.value = response.message || '无法连接到 Caddy2'
        }
    } catch (error: any) {
        detectStatus.value = 'failed'
        detectMessage.value = '连接失败: ' + (error.message || '请检查 Caddy2 是否运行')
    }
}

const initCaddy = async () => {
    initStatus.value = 'checking'
    initMessage.value = '正在初始化 Caddy 配置...'

    try {
        const response: any = await settingsAPI.initCaddy(false)
        if (response.warning) {
            initStatus.value = 'warning'
            initMessage.value = response.message || '检测到已存在的 Caddy 配置'
        } else if (response.success) {
            initStatus.value = 'success'
            initMessage.value = '✅ Caddy 初始化成功！'
        } else {
            initStatus.value = 'failed'
            initMessage.value = response.message || '初始化失败'
        }
    } catch (error: any) {
        initStatus.value = 'failed'
        initMessage.value = '初始化失败: ' + (error.message || '未知错误')
    }
}

const forceInitCaddy = async () => {
    initStatus.value = 'checking'
    initMessage.value = '正在强制初始化 Caddy 配置...'

    try {
        const response: any = await settingsAPI.initCaddy(true)
        if (response.success) {
            initStatus.value = 'success'
            initMessage.value = '✅ Caddy 初始化成功！'
        } else {
            initStatus.value = 'failed'
            initMessage.value = response.message || '初始化失败'
        }
    } catch (error: any) {
        initStatus.value = 'failed'
        initMessage.value = '初始化失败: ' + (error.message || '未知错误')
    }
}

const skipInit = () => {
    initStatus.value = 'none'
    currentStep.value++
}

const finishSetup = () => {
    settingsStore.completeSetup()
    router.push('/web/dashboard')
}

onMounted(async () => {
    await checkCaddyInstallStatus()
})
</script>

<style scoped lang="scss">
.setup-container {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
}

.setup-card {
    width: 100%;
    max-width: 650px;
}

:deep(.el-card__header) {
    text-align: center;
    padding: 32px 24px;
    border-bottom: 1px solid var(--border-subtle);

    h2 {
        margin: 0 0 10px;
        color: var(--text-primary);
        font-size: 24px;
    }

    .subtitle {
        color: var(--text-secondary);
        margin: 0;
        font-size: 14px;
    }
}

:deep(.el-steps) {
    margin: 40px 0;

    .el-step__title {
        color: var(--text-secondary);
    }

    .el-step__title.is-finish {
        color: var(--el-color-primary);
    }

    .el-step__icon {
        background: var(--bg-hover);
        border-color: var(--border-subtle);
    }

    .el-step__icon-inner {
        color: var(--text-muted);
    }

    .el-step__line {
        background: var(--border-subtle);
    }

    .el-step.is-finish .el-step__line {
        background: var(--el-color-primary);
    }
}

.step-content {
    min-height: 320px;
    padding: 20px 0;
}

.step-panel {
    h3 {
        text-align: center;
        margin-bottom: 30px;
        color: var(--text-primary);
    }
}

.check-status,
.detect-status,
.init-status {
    text-align: center;
    margin: 30px 0;
}

.check-message,
.detect-message,
.init-message {
    text-align: center;
    color: var(--text-secondary);
    font-size: 18px;
    margin-bottom: 30px;
}

.info-text {
    text-align: center;
    color: var(--text-secondary);
    font-size: 16px;
    margin-bottom: 20px;

    &.success {
        color: var(--el-color-success);
    }
}

.version-info {
    text-align: center;
    color: var(--text-muted);
    font-size: 14px;
    margin-top: 15px;
    font-family: var(--font-mono);
}

.help-text {
    text-align: center;
    color: var(--text-muted);
    font-size: 13px;
    margin-top: 15px;

    a {
        color: var(--el-color-primary);
    }
}

.option-buttons {
    display: flex;
    flex-direction: column;
    gap: 15px;
    max-width: 280px;
    margin: 0 auto;
}

.install-progress {
    max-width: 300px;
    margin: 30px auto 0;

    p {
        text-align: center;
        color: var(--text-muted);
        margin-top: 10px;
    }
}

.config-form {
    max-width: 400px;
    margin: 0 auto;

    :deep(.el-form-item__label) {
        color: var(--text-secondary);
    }
}

.form-tip {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 8px;
    line-height: 1.4;
}

.retry-section {
    text-align: center;
    margin-top: 30px;
    display: flex;
    gap: 15px;
    justify-content: center;
}

.warning-section {
    max-width: 400px;
    margin: 0 auto;

    :deep(.el-alert) {
        margin-bottom: 20px;
    }

    .warning-buttons {
        display: flex;
        gap: 15px;
        justify-content: center;
    }
}

.success-panel {
    text-align: center;
    padding: 50px 0;

    h3 {
        margin: 25px 0 15px;
        font-size: 24px;
        color: var(--text-primary);
    }

    p {
        color: var(--text-secondary);
        font-size: 16px;
    }
}

.card-footer {
    display: flex;
    justify-content: center;
    gap: 15px;
    padding: 20px 0;
    border-top: 1px solid var(--border-subtle);
}

:deep(.el-radio) {
    color: var(--text-primary);
    margin-right: 16px;
}
</style>
