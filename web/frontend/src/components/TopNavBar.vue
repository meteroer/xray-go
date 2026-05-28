<template>
  <el-menu mode="horizontal" :ellipsis="false" class="top-nav">
    <el-menu-item class="logo-item">
      <span class="logo">Xray-Go</span>
      <el-badge is-dot :type="proxyStore.status.running ? 'success' : 'danger'" class="status-badge" />
    </el-menu-item>
    <el-menu-item index="/" @click="router.push('/')">
      {{ t('nav.nodes') }}
    </el-menu-item>
    <el-menu-item index="/subscription" @click="router.push('/subscription')">
      {{ t('nav.subscription') }}
    </el-menu-item>
    <el-menu-item index="/routing" @click="router.push('/routing')">
      {{ t('nav.routing') }}
    </el-menu-item>
    <el-menu-item index="/settings" @click="router.push('/settings')">
      {{ t('nav.settings') }}
    </el-menu-item>
    <div class="flex-grow" />
    <el-menu-item @click="toggleLang">
      {{ locale === 'zh' ? 'EN' : '中文' }}
    </el-menu-item>
    <el-sub-menu index="user">
      <template #title>
        <el-icon><User /></el-icon>
      </template>
      <el-menu-item @click="handleLogout">
        {{ t('nav.logout') }}
      </el-menu-item>
    </el-sub-menu>
  </el-menu>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { User } from '@element-plus/icons-vue'
import { useProxyStore } from '@/stores/proxy'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const { t, locale } = useI18n()
const proxyStore = useProxyStore()
const authStore = useAuthStore()

const toggleLang = () => {
  const newLang = locale.value === 'zh' ? 'en' : 'zh'
  locale.value = newLang
  localStorage.setItem('lang', newLang)
}

const handleLogout = () => {
  authStore.clearToken()
  router.push('/login')
}
</script>

<style scoped>
.top-nav {
  padding: 0 20px;
}
.logo-item {
  font-weight: bold;
}
.logo {
  font-size: 18px;
  margin-right: 8px;
}
.status-badge {
  margin-top: -8px;
}
.flex-grow {
  flex-grow: 1;
}
</style>
