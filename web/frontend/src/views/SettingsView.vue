<template>
  <div class="settings-page">
    <h2 class="geek-title">{{ t('settings.title') }}</h2>

    <el-card shadow="never" class="settings-card">
      <el-form label-width="140px">
        <el-form-item :label="t('settings.language')">
          <el-switch
            :model-value="locale === 'zh'"
            active-text="中文"
            inactive-text="EN"
            @change="toggleLang"
          />
        </el-form-item>

        <el-form-item :label="t('settings.routeMode')">
          <el-select v-model="routeMode" class="mode-select">
            <el-option value="global" :label="t('routing.global')" />
            <el-option value="whitelist" :label="t('routing.whitelist')" />
            <el-option value="blacklist" :label="t('routing.blacklist')" />
          </el-select>
          <el-button type="primary" size="small" class="save-mode-btn" @click="saveRouteMode">
            {{ t('common.save') }}
          </el-button>
        </el-form-item>

        <el-divider />

        <el-form-item :label="t('settings.httpPort')">
          <el-input-number
            v-model="httpPort"
            :min="0"
            :max="65535"
            :step="1"
            :disabled="proxyStore.status.running"
            :placeholder="portPlaceholder"
            controls-position="right"
            class="port-input"
          />
        </el-form-item>

        <el-form-item :label="t('settings.socksPort')">
          <el-input-number
            v-model="socksPort"
            :min="0"
            :max="65535"
            :step="1"
            :disabled="proxyStore.status.running"
            :placeholder="portPlaceholder"
            controls-position="right"
            class="port-input"
          />
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
const { t, locale } = useI18n()
const authStore = useAuthStore()
const proxyStore = useProxyStore()
const api = useApi()

const routeMode = ref('global')
const httpPort = ref(0)
const socksPort = ref(0)
const portPlaceholder = '0 = auto'

const toggleLang = (val: boolean) => {
  const newLang = val ? 'zh' : 'en'
  locale.value = newLang
  localStorage.setItem('lang', newLang)
}

const saveRouteMode = async () => {
  try {
    await api.put('/api/settings/route-mode', { route_mode: routeMode.value })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  }
}

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
    const res = await api.get('/api/settings/route-mode')
    routeMode.value = res.route_mode || res.mode || 'global'
  } catch {}
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
.mode-select {
  width: 200px;
}
.save-mode-btn {
  margin-left: 12px;
  font-size: 12px;
}
.port-input {
  width: 180px;
}
.port-hint {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-color-warning);
}
.logout-btn {
  font-size: 13px;
}
</style>
