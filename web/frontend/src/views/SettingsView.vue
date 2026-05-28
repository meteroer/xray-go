<template>
  <div class="settings-page">
    <h2>{{ t('settings.title') }}</h2>

    <el-card shadow="never" style="margin-top: 16px">
      <el-form label-width="120px">
        <el-form-item :label="t('settings.language')">
          <el-switch
            :model-value="locale === 'zh'"
            active-text="中文"
            inactive-text="EN"
            @change="toggleLang"
          />
        </el-form-item>

        <el-form-item :label="t('settings.routeMode')">
          <el-select v-model="routeMode" style="width: 200px">
            <el-option value="global" :label="t('routing.global')" />
            <el-option value="whitelist" :label="t('routing.whitelist')" />
            <el-option value="blacklist" :label="t('routing.blacklist')" />
          </el-select>
          <el-button type="primary" size="small" style="margin-left: 12px" @click="saveRouteMode">
            {{ t('common.save') }}
          </el-button>
        </el-form-item>

        <el-form-item>
          <el-button type="danger" @click="handleLogout">{{ t('settings.logout') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'

const router = useRouter()
const { t, locale } = useI18n()
const authStore = useAuthStore()
const api = useApi()

const routeMode = ref('global')

const toggleLang = (val: boolean) => {
  const newLang = val ? 'zh' : 'en'
  locale.value = newLang
  localStorage.setItem('lang', newLang)
}

const saveRouteMode = async () => {
  try {
    await api.put('/api/settings/route-mode', { mode: routeMode.value })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  }
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm(t('settings.logoutConfirm'), t('common.confirm'), {
      type: 'warning',
    })
    authStore.clearToken()
    router.push('/login')
  } catch {}
}

onMounted(async () => {
  try {
    const res = await api.get('/api/settings/route-mode')
    routeMode.value = res.mode || 'global'
  } catch {}
})
</script>

<style scoped>
.settings-page {
  max-width: 1200px;
  margin: 0 auto;
}
</style>
