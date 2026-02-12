<template>
  <div
    class="file-upload-area"
    :class="{ dragging: isDragging }"
    @drop.prevent="handleDrop"
    @dragover.prevent="isDragging = true"
    @dragleave.prevent="isDragging = false"
  >
    <input
      ref="fileInputRef"
      type="file"
      multiple
      accept=".txt,.md,.json,.yaml,.yml,.log,.conf,.sh,.py,.js,.ts"
      @change="handleFileSelect"
      style="display: none"
    />
    
    <el-button text size="small" @click="fileInputRef?.click()">
      <el-icon><Paperclip /></el-icon>
    </el-button>
    
    <!-- 已选文件列表 -->
    <div v-if="files.length > 0" class="file-list">
      <el-tag
        v-for="(file, i) in files"
        :key="i"
        closable
        @close="removeFile(i)"
        size="small"
      >
        {{ file.name }} ({{ formatSize(file.size) }})
      </el-tag>
    </div>
    
    <!-- 拖拽提示 -->
    <div v-if="isDragging" class="drag-overlay">
      <el-icon :size="48"><Upload /></el-icon>
      <p>释放文件以上传</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Paperclip, Upload } from '@element-plus/icons-vue'

const emit = defineEmits<{
  filesSelected: [files: File[]]
}>()

const fileInputRef = ref<HTMLInputElement>()
const files = ref<File[]>([])
const isDragging = ref(false)

const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB

function handleFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files) {
    addFiles(Array.from(input.files))
  }
}

function handleDrop(e: DragEvent) {
  isDragging.value = false
  if (e.dataTransfer?.files) {
    addFiles(Array.from(e.dataTransfer.files))
  }
}

function addFiles(newFiles: File[]) {
  for (const file of newFiles) {
    if (file.size > MAX_FILE_SIZE) {
      ElMessage.warning(`文件 ${file.name} 超过 10MB 限制`)
      continue
    }
    files.value.push(file)
  }
  
  if (files.value.length > 0) {
    emit('filesSelected', files.value)
  }
}

function removeFile(index: number) {
  files.value.splice(index, 1)
  emit('filesSelected', files.value)
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function clear() {
  files.value = []
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

defineExpose({ clear })
</script>

<style scoped>
.file-upload-area {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.drag-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(37, 99, 235, 0.1);
  backdrop-filter: blur(4px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  color: var(--ai-primary);
  pointer-events: none;
}

.drag-overlay p {
  margin-top: 12px;
  font-size: 18px;
  font-weight: 600;
}
</style>
