<template>
  <div class="layout">
    <TopNavBar />
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import TopNavBar from '@/components/TopNavBar.vue'
import { useWebSocket } from '@/composables/useWebSocket'

const ws = useWebSocket()

onMounted(() => {
  ws.connect()
})

onUnmounted(() => {
  ws.disconnect()
})
</script>

<style scoped>
.layout {
  min-height: 100vh;
  background-color: var(--geek-bg);
}
.main-content {
  padding: 24px;
  position: relative;
  z-index: 1;
}
@media (max-width: 768px) {
  .main-content {
    padding: 16px 12px;
  }
}
</style>
