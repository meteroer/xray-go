<template>
  <el-menu mode="horizontal" :ellipsis="false" class="top-nav">
    <el-menu-item class="logo-item" index="logo">
      <span class="logo">XRAY-GO</span>
      <span
        class="status-dot"
        :class="proxyStore.status.running ? 'running' : 'stopped'"
      />
      <span class="status-text">{{ proxyStore.status.running ? 'ONLINE' : 'OFFLINE' }}</span>
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
      <span class="lang-toggle">{{ locale === 'zh' ? 'EN' : '中文' }}</span>
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
  padding: 0 24px;
  background: var(--geek-bg-secondary) !important;
  border-bottom: 1px solid var(--geek-border) !important;
  position: relative;
  z-index: 10;
}
.logo-item {
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 10px;
}
.logo {
  font-size: 16px;
  font-weight: 600;
  color: var(--geek-text);
}
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
.status-dot.running {
  background-color: #52c41a;
}
.status-dot.stopped {
  background-color: var(--geek-danger);
}
.status-text {
  font-size: 11px;
  color: var(--geek-text-secondary);
}
.lang-toggle {
  font-size: 13px;
  color: var(--geek-text-secondary);
}
.flex-grow {
  flex-grow: 1;
}
</style>
