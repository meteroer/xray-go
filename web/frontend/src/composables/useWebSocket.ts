import { ref } from 'vue'
import { useProxyStore } from '@/stores/proxy'

export function useWebSocket() {
  const ws = ref<WebSocket | null>(null)
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectDelay = 1000

  const connect = () => {
    const token = localStorage.getItem('token')
    if (!token) return

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${protocol}//${location.host}/api/ws?token=${token}`

    ws.value = new WebSocket(url)

    ws.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'proxy_status') {
          const proxyStore = useProxyStore()
          proxyStore.updateStatus(data)
        }
      } catch {}
    }

    ws.value.onclose = () => {
      reconnectTimer = setTimeout(() => {
        reconnectDelay = Math.min(reconnectDelay * 2, 30000)
        connect()
      }, reconnectDelay)
    }
  }

  const disconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    reconnectDelay = 1000
  }

  return { connect, disconnect }
}
