<template>
  <el-dialog
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    :title="t('node.add')"
    width="520px"
    class="add-node-dialog"
  >
    <el-form @submit.prevent="handleSubmit">
      <el-form-item :label="t('node.addHint')">
        <el-input
          v-model="link"
          type="textarea"
          :rows="6"
          placeholder="vmess://..., vless://..., trojan://..., ss://..."
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">{{ t('common.confirm') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useApi } from '@/composables/useApi'

defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'added': []
}>()

const { t } = useI18n()
const api = useApi()
const link = ref('')
const loading = ref(false)

const handleSubmit = async () => {
  if (!link.value.trim()) return
  loading.value = true
  try {
    await api.post('/api/nodes', { link: link.value.trim() })
    ElMessage.success(t('common.success'))
    link.value = ''
    emit('update:modelValue', false)
    emit('added')
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    loading.value = false
  }
}
</script>

<style>
@media (max-width: 768px) {
  .add-node-dialog .el-dialog {
    width: 92vw !important;
    margin: 4vh auto !important;
  }
}
</style>
