<template>
  <div class="node-table">
    <div class="toolbar">
      <el-select v-model="selectedRegion" :placeholder="t('node.region')" clearable style="width: 200px">
        <el-option :label="t('node.allRegions')" value="" />
        <el-option v-for="r in regions" :key="r" :label="r" :value="r" />
      </el-select>
      <el-button type="primary" :loading="testLoading" @click="handleTestLatency">
        {{ t('node.testLatency') }}
      </el-button>
      <div class="flex-grow" />
      <el-button type="success" @click="addDialogVisible = true">{{ t('node.add') }}</el-button>
    </div>

    <el-collapse v-model="activeGroups">
      <el-collapse-item
        v-for="sub in subStore.subscriptions"
        :key="sub.name"
        :title="`${sub.name} (${t('node.nodesCount', { count: filteredNodes(sub.nodes).length })})`"
        :name="sub.name"
      >
        <el-table :data="filteredNodes(sub.nodes)" stripe style="width: 100%" size="small">
          <el-table-column prop="name" :label="t('node.name')" min-width="160" show-overflow-tooltip />
          <el-table-column :label="t('node.address')" min-width="180">
            <template #default="{ row }">{{ row.address }}:{{ row.port }}</template>
          </el-table-column>
          <el-table-column prop="protocol" :label="t('node.protocol')" width="100" />
          <el-table-column :label="t('node.latency')" width="120" align="center">
            <template #default="{ row }">
              <LatencyTag :latency="latencyMap[row.name]" />
            </template>
          </el-table-column>
          <el-table-column :label="t('common.edit')" width="100" align="center">
            <template #default="{ row }">
              <el-button
                v-if="proxyStore.status.running && proxyStore.status.node === row.name"
                type="danger"
                size="small"
                @click="handleStop"
              >
                {{ t('node.disconnect') }}
              </el-button>
              <el-button v-else type="primary" size="small" @click="handleConnect(row.name)">
                {{ t('node.connect') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>

      <el-collapse-item
        v-if="subStore.standaloneNodes.length > 0"
        :title="`${t('node.standalone')} (${t('node.nodesCount', { count: filteredNodes(subStore.standaloneNodes).length })})`"
        name="standalone"
      >
        <el-table :data="filteredNodes(subStore.standaloneNodes)" stripe style="width: 100%" size="small">
          <el-table-column prop="name" :label="t('node.name')" min-width="160" show-overflow-tooltip />
          <el-table-column :label="t('node.address')" min-width="180">
            <template #default="{ row }">{{ row.address }}:{{ row.port }}</template>
          </el-table-column>
          <el-table-column prop="protocol" :label="t('node.protocol')" width="100" />
          <el-table-column :label="t('node.latency')" width="120" align="center">
            <template #default="{ row }">
              <LatencyTag :latency="latencyMap[row.name]" />
            </template>
          </el-table-column>
          <el-table-column :label="t('common.edit')" width="160" align="center">
            <template #default="{ row }">
              <el-button
                v-if="proxyStore.status.running && proxyStore.status.node === row.name"
                type="danger"
                size="small"
                @click="handleStop"
              >
                {{ t('node.disconnect') }}
              </el-button>
              <el-button v-else type="primary" size="small" @click="handleConnect(row.name)">
                {{ t('node.connect') }}
              </el-button>
              <el-button type="danger" size="small" plain @click="handleDeleteNode(row.name)">
                {{ t('common.delete') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>
    </el-collapse>

    <AddNodeDialog
      v-model="addDialogVisible"
      @added="handleNodeAdded"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h, defineComponent } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
import { useProxyStore } from '@/stores/proxy'
import { useSubscriptionStore } from '@/stores/subscription'
import { useApi } from '@/composables/useApi'
import AddNodeDialog from './AddNodeDialog.vue'

const LatencyTag = defineComponent({
  props: { latency: { type: Number, default: undefined } },
  setup(props) {
    return () => {
      if (props.latency === undefined || props.latency === null) return h('span', { style: 'color:#c0c4cc' }, '—')
      if (props.latency < 0) return h(ElTag, { type: 'danger', size: 'small' }, () => '✕')
      const type = props.latency < 200 ? 'success' : props.latency < 500 ? 'warning' : 'danger'
      return h(ElTag, { type, size: 'small' }, () => `${props.latency}ms`)
    }
  },
})

const { t } = useI18n()
const proxyStore = useProxyStore()
const subStore = useSubscriptionStore()
const api = useApi()

const regions = ref<string[]>([])
const selectedRegion = ref('')
const testLoading = ref(false)
const latencyMap = ref<Record<string, number>>({})
const addDialogVisible = ref(false)
const activeGroups = ref<string[]>([])

const filteredNodes = (nodes: any[]) => {
  if (!selectedRegion.value) return nodes
  return nodes.filter(n => n.region === selectedRegion.value)
}

const loadRegions = async () => {
  try {
    const res = await api.get('/api/nodes/regions')
    regions.value = res.regions || res || []
  } catch {}
}

const handleTestLatency = async () => {
  testLoading.value = true
  try {
    const body: any = {}
    if (selectedRegion.value) body.region = selectedRegion.value
    const res = await api.post('/api/proxy/test', body)
    if (res.results) {
      const map: Record<string, number> = {}
      for (const r of res.results) {
        map[r.name] = r.latency
      }
      latencyMap.value = map
    }
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    testLoading.value = false
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

onMounted(() => {
  loadRegions()
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
.flex-grow {
  flex-grow: 1;
}
</style>
