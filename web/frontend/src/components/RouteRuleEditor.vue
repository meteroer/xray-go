<template>
  <div class="route-rule-editor">
    <div class="rule-tags">
      <el-tag
        v-for="(rule, index) in rules"
        :key="index"
        closable
        @close="removeRule(index)"
        class="rule-tag"
        effect="dark"
      >
        {{ rule }}
      </el-tag>
      <el-tag v-if="rules.length === 0" type="info" effect="dark">{{ t('common.noData') }}</el-tag>
    </div>
    <div v-if="showInput" class="rule-input">
      <el-input
        v-model="newRule"
        size="small"
        :placeholder="t('routing.addRule')"
        @keyup.enter="addRule"
        class="rule-input-field"
      >
        <template #append>
          <el-button @click="addRule" :disabled="!newRule.trim()">{{ t('common.confirm') }}</el-button>
        </template>
      </el-input>
    </div>
    <el-button v-else size="small" class="add-rule-btn" @click="showInput = true">+ {{ t('routing.addRule') }}</el-button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ rules: string[] }>()
const emit = defineEmits<{
  'update:rules': [value: string[]]
}>()

const { t } = useI18n()
const showInput = ref(false)
const newRule = ref('')

const addRule = () => {
  const rule = newRule.value.trim()
  if (!rule) return
  emit('update:rules', [...props.rules, rule])
  newRule.value = ''
  showInput.value = false
}

const removeRule = (index: number) => {
  const updated = [...props.rules]
  updated.splice(index, 1)
  emit('update:rules', updated)
}
</script>

<style scoped>
.route-rule-editor {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--geek-border);
  border-radius: 6px;
  padding: 16px;
}
.rule-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}
.rule-tag {
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 12px;
}
.rule-input {
  margin-top: 8px;
}
.rule-input-field {
  max-width: 400px;
}
.add-rule-btn {
  font-size: 12px;
}
</style>
