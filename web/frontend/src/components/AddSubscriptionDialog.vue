<template>
  <el-dialog
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    :title="t('sub.add')"
    width="500px"
  >
    <el-form :model="form" @submit.prevent="handleSubmit">
      <el-form-item :label="t('sub.name')">
        <el-input v-model="form.name" />
      </el-form-item>
      <el-form-item :label="t('sub.url')">
        <el-input v-model="form.url" />
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
const loading = ref(false)
const form = ref({ name: '', url: '' })

const handleSubmit = async () => {
  if (!form.value.name || !form.value.url) return
  loading.value = true
  try {
    await api.post('/api/subscriptions', form.value)
    ElMessage.success(t('common.success'))
    form.value = { name: '', url: '' }
    emit('update:modelValue', false)
    emit('added')
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    loading.value = false
  }
}
</script>
