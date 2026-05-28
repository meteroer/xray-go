<template>
  <div class="nodes-page">
    <h2 class="geek-title">{{ t('node.title') }}</h2>
    <ProxyControl />
    <NodeTable />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import ProxyControl from '@/components/ProxyControl.vue'
import NodeTable from '@/components/NodeTable.vue'

const { t } = useI18n()
</script>

<style scoped>
.nodes-page {
  max-width: 1200px;
  margin: 0 auto;
}
</style>
