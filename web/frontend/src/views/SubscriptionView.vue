<template>
  <div class="subscription-page">
    <div class="page-header">
      <h2>{{ t('sub.title') }}</h2>
      <el-button type="primary" @click="addDialogVisible = true">+ {{ t('sub.add') }}</el-button>
    </div>

    <el-card shadow="never">
      <el-table :data="subStore.subscriptions" stripe style="width: 100%">
        <el-table-column prop="name" :label="t('sub.name')" width="180" />
        <el-table-column prop="url" :label="t('sub.url')" min-width="260" show-overflow-tooltip />
        <el-table-column :label="t('sub.nodesCount')" width="100" align="center">
          <template #default="{ row }">{{ (row.nodes || []).length }}</template>
        </el-table-column>
        <el-table-column :label="t('sub.lastUpdate')" width="180">
          <template #default="{ row }">{{ row.last_fetched || row.updated_at || '—' }}</template>
        </el-table-column>
        <el-table-column :label="t('common.edit')" width="180" align="center">
          <template #default="{ row }">
            <el-button
              size="small"
              type="primary"
              :loading="refreshingMap[row.name]"
              @click="handleRefresh(row.name)"
            >
              {{ t('sub.refresh') }}
            </el-button>
            <el-button size="small" type="danger" plain @click="handleDelete(row.name)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <AddSubscriptionDialog
      v-model="addDialogVisible"
      @added="handleAdded"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSubscriptionStore } from '@/stores/subscription'
import { useApi } from '@/composables/useApi'
import AddSubscriptionDialog from '@/components/AddSubscriptionDialog.vue'

const { t } = useI18n()
const subStore = useSubscriptionStore()
const api = useApi()

const addDialogVisible = ref(false)
const refreshingMap = reactive<Record<string, boolean>>({})

const handleRefresh = async (name: string) => {
  refreshingMap[name] = true
  try {
    await api.post(`/api/subscriptions/${encodeURIComponent(name)}/refresh`)
    ElMessage.success(t('sub.refreshSuccess'))
    await subStore.loadConfig()
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    refreshingMap[name] = false
  }
}

const handleDelete = async (name: string) => {
  try {
    await ElMessageBox.confirm(t('sub.deleteConfirm', { name }), t('common.confirm'), {
      type: 'warning',
    })
    await api.del(`/api/subscriptions/${encodeURIComponent(name)}`)
    ElMessage.success(t('common.success'))
    await subStore.loadConfig()
  } catch {}
}

const handleAdded = async () => {
  await subStore.loadConfig()
}

onMounted(() => {
  subStore.loadConfig()
})
</script>

<style scoped>
.subscription-page {
  max-width: 1200px;
  margin: 0 auto;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.page-header h2 {
  margin: 0;
}
</style>
