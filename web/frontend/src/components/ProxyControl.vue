<template>
  <el-card class="proxy-control" shadow="hover">
    <div class="proxy-status">
      <div class="status-left">
        <el-tag :type="proxyStore.status.running ? 'success' : 'info'" size="large" effect="dark">
          {{ proxyStore.status.running ? t('proxy.running') : t('proxy.stopped') }}
        </el-tag>
        <template v-if="proxyStore.status.running">
          <span class="status-detail">
            {{ t('proxy.currentNode') }}: <strong>{{ proxyStore.status.node }}</strong>
          </span>
          <span class="status-detail">
            {{ t('proxy.httpPort') }}: {{ proxyStore.status.http_port }}
          </span>
          <span class="status-detail">
            {{ t('proxy.socksPort') }}: {{ proxyStore.status.socks_port }}
          </span>
        </template>
      </div>
      <div class="status-right">
        <el-button
          v-if="!proxyStore.status.running"
          type="success"
          @click="handleStart"
          :loading="startLoading"
        >
          {{ t('proxy.start') }}
        </el-button>
        <el-button
          v-else
          type="danger"
          @click="handleStop"
          :loading="stopLoading"
        >
          {{ t('proxy.stop') }}
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
  margin-bottom: 20px;
}
.proxy-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.status-left {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}
.status-detail {
  color: #606266;
  font-size: 14px;
}
.status-right {
  display: flex;
  gap: 8px;
}
</style>
