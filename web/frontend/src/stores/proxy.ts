import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface ProxyStatus {
  running: boolean
  node: string
  http_port: number
  socks_port: number
  route_mode: string
}

export const useProxyStore = defineStore('proxy', () => {
  const status = ref<ProxyStatus>({
    running: false,
    node: '',
    http_port: 0,
    socks_port: 0,
    route_mode: '',
  })

  const updateStatus = (data: Partial<ProxyStatus>) => {
    status.value = { ...status.value, ...data }
  }

  return { status, updateStatus }
})
