<template>
  <div class="node-table">
    <div class="toolbar">
      <el-select v-model="selectedRegion" :placeholder="t('node.region')" clearable class="region-select">
        <el-option :label="t('node.allRegions')" value="" />
        <el-option v-for="r in regions" :key="r" :label="r" :value="r" />
      </el-select>
      <el-button type="primary" :loading="testLoading" @click="handleTestLatency" class="test-btn">
        {{ t('node.testLatency') }}
      </el-button>
      <span v-if="testProgress" class="test-progress">{{ testProgress }}</span>
      <div class="flex-grow" />
      <el-button type="success" @click="addDialogVisible = true" class="add-btn">+ {{ t('node.add') }}</el-button>
    </div>

    <el-collapse v-model="activeGroups" class="node-collapse">
      <el-collapse-item
        v-for="sub in subStore.subscriptions"
        :key="sub.name"
        :title="`[${sub.name}] — ${filteredNodes(sub.nodes).length} nodes`"
        :name="sub.name"
        class="sub-group"
      >
        <div class="table-scroll">
        <el-table :data="filteredNodes(sub.nodes)" stripe style="width: 100%" size="small" :row-class-name="rowClassName">
          <el-table-column prop="name" :label="t('node.name')" min-width="160" show-overflow-tooltip />
          <el-table-column :label="t('node.address')" min-width="180">
            <template #default="{ row }"><span class="mono">{{ row.address }}:{{ row.port }}</span></template>
          </el-table-column>
          <el-table-column prop="protocol" :label="t('node.protocol')" width="100">
            <template #default="{ row }"><span class="protocol-tag">{{ row.protocol }}</span></template>
          </el-table-column>
          <el-table-column :label="t('node.latency')" width="120" align="center" sortable>
            <template #default="{ row }">
              <LatencyTag :latency="latencyMap[row.name]" />
            </template>
          </el-table-column>
          <el-table-column :label="t('common.edit')" width="140" align="center">
            <template #default="{ row }">
              <template v-if="proxyStore.status.running && proxyStore.status.node === row.name">
                <el-tag type="success" size="small" effect="dark">{{ t('node.current') }}</el-tag>
                <el-button type="danger" size="small" @click="handleStop" style="margin-left:4px">
                  {{ t('node.disconnect') }}
                </el-button>
              </template>
              <template v-else-if="proxyStore.status.running">
                <el-button type="warning" size="small" @click="handleSwitch(row.name)">
                  {{ t('node.switchTo') }}
                </el-button>
              </template>
              <template v-else>
                <el-button type="primary" size="small" @click="handleConnect(row.name)">
                  {{ t('node.connect') }}
                </el-button>
              </template>
            </template>
          </el-table-column>
        </el-table>
        </div>
      </el-collapse-item>

      <el-collapse-item
        v-if="subStore.standaloneNodes.length > 0"
        :title="`[${t('node.standalone')}] — ${filteredNodes(subStore.standaloneNodes).length} nodes`"
        name="standalone"
        class="sub-group"
      >
        <div class="table-scroll">
        <el-table :data="filteredNodes(subStore.standaloneNodes)" stripe style="width: 100%" size="small" :row-class-name="rowClassName">
          <el-table-column prop="name" :label="t('node.name')" min-width="160" show-overflow-tooltip />
          <el-table-column :label="t('node.address')" min-width="180">
            <template #default="{ row }"><span class="mono">{{ row.address }}:{{ row.port }}</span></template>
          </el-table-column>
          <el-table-column prop="protocol" :label="t('node.protocol')" width="100">
            <template #default="{ row }"><span class="protocol-tag">{{ row.protocol }}</span></template>
          </el-table-column>
          <el-table-column :label="t('node.latency')" width="120" align="center" sortable>
            <template #default="{ row }">
              <LatencyTag :latency="latencyMap[row.name]" />
            </template>
          </el-table-column>
          <el-table-column :label="t('common.edit')" width="180" align="center">
            <template #default="{ row }">
              <template v-if="proxyStore.status.running && proxyStore.status.node === row.name">
                <el-tag type="success" size="small" effect="dark">{{ t('node.current') }}</el-tag>
                <el-button type="danger" size="small" @click="handleStop" style="margin-left:4px">
                  {{ t('node.disconnect') }}
                </el-button>
              </template>
              <template v-else-if="proxyStore.status.running">
                <el-button type="warning" size="small" @click="handleSwitch(row.name)">
                  {{ t('node.switchTo') }}
                </el-button>
              </template>
              <template v-else>
                <el-button type="primary" size="small" @click="handleConnect(row.name)">
                  {{ t('node.connect') }}
                </el-button>
              </template>
              <el-button type="danger" size="small" plain @click="handleDeleteNode(row.name)">
                {{ t('common.delete') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        </div>
      </el-collapse-item>
    </el-collapse>

    <AddNodeDialog
      v-model="addDialogVisible"
      @added="handleNodeAdded"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, h, defineComponent } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
import { useProxyStore } from '@/stores/proxy'
import { useSubscriptionStore } from '@/stores/subscription'
import { useWebSocket } from '@/composables/useWebSocket'
import { useApi } from '@/composables/useApi'
import AddNodeDialog from './AddNodeDialog.vue'

const LatencyTag = defineComponent({
  props: { latency: { type: Number, default: undefined } },
  setup(props) {
    return () => {
      if (props.latency === undefined || props.latency === null) return h('span', { style: 'color:#555;' }, '—')
      if (props.latency < 0) return h(ElTag, { type: 'danger', size: 'small' }, () => '✕')
      const type = props.latency < 200 ? 'success' : props.latency < 500 ? 'warning' : 'danger'
      return h(ElTag, { type, size: 'small' }, () => `${props.latency}ms`)
    }
  },
})

const { t } = useI18n()
const proxyStore = useProxyStore()
const subStore = useSubscriptionStore()
const ws = useWebSocket()
const api = useApi()

const regions = ref<string[]>([])
const selectedRegion = ref('')
const testLoading = ref(false)
const testProgress = ref('')
const latencyMap = ref<Record<string, number>>({})
const addDialogVisible = ref(false)
const activeGroups = ref<string[]>([])

const rowClassName = ({ row }: { row: any }) => {
  if (proxyStore.status.running && proxyStore.status.node === row.name) {
    return 'current-node-row'
  }
  return ''
}

const filteredNodes = (nodes: any[]) => {
  if (!selectedRegion.value) return nodes
  return nodes.filter(n => n.region === selectedRegion.value)
}

const loadRegions = async () => {
  try {
    const res = await api.get('/api/nodes/regions')
    regions.value = res.regions || (Array.isArray(res) ? res : Object.keys(res))
  } catch {}
}

const handleTestLatency = async () => {
  testLoading.value = true
  testProgress.value = ''
  try {
    const body: any = {}
    if (selectedRegion.value) body.region = selectedRegion.value
    const res = await api.post('/api/proxy/test', body)
    const results = Array.isArray(res) ? res : res.results
    if (results) {
      const map: Record<string, number> = {}
      for (const r of results) {
        map[r.name] = r.latency
      }
      latencyMap.value = map
    }
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    testLoading.value = false
    testProgress.value = ''
  }
}

const handleConnect = async (nodeName: string) => {
  try {
    const res = await api.post('/api/proxy/start', { node_name: nodeName })
    proxyStore.updateStatus(res)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  }
}

const handleSwitch = async (nodeName: string) => {
  try {
    const res = await api.post('/api/proxy/start', { node_name: nodeName })
    proxyStore.updateStatus(res)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  }
}

const handleStop = async () => {
  try {
    const res = await api.post('/api/proxy/stop')
    proxyStore.updateStatus(res)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  }
}

const handleDeleteNode = async (name: string) => {
  try {
    await ElMessageBox.confirm(t('node.deleteConfirm', { name }), t('common.confirm'), {
      type: 'warning',
    })
    await api.del(`/api/nodes/${encodeURIComponent(name)}`)
    ElMessage.success(t('common.success'))
    await subStore.loadConfig()
  } catch {}
}

const handleNodeAdded = async () => {
  await subStore.loadConfig()
}

// Listen for latency progress via WebSocket
const wsCleanup = ws.onMessage((data: any) => {
  if (data.type === 'latency_progress') {
    if (data.status === 'started') {
      testProgress.value = t('node.testingProgress', { done: 0, total: data.total })
    } else if (data.status === 'completed') {
      testProgress.value = ''
    }
  }
})

onMounted(async () => {
  loadRegions()
  await subStore.loadConfig()
  // Default expand first subscription group
  if (subStore.subscriptions.length > 0) {
    activeGroups.value = [subStore.subscriptions[0].name]
  } else if (subStore.standaloneNodes.length > 0) {
    activeGroups.value = ['standalone']
  }
})

onUnmounted(() => {
  if (wsCleanup) wsCleanup()
})
</script>

<style scoped>
.node-table {
  margin-top: 0;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.region-select {
  width: 200px;
}
.test-btn, .add-btn {
  font-size: 13px;
}
.test-progress {
  font-size: 12px;
  color: var(--geek-text-secondary);
}
.flex-grow {
  flex-grow: 1;
}
.node-collapse {
  border: 1px solid var(--geek-border);
  border-radius: 8px;
  overflow: hidden;
}
.sub-group :deep(.el-collapse-item__header) {
  font-size: 14px;
  padding-left: 16px;
}
.mono {
  font-size: 13px;
  color: var(--geek-text-secondary);
}
.protocol-tag {
  font-size: 12px;
  color: var(--geek-text-secondary);
  background: rgba(255, 255, 255, 0.06);
  padding: 2px 8px;
  border-radius: 4px;
}
.table-scroll {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}
.node-table :deep(.current-node-row) {
  background-color: rgba(82, 196, 26, 0.08) !important;
}
.node-table :deep(.current-node-row td) {
  font-weight: 500;
}
@media (max-width: 768px) {
  .toolbar {
    flex-wrap: wrap;
  }
  .region-select {
    width: 100%;
  }
  .flex-grow {
    display: none;
  }
}
</style>
