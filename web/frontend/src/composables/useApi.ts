import { useAuthStore } from '@/stores/auth'

let _showSessionExpired: (() => void) | null = null

export function registerSessionExpiredHandler(fn: () => void) {
  _showSessionExpired = fn
}

export function useApi() {
  const authStore = useAuthStore()

  const request = async (url: string, options: RequestInit = {}) => {
    const headers: Record<string, string> = {
      ...(options.headers as Record<string, string> || {}),
    }
    const token = authStore.token
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    if (options.body && typeof options.body === 'string') {
      headers['Content-Type'] = 'application/json'
    }

    const res = await fetch(url, { ...options, headers })

    if (res.status === 401) {
      if (_showSessionExpired) _showSessionExpired()
      authStore.clearToken()
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }

    if (!res.ok) {
      let errorMsg = ''
      try {
        const data = await res.json()
        errorMsg = data.error || data.message || ''
      } catch {}
      throw new Error(errorMsg || `HTTP ${res.status}`)
    }

    return res.json()
  }

  return {
    get: (url: string) => request(url),
    post: (url: string, body?: any) => request(url, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
    put: (url: string, body?: any) => request(url, { method: 'PUT', body: body ? JSON.stringify(body) : undefined }),
    del: (url: string) => request(url, { method: 'DELETE' }),
  }
}