<template>
  <el-form
    ref="formRef"
    :model="formData"
    :rules="rules"
    label-width="120px"
    @submit.prevent="handleSubmit"
  >
    <template v-for="(field, key) in sortedFields" :key="key">
      <!-- 字符串输入 -->
      <el-form-item
        v-if="field.type === 'string' && field.format !== 'password'"
        :label="field.title || key"
        :prop="key"
      >
        <el-input
          v-if="!field.ui?.widget || field.ui.widget === 'input'"
          v-model="formData[key]"
          :placeholder="field.ui?.placeholder"
          clearable
        />
        <el-input
          v-else-if="field.ui.widget === 'textarea'"
          v-model="formData[key]"
          type="textarea"
          :rows="4"
          :placeholder="field.ui?.placeholder"
        />
        <el-select
          v-else-if="field.ui.widget === 'select'"
          v-model="formData[key]"
          :placeholder="field.ui?.placeholder"
        >
          <el-option
            v-for="option in field.enum"
            :key="option"
            :label="option"
            :value="option"
          />
        </el-select>
        <template v-if="field.description" #help>
          <div class="field-help">{{ field.description }}</div>
        </template>
      </el-form-item>

      <!-- 密码输入 -->
      <el-form-item
        v-else-if="field.type === 'string' && field.format === 'password'"
        :label="field.title || key"
        :prop="key"
      >
        <el-input
          v-model="formData[key]"
          type="password"
          show-password
          :placeholder="field.ui?.placeholder"
        />
        <template v-if="field.description" #help>
          <div class="field-help">{{ field.description }}</div>
        </template>
      </el-form-item>

      <!-- 数字输入 -->
      <el-form-item
        v-else-if="field.type === 'number'"
        :label="field.title || key"
        :prop="key"
      >
        <el-input-number
          v-if="!field.ui?.widget || field.ui.widget === 'input'"
          v-model="formData[key]"
          :min="field.minimum"
          :max="field.maximum"
        />
        <el-slider
          v-else-if="field.ui.widget === 'slider'"
          v-model="formData[key]"
          :min="field.minimum"
          :max="field.maximum"
        />
        <template v-if="field.description" #help>
          <div class="field-help">{{ field.description }}</div>
        </template>
      </el-form-item>

      <!-- 布尔值 -->
      <el-form-item
        v-else-if="field.type === 'boolean'"
        :label="field.title || key"
        :prop="key"
      >
        <el-switch
          v-if="!field.ui?.widget || field.ui.widget === 'switch'"
          v-model="formData[key]"
        />
        <el-checkbox
          v-else-if="field.ui.widget === 'checkbox'"
          v-model="formData[key]"
        >
          {{ field.description }}
        </el-checkbox>
        <template v-if="field.description && field.ui?.widget !== 'checkbox'" #help>
          <div class="field-help">{{ field.description }}</div>
        </template>
      </el-form-item>

      <!-- 颜色选择 -->
      <el-form-item
        v-else-if="field.format === 'color'"
        :label="field.title || key"
        :prop="key"
      >
        <el-color-picker v-model="formData[key]" />
        <template v-if="field.description" #help>
          <div class="field-help">{{ field.description }}</div>
        </template>
      </el-form-item>

      <!-- 日期选择 -->
      <el-form-item
        v-else-if="field.format === 'date'"
        :label="field.title || key"
        :prop="key"
      >
        <el-date-picker v-model="formData[key]" type="date" />
        <template v-if="field.description" #help>
          <div class="field-help">{{ field.description }}</div>
        </template>
      </el-form-item>
    </template>

    <el-form-item>
      <el-button type="primary" @click="handleSubmit">保存</el-button>
      <el-button @click="handleReset">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { PluginConfigSchema, ConfigFieldSchema } from 'runixo-plugin-types'

interface Props {
  schema: PluginConfigSchema
  modelValue: Record<string, any>
}

interface Emits {
  (e: 'update:modelValue', value: Record<string, any>): void
  (e: 'submit', value: Record<string, any>): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const formRef = ref()
const formData = ref<Record<string, any>>({})

// 按order排序字段
const sortedFields = computed(() => {
  const fields = Object.entries(props.schema.properties)
  return fields.sort((a, b) => {
    const orderA = a[1].ui?.order ?? 999
    const orderB = b[1].ui?.order ?? 999
    return orderA - orderB
  })
})

// 生成验证规则
const rules = computed(() => {
  const result: Record<string, any[]> = {}
  const required = props.schema.required || []

  for (const [key, field] of Object.entries(props.schema.properties)) {
    const fieldRules: any[] = []

    if (required.includes(key)) {
      fieldRules.push({
        required: true,
        message: `${field.title || key}不能为空`
      })
    }

    if (field.type === 'string') {
      if (field.minLength) {
        fieldRules.push({
          min: field.minLength,
          message: `最少${field.minLength}个字符`
        })
      }
      if (field.maxLength) {
        fieldRules.push({
          max: field.maxLength,
          message: `最多${field.maxLength}个字符`
        })
      }
      if (field.pattern) {
        fieldRules.push({
          pattern: new RegExp(field.pattern),
          message: '格式不正确'
        })
      }
      if (field.format === 'email') {
        fieldRules.push({
          type: 'email',
          message: '请输入有效的邮箱地址'
        })
      }
      if (field.format === 'url') {
        fieldRules.push({
          type: 'url',
          message: '请输入有效的URL'
        })
      }
    }

    if (field.type === 'number') {
      if (field.minimum !== undefined) {
        fieldRules.push({
          type: 'number',
          min: field.minimum,
          message: `最小值为${field.minimum}`
        })
      }
      if (field.maximum !== undefined) {
        fieldRules.push({
          type: 'number',
          max: field.maximum,
          message: `最大值为${field.maximum}`
        })
      }
    }

    if (fieldRules.length > 0) {
      result[key] = fieldRules
    }
  }

  return result
})

// 初始化表单数据
watch(
  () => props.modelValue,
  (newValue) => {
    formData.value = { ...newValue }
    // 设置默认值
    for (const [key, field] of Object.entries(props.schema.properties)) {
      if (formData.value[key] === undefined && field.default !== undefined) {
        formData.value[key] = field.default
      }
    }
  },
  { immediate: true }
)

// 监听表单变化
watch(
  formData,
  (newValue) => {
    emit('update:modelValue', newValue)
  },
  { deep: true }
)

async function handleSubmit() {
  const valid = await formRef.value.validate()
  if (valid) {
    emit('submit', formData.value)
  }
}

function handleReset() {
  formRef.value.resetFields()
}
</script>

<style scoped>
.field-help {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
