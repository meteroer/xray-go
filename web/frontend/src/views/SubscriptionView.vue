<template>
  <div class="subscription-page">
    <div class="page-header">
      <h2 class="geek-title">{{ t('sub.title') }}</h2>
      <el-button type="success" @click="addDialogVisible = true" class="add-btn">+ {{ t('sub.add') }}</el-button>
    </div>

    <el-card shadow="never" class="sub-card">
      <div class="table-scroll">
      <el-table :data="subStore.subscriptions" stripe style="width: 100%">
        <el-table-column prop="name" :label="t('sub.name')" width="180" />
        <el-table-column prop="url" :label="t('sub.url')" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="mono-url">{{ row.url }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('sub.nodesCount')" width="100" align="center">
          <template #default="{ row }"><span class="mono">{{ (row.nodes || []).length }}</span></template>
        </el-table-column>
        <el-table-column :label="t('sub.lastUpdate')" width="180">
          <template #default="{ row }"><span class="mono">{{ row.last_fetched || row.updated_at || '—' }}</span></template>
        </el-table-column>
        <el-table-column :label="t('common.edit')" width="200" align="center">
          <template #default="{ row }">
            <el-button
              size="small"
              type="primary"
              :loading="refreshingMap[row.name]"
              @click="handleRefresh(row.name)"
            >
              ↻ {{ t('sub.refresh') }}
            </el-button>
            <el-button size="small" type="danger" plain @click="handleDelete(row.name)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      </div>
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
.add-btn {
  font-size: 13px;
}
.mono {
  font-size: 13px;
  color: var(--geek-text-secondary);
}
.mono-url {
  font-size: 12px;
  color: var(--geek-text-secondary);
}
.table-scroll {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
}
</style>
