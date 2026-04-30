<template>
  <div class="login-container">
    <div class="login-bg">
      <div class="grid-lines"></div>
      <div class="glow-orb orb-1"></div>
      <div class="glow-orb orb-2"></div>
    </div>

    <div class="login-wrapper">
      <div class="login-card">
        <div class="card-glow"></div>

        <div class="card-header">
          <div class="logo-mark">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
              <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="var(--accent-cyan)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M2 17L12 22L22 17" stroke="var(--accent-cyan)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M2 12L12 17L22 12" stroke="var(--accent-cyan)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <h1 class="title">CaddyWeb</h1>
          <p class="subtitle">Caddy2 代理配置管理平台</p>
        </div>

        <el-form :model="loginForm" :rules="rules" ref="formRef" class="login-form">
          <div class="form-group">
            <label class="form-label">用户名</label>
            <el-input
              v-model="loginForm.username"
              placeholder="输入用户名"
              size="large"
              class="form-input"
            >
              <template #prefix>
                <el-icon class="input-icon"><i-ep-user /></el-icon>
              </template>
            </el-input>
          </div>

          <div class="form-group">
            <label class="form-label">密码</label>
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="输入密码"
              size="large"
              class="form-input"
              @keyup.enter="handleLogin"
            >
              <template #prefix>
                <el-icon class="input-icon"><i-ep-lock /></el-icon>
              </template>
            </el-input>
          </div>

          <button type="button" class="login-btn" :class="{ loading }" :disabled="loading" @click="handleLogin">
            <span class="btn-text">{{ loading ? '登录中...' : '登录' }}</span>
            <span class="btn-icon">
              <el-icon v-if="!loading"><i-ep-arrow-right /></el-icon>
              <el-icon v-else class="spin"><i-ep-loading /></el-icon>
            </span>
          </button>
        </el-form>

        <div class="card-footer">
          <span class="footer-text">首次使用？</span>
          <a href="#" class="footer-link" @click.prevent="$router.push('/setup')">初始化设置</a>
        </div>
      </div>

      <div class="version-tag">v2.0.0</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useSettingsStore } from '@/stores/settings'

const router = useRouter()
const settingsStore = useSettingsStore()

const formRef = ref()
const loading = ref(false)

const loginForm = ref({
  username: 'admin',
  password: 'admin'
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  try {
    await formRef.value?.validate()
    loading.value = true

    await new Promise(resolve => setTimeout(resolve, 800))

    if (loginForm.value.username === 'admin' && loginForm.value.password === 'admin') {
      ElMessage.success('登录成功')
      router.push('/web/dashboard')
    } else {
      ElMessage.error('用户名或密码错误')
    }
  } catch {
    // 验证失败
  } finally {
    loading.value = false
  }
}

if (settingsStore.settings.firstLaunch) {
  router.push('/setup')
}
</script>

<style scoped lang="scss">
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.login-bg {
  position: absolute;
  inset: 0;
  background: var(--bg-primary);

  .grid-lines {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(var(--border-subtle) 1px, transparent 1px),
      linear-gradient(90deg, var(--border-subtle) 1px, transparent 1px);
    background-size: 60px 60px;
    opacity: 0.3;
  }

  .glow-orb {
    position: absolute;
    border-radius: 50%;
    filter: blur(100px);

    &.orb-1 {
      width: 500px;
      height: 500px;
      background: var(--accent-cyan);
      opacity: 0.08;
      top: -200px;
      right: -100px;
    }

    &.orb-2 {
      width: 400px;
      height: 400px;
      background: var(--accent-magenta);
      opacity: 0.06;
      bottom: -150px;
      left: -100px;
    }
  }
}

.login-wrapper {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.login-card {
  width: 420px;
  padding: 48px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  position: relative;
  overflow: hidden;
  animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1);

  .card-glow {
    position: absolute;
    top: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 200px;
    height: 200px;
    background: var(--accent-cyan);
    opacity: 0.1;
    filter: blur(80px);
    pointer-events: none;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-header {
  text-align: center;
  margin-bottom: 40px;

  .logo-mark {
    width: 80px;
    height: 80px;
    margin: 0 auto 24px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: var(--glow-cyan);
  }

  .title {
    font-size: 32px;
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: 8px;
    letter-spacing: -1px;
  }

  .subtitle {
    font-size: 14px;
    color: var(--text-secondary);
    margin: 0;
  }
}

.login-form {
  .form-group {
    margin-bottom: 24px;

    .form-label {
      display: block;
      font-size: 12px;
      font-weight: 600;
      color: var(--text-secondary);
      text-transform: uppercase;
      letter-spacing: 0.5px;
      margin-bottom: 10px;
    }

    .form-input {
      :deep(.el-input__wrapper) {
        padding: 14px 16px;
        border-radius: var(--radius-md);
      }

      :deep(.el-input__inner) {
        font-size: 15px;
        color: var(--text-primary);

        &::placeholder {
          color: var(--text-muted);
        }
      }

      .input-icon {
        color: var(--text-muted);
        margin-right: 4px;
      }
    }
  }
}

.login-btn {
  width: 100%;
  height: 52px;
  margin-top: 8px;
  background: linear-gradient(135deg, var(--accent-cyan), var(--accent-cyan-dim));
  border: none;
  border-radius: var(--radius-md);
  color: var(--bg-primary);
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(135deg, transparent, rgba(255,255,255,0.2), transparent);
    transform: translateX(-100%);
    transition: transform 0.5s;
  }

  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0, 212, 255, 0.3);

    &::before {
      transform: translateX(100%);
    }
  }

  &:active:not(:disabled) {
    transform: translateY(0);
  }

  &.loading {
    pointer-events: none;
    opacity: 0.8;
  }

  .btn-icon {
    display: flex;
    align-items: center;

    .spin {
      animation: spin 1s linear infinite;
    }
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.card-footer {
  margin-top: 32px;
  text-align: center;

  .footer-text {
    font-size: 13px;
    color: var(--text-muted);
  }

  .footer-link {
    font-size: 13px;
    color: var(--accent-cyan);
    text-decoration: none;
    margin-left: 6px;
    transition: color 0.2s;

    &:hover {
      color: var(--accent-cyan-dim);
    }
  }
}

.version-tag {
  margin-top: 24px;
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 2px;
}
</style>