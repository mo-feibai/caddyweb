
<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="card-header">
          <h2>🚀 CaddyWeb</h2>
          <p>Caddy2 代理配置管理平台</p>
        </div>
      </template>

      <el-form :model="loginForm" :rules="rules" ref="formRef" label-width="0">
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="用户名"
            prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="密码"
            prefix-icon="Lock"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
      </el-form>

      <el-button type="primary" class="login-btn" :loading="loading" @click="handleLogin">
        登录
      </el-button>

      <div class="login-footer">
        <el-link type="primary" @click="$router.push('/setup')">初始化设置</el-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { ElMessage } from 'element-plus'

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

    await new Promise(resolve => setTimeout(resolve, 500))

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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;

  .card-header {
    text-align: center;

    h2 {
      margin: 0 0 10px;
      color: #303133;
    }

    p {
      margin: 0;
      color: #909399;
    }
  }
}

.login-btn {
  width: 100%;
  margin-top: 20px;
}

.login-footer {
  margin-top: 20px;
  text-align: center;
}
</style>
