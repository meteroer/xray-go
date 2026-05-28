<template>
  <div class="routing-page">
    <h2>{{ t('routing.title') }}</h2>

    <el-card shadow="never" style="margin-top: 16px">
      <el-form label-width="120px">
        <el-form-item :label="t('routing.mode')">
          <el-radio-group v-model="routeMode">
            <el-radio value="global">{{ t('routing.global') }}</el-radio>
            <el-radio value="whitelist">{{ t('routing.whitelist') }}</el-radio>
            <el-radio value="blacklist">{{ t('routing.blacklist') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="routeMode === 'whitelist'" :label="t('routing.rules')">
          <p class="mode-hint">{{ t('routing.whitelistHint') }}</p>
          <RouteRuleEditor :rules="whitelist" @update:rules="whitelist = $event" />
        </el-form-item>

        <el-form-item v-if="routeMode === 'blacklist'" :label="t('routing.rules')">
          <p class="mode-hint">{{ t('routing.blacklistHint') }}</p>
          <RouteRuleEditor :rules="blacklist" @update:rules="blacklist = $event" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSave">{{ t('common.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useApi } from '@/composables/useApi'
import RouteRuleEditor from '@/components/RouteRuleEditor.vue'

const { t } = useI18n()
const api = useApi()

const routeMode = ref('global')
const whitelist = ref<string[]>([])
const blacklist = ref<string[]>([])
const saving = ref(false)

const handleSave = async () => {
  saving.value = true
  try {
    await api.put('/api/settings/route-mode', { mode: routeMode.value })
    if (routeMode.value === 'whitelist') {
      await api.put('/api/settings/whitelist', { rules: whitelist.value })
    } else if (routeMode.value === 'blacklist') {
      await api.put('/api/settings/blacklist', { rules: blacklist.value })
    }
    ElMessage.success(t('routing.saveSuccess'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    const [modeRes, wlRes, blRes] = await Promise.all([
      api.get('/api/settings/route-mode'),
      api.get('/api/settings/whitelist'),
      api.get('/api/settings/blacklist'),
    ])
    routeMode.value = modeRes.mode || 'global'
    whitelist.value = wlRes.rules || wlRes.whitelist || []
    blacklist.value = blRes.rules || blRes.blacklist || []
  } catch {}
})
</script>

<style scoped>
.routing-page {
  max-width: 1200px;
  margin: 0 auto;
}
.mode-hint {
  color: #909399;
  font-size: 13px;
  margin: 0 0 12px 0;
}
</style>
