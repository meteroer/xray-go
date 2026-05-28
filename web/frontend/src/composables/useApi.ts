import { useRouter } from 'vue-router'

export function useApi() {
  const router = useRouter()

  const getToken = () => localStorage.getItem('token') || ''

  const request = async (url: string, options: RequestInit = {}) => {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string> || {}),
    }
    const token = getToken()
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const response = await fetch(url, { ...options, headers })

    if (response.status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
      throw new Error('Unauthorized')
    }

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      throw new Error(data.message || response.statusText)
    }

    return response.json()
  }

  const get = (url: string) => request(url)
  const post = (url: string, data?: any) => request(url, { method: 'POST', body: JSON.stringify(data) })
  const put = (url: string, data?: any) => request(url, { method: 'PUT', body: JSON.stringify(data) })
  const del = (url: string) => request(url, { method: 'DELETE' })

  return { get, post, put, del }
}
