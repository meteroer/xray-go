import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref('')

  const isLoggedIn = computed(() => !!token.value)

  const setToken = (newToken: string, user?: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
    if (user) username.value = user
  }

  const clearToken = () => {
    token.value = ''
    username.value = ''
    localStorage.removeItem('token')
  }

  return { token, username, isLoggedIn, setToken, clearToken }
})
