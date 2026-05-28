<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <h2>{{ isRegister ? t('auth.register') : t('auth.login') }}</h2>
      </template>
      <p v-if="isRegister" class="init-hint">{{ t('auth.needInit') }}</p>
      <el-form :model="form" @submit.prevent="handleSubmit">
        <el-form-item :label="t('auth.username')">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item :label="t('auth.password')">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%">
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
    isRegister.value = res.needs_init === true
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
}
.login-card {
  width: 400px;
}
.init-hint {
  color: #e6a23c;
  margin-bottom: 16px;
}
</style>
