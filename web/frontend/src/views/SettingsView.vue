<template>
  <div class="settings-page">
    <h2 class="geek-title">{{ t('settings.title') }}</h2>

    <el-card shadow="never" class="settings-card">
      <el-form label-width="140px">
        <el-form-item :label="t('settings.httpPort')">
          <el-input-number
            v-model="httpPort"
            :min="0"
            :max="65535"
            :step="1"
            :disabled="proxyStore.status.running"
            controls-position="right"
            class="port-input"
          />
          <span class="port-hint-inline">{{ t('settings.portAutoHint') }}</span>
        </el-form-item>

        <el-form-item :label="t('settings.socksPort')">
          <el-input-number
            v-model="socksPort"
            :min="0"
            :max="65535"
            :step="1"
            :disabled="proxyStore.status.running"
            controls-position="right"
            class="port-input"
          />
          <span class="port-hint-inline">{{ t('settings.portAutoHint') }}</span>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="small"
            :disabled="proxyStore.status.running"
            @click="savePorts"
          >
            {{ t('settings.save') }}
          </el-button>
          <span v-if="proxyStore.status.running" class="port-hint">
            {{ t('settings.stopProxyFirst') }}
          </span>
        </el-form-item>

        <el-divider />

        <el-form-item>
          <el-button type="danger" class="logout-btn" @click="handleLogout">
            {{ t('settings.logout') }}
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useProxyStore } from '@/stores/proxy'
import { useApi } from '@/composables/useApi'

const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()
const proxyStore = useProxyStore()
const api = useApi()

const httpPort = ref(0)
const socksPort = ref(0)

const savePorts = async () => {
  try {
    await api.put('/api/settings/proxy-ports', {
      http_port: httpPort.value,
      socks_port: socksPort.value,
    })
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
    const res = await api.get('/api/settings/proxy-ports')
    httpPort.value = res.http_port || 0
    socksPort.value = res.socks_port || 0
  } catch {}
})
</script>

<style scoped>
.settings-page {
  max-width: 1200px;
  margin: 0 auto;
}
.port-input {
  width: 180px;
}
.port-hint-inline {
  margin-left: 12px;
  font-size: 12px;
  color: var(--geek-text-secondary);
}
.port-hint {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-color-warning);
}
.logout-btn {
  font-size: 13px;
}
@media (max-width: 768px) {
  .settings-page :deep(.el-form-item__label) {
    width: auto !important;
    min-width: 0;
    text-align: left;
    padding-bottom: 4px;
  }
  .settings-page :deep(.el-form-item) {
    flex-direction: column;
    align-items: stretch;
  }
  .settings-page :deep(.el-form-item__content) {
    margin-left: 0 !important;
  }
  .port-input {
    width: 100%;
  }
  .port-hint-inline {
    margin-left: 0;
    margin-top: 4px;
  }
}
</style>
