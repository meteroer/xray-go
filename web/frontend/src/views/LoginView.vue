<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="login-header">
          <div class="logo-icon">
            <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="#888888" stroke-width="2">
              <path d="M12 2L2 7l10 5 10-5-10-5z"/>
              <path d="M2 17l10 5 10-5"/>
              <path d="M2 12l10 5 10-5"/>
            </svg>
          </div>
          <h2 class="login-title">{{ isRegister ? t('auth.register') : t('auth.login') }}</h2>
          <div class="login-subtitle">XRAY-GO PROXY SYSTEM</div>
        </div>
      </template>
      <p v-if="isRegister" class="init-hint">{{ t('auth.needInit') }}</p>
      <el-form :model="form" @submit.prevent="handleSubmit">
        <el-form-item :label="t('auth.username')">
          <el-input v-model="form.username" placeholder="root" />
        </el-form-item>
        <el-form-item :label="t('auth.password')">
          <el-input v-model="form.password" type="password" show-password placeholder="********" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" class="login-btn">
            {{ isRegister ? t('auth.register') : t('auth.login') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useApi } from '@/composables/useApi'

const { t } = useI18n()
const router = useRouter()
const api = useApi()

const isRegister = ref(false)
const loading = ref(false)
const form = ref({ username: '', password: '' })

onMounted(async () => {
  try {
    const res = await api.get('/api/auth/status')
    isRegister.value = res.needs_init === true || res.initialized === false
  } catch {
    isRegister.value = false
  }
})

const handleSubmit = async () => {
  loading.value = true
  try {
    const endpoint = isRegister.value ? '/api/auth/init' : '/api/auth/login'
    const res = await api.post(endpoint, form.value)
    localStorage.setItem('token', res.token)
    ElMessage.success(t(isRegister.value ? 'auth.registerSuccess' : 'auth.loginSuccess'))
    router.push('/')
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: var(--geek-bg);
}
.login-card {
  width: 420px;
}
.login-header {
  text-align: center;
}
.logo-icon {
  margin-bottom: 12px;
}
.login-title {
  margin: 0 0 4px 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--geek-text);
}
.login-subtitle {
  font-size: 12px;
  color: var(--geek-text-secondary);
}
.init-hint {
  color: var(--geek-warning);
  margin-bottom: 16px;
  font-size: 13px;
}
.login-btn {
  width: 100%;
  height: 40px;
  font-size: 14px;
}
@media (max-width: 768px) {
  .login-container {
    padding: 16px;
  }
  .login-card {
    width: 100%;
    max-width: 420px;
  }
}
</style>
