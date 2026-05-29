<template>
  <el-dialog
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    :title="t('sub.edit')"
    width="520px"
    class="edit-sub-dialog"
  >
    <el-form :model="form" @submit.prevent="handleSubmit">
      <el-form-item :label="t('sub.name')">
        <el-input v-model="form.name" disabled />
      </el-form-item>
      <el-form-item :label="t('sub.url')">
        <el-input v-model="form.url" placeholder="https://example.com/sub" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">{{ t('common.save') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useApi } from '@/composables/useApi'

const props = defineProps<{
  modelValue: boolean
  subscription: { name: string; url: string } | null
}>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'saved': []
}>()

const { t } = useI18n()
const api = useApi()
const loading = ref(false)
const form = ref({ name: '', url: '' })

watch(() => props.modelValue, (val) => {
  if (val && props.subscription) {
    form.value = { name: props.subscription.name, url: props.subscription.url }
  }
})

const handleSubmit = async () => {
  if (!form.value.url.trim()) return
  loading.value = true
  try {
    await api.put(`/api/subscriptions/${encodeURIComponent(form.value.name)}`, { url: form.value.url.trim() })
    ElMessage.success(t('common.success'))
    emit('update:modelValue', false)
    emit('saved')
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    loading.value = false
  }
}
</script>

<style>
@media (max-width: 768px) {
  .edit-sub-dialog .el-dialog {
    width: 92vw !important;
    margin: 4vh auto !important;
  }
}
</style>
