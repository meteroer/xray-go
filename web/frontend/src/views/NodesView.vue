<template>
  <div class="nodes-page">
    <ProxyControl />
    <NodeTable />
  </div>
</template>

<script setup lang="ts">
import ProxyControl from '@/components/ProxyControl.vue'
import NodeTable from '@/components/NodeTable.vue'
</script>

<style scoped>
.nodes-page {
  max-width: 1200px;
  margin: 0 auto;
}
</style>
