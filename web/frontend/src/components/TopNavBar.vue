<template>
  <div class="top-nav">
    <div class="nav-bar">
      <div class="logo-item" @click="router.push('/')">
        <span class="logo">XRAY-GO</span>
        <span
          class="status-dot"
          :class="proxyStore.status.running ? 'running' : 'stopped'"
        />
        <span class="status-text">{{ proxyStore.status.running ? 'ONLINE' : 'OFFLINE' }}</span>
      </div>
      <div class="flex-grow" />
      <button class="hamburger" @click="menuOpen = !menuOpen" :class="{ active: menuOpen }">
        <span /><span /><span />
      </button>
    </div>

    <!-- Desktop menu -->
    <el-menu mode="horizontal" :ellipsis="false" class="desktop-menu">
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

    <!-- Mobile drawer -->
    <transition name="slide">
      <div v-if="menuOpen" class="mobile-drawer" @click.self="menuOpen = false">
        <div class="drawer-content">
          <div class="drawer-header">
            <span class="logo">XRAY-GO</span>
            <span
              class="status-dot"
              :class="proxyStore.status.running ? 'running' : 'stopped'"
            />
            <span class="status-text">{{ proxyStore.status.running ? 'ONLINE' : 'OFFLINE' }}</span>
            <div class="flex-grow" />
            <button class="close-btn" @click="menuOpen = false">✕</button>
          </div>
          <div class="drawer-links">
            <div class="drawer-link" @click="navigate('/')">
              {{ t('nav.nodes') }}
            </div>
            <div class="drawer-link" @click="navigate('/subscription')">
              {{ t('nav.subscription') }}
            </div>
            <div class="drawer-link" @click="navigate('/routing')">
              {{ t('nav.routing') }}
            </div>
            <div class="drawer-link" @click="navigate('/settings')">
              {{ t('nav.settings') }}
            </div>
            <div class="drawer-divider" />
            <div class="drawer-link" @click="toggleLang(); menuOpen = false">
              {{ locale === 'zh' ? 'English' : '中文' }}
            </div>
            <div class="drawer-link logout" @click="handleLogout">
              {{ t('nav.logout') }}
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { User } from '@element-plus/icons-vue'
import { useProxyStore } from '@/stores/proxy'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const { t, locale } = useI18n()
const proxyStore = useProxyStore()
const authStore = useAuthStore()
const menuOpen = ref(false)

const toggleLang = () => {
  const newLang = locale.value === 'zh' ? 'en' : 'zh'
  locale.value = newLang
  localStorage.setItem('lang', newLang)
}

const handleLogout = () => {
  menuOpen.value = false
  authStore.clearToken()
  router.push('/login')
}

const navigate = (path: string) => {
  menuOpen.value = false
  if (route.path !== path) {
    router.push(path)
  }
}
</script>

<style scoped>
.top-nav {
  position: relative;
  z-index: 10;
}

/* Mobile nav bar - hidden on desktop */
.nav-bar {
  display: none;
}

/* Desktop menu */
.desktop-menu {
  padding: 0 24px;
  background: var(--geek-bg-secondary) !important;
  border-bottom: 1px solid var(--geek-border) !important;
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

/* Mobile hamburger */
.hamburger {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 5px;
  width: 40px;
  height: 40px;
  background: none;
  border: 1px solid var(--geek-border);
  border-radius: 6px;
  cursor: pointer;
  padding: 8px;
}
.hamburger span {
  display: block;
  width: 20px;
  height: 2px;
  background: var(--geek-text);
  border-radius: 1px;
  transition: all 0.3s;
}
.hamburger.active span:nth-child(1) {
  transform: rotate(45deg) translate(5px, 5px);
}
.hamburger.active span:nth-child(2) {
  opacity: 0;
}
.hamburger.active span:nth-child(3) {
  transform: rotate(-45deg) translate(5px, -5px);
}

/* Mobile drawer */
.mobile-drawer {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 100;
}
.drawer-content {
  position: absolute;
  top: 0;
  right: 0;
  width: 260px;
  max-width: 80vw;
  height: 100%;
  background: var(--geek-bg-secondary);
  box-shadow: -4px 0 16px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
}
.drawer-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--geek-border);
}
.close-btn {
  background: none;
  border: none;
  font-size: 18px;
  color: var(--geek-text-secondary);
  cursor: pointer;
  padding: 4px 8px;
}
.drawer-links {
  padding: 8px 0;
}
.drawer-link {
  padding: 14px 24px;
  font-size: 15px;
  color: var(--geek-text);
  cursor: pointer;
  transition: background 0.2s;
}
.drawer-link:active {
  background: rgba(0, 0, 0, 0.06);
}
.drawer-link.logout {
  color: var(--geek-danger);
}
.drawer-divider {
  height: 1px;
  background: var(--geek-border);
  margin: 8px 20px;
}

/* Slide transition */
.slide-enter-active,
.slide-leave-active {
  transition: opacity 0.25s ease;
}
.slide-enter-active .drawer-content,
.slide-leave-active .drawer-content {
  transition: transform 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
}
.slide-enter-from .drawer-content,
.slide-leave-to .drawer-content {
  transform: translateX(100%);
}

/* Mobile responsive */
@media (max-width: 768px) {
  .nav-bar {
    display: flex;
    align-items: center;
    padding: 0 16px;
    height: 52px;
    background: var(--geek-bg-secondary);
    border-bottom: 1px solid var(--geek-border);
  }
  .desktop-menu {
    display: none !important;
  }
  .mobile-drawer {
    display: block;
  }
}
</style>
