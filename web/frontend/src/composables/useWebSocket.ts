import { useProxyStore } from '@/stores/proxy'

type MessageHandler = (data: any) => void

const handlers: MessageHandler[] = []
let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectDelay = 1000

function doConnect() {
  const token = localStorage.getItem('token')
  if (!token) return

  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${protocol}//${location.host}/api/ws?token=${token}`

  ws = new WebSocket(url)

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.type === 'proxy_status') {
        const proxyStore = useProxyStore()
        proxyStore.updateStatus(data)
      }
      for (const handler of handlers) {
        handler(data)
      }
    } catch {}
  }

  ws.onclose = () => {
    reconnectTimer = setTimeout(() => {
      reconnectDelay = Math.min(reconnectDelay * 2, 30000)
      doConnect()
    }, reconnectDelay)
  }
}

function doDisconnect() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (ws) {
    ws.close()
    ws = null
  }
  reconnectDelay = 1000
}

export function useWebSocket() {
  const connect = () => doConnect()
  const disconnect = () => doDisconnect()

  const onMessage = (handler: MessageHandler) => {
    handlers.push(handler)
    return () => {
      const idx = handlers.indexOf(handler)
      if (idx >= 0) handlers.splice(idx, 1)
    }
  }

  return { connect, disconnect, onMessage }
}