import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useApi } from '@/composables/useApi'

export const useSubscriptionStore = defineStore('subscription', () => {
  const subscriptions = ref<any[]>([])
  const standaloneNodes = ref<any[]>([])
  const api = useApi()

  const loadConfig = async () => {
    const res = await api.get('/api/config')
    subscriptions.value = res.subscriptions || []
    standaloneNodes.value = res.standalone_nodes || []
  }

  const allNodes = computed(() => {
    const nodes: any[] = []
    for (const sub of subscriptions.value) {
      for (const node of sub.nodes || []) {
        nodes.push({ ...node, _source: sub.name, _type: 'subscription' })
      }
    }
    for (const node of standaloneNodes.value) {
      nodes.push({ ...node, _source: 'standalone', _type: 'standalone' })
    }
    return nodes
  })

  return { subscriptions, standaloneNodes, loadConfig, allNodes }
})
