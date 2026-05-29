<template>
  <el-card class="proxy-control" shadow="hover">
    <div class="proxy-status">
      <div class="status-left">
        <div class="status-block">
          <span class="data-label">STATUS</span>
          <el-tag :type="proxyStore.status.running ? 'success' : 'danger'" size="large" effect="dark" class="status-tag">
            {{ proxyStore.status.running ? t('proxy.running') : t('proxy.stopped') }}
          </el-tag>
        </div>
        <template v-if="proxyStore.status.running">
          <div class="status-block">
            <span class="data-label">NODE</span>
            <span class="data-value">{{ proxyStore.status.node }}</span>
          </div>
          <div class="status-block">
            <span class="data-label">HTTP</span>
            <span class="data-value">{{ proxyStore.status.http_port }}</span>
          </div>
          <div class="status-block">
            <span class="data-label">SOCKS</span>
            <span class="data-value">{{ proxyStore.status.socks_port }}</span>
          </div>
        </template>
      </div>
      <div class="status-right">
        <el-button
          v-if="!proxyStore.status.running"
          type="success"
          size="large"
          @click="handleStart"
          :loading="startLoading"
          class="action-btn"
        >
          <span class="btn-icon">▶</span> {{ t('proxy.start') }}
        </el-button>
        <el-button
          v-else
          type="danger"
          size="large"
          @click="handleStop"
          :loading="stopLoading"
          class="action-btn"
        >
          <span class="btn-icon">■</span> {{ t('proxy.stop') }}
        </el-button>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useProxyStore } from '@/stores/proxy'
import { useApi } from '@/composables/useApi'

const { t } = useI18n()
const proxyStore = useProxyStore()
const api = useApi()

const startLoading = ref(false)
const stopLoading = ref(false)

const handleStart = async () => {
  startLoading.value = true
  try {
    const res = await api.post('/api/proxy/start')
    proxyStore.updateStatus(res)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    startLoading.value = false
  }
}

const handleStop = async () => {
  stopLoading.value = true
  try {
    const res = await api.post('/api/proxy/stop')
    proxyStore.updateStatus(res)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    stopLoading.value = false
  }
}
</script>

<style scoped>
.proxy-control {
  margin-bottom: 24px;
}
.proxy-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.status-left {
  display: flex;
  align-items: center;
  gap: 32px;
  flex-wrap: wrap;
}
.status-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.status-tag {
  font-size: 14px;
}
.status-right {
  display: flex;
  gap: 12px;
}
.action-btn {
  font-size: 14px;
  min-width: 120px;
}
.btn-icon {
  margin-right: 6px;
  font-size: 10px;
}
@media (max-width: 768px) {
  .proxy-status {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }
  .status-left {
    gap: 16px;
  }
  .status-right {
    justify-content: stretch;
  }
  .action-btn {
    min-width: 0;
    flex: 1;
  }
}
</style>
