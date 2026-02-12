<template>
  <div class="plugin-market">
    <div class="market-header">
      <el-input
        v-model="searchQuery"
        placeholder="搜索插件..."
        prefix-icon="Search"
        clearable
        @input="handleSearch"
      />
      <el-select v-model="selectedCategory" placeholder="分类" @change="handleSearch">
        <el-option label="全部" value="" />
        <el-option v-for="cat in categories" :key="cat.id" :label="cat.name" :value="cat.id" />
      </el-select>
      <el-select v-model="sortBy" @change="handleSearch">
        <el-option label="下载量" value="downloads" />
        <el-option label="评分" value="rating" />
        <el-option label="最近更新" value="updated" />
        <el-option label="名称" value="name" />
      </el-select>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="发现" name="discover">
        <div v-loading="loading" class="plugin-grid">
          <PluginCard
            v-for="plugin in plugins"
            :key="plugin.id"
            :plugin="plugin"
            @install="handleInstall"
            @uninstall="handleUninstall"
            @update="handleUpdate"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="已安装" name="installed">
        <div v-loading="loading" class="plugin-grid">
          <PluginCard
            v-for="plugin in installedPlugins"
            :key="plugin.id"
            :plugin="plugin"
            :installed="true"
            @uninstall="handleUninstall"
            @update="handleUpdate"
            @configure="handleConfigure"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="更新" name="updates">
        <el-alert v-if="updates.length === 0" type="success" :closable="false">
          所有插件都是最新版本
        </el-alert>
        <div v-else class="updates-list">
          <UpdateCard
            v-for="update in updates"
            :key="update.pluginId"
            :update="update"
            @update="handleUpdate"
          />
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import PluginCard from '../components/PluginCard.vue'
import UpdateCard from '../components/UpdateCard.vue'

const searchQuery = ref('')
const selectedCategory = ref('')
const sortBy = ref('downloads')
const activeTab = ref('discover')
const loading = ref(false)
const plugins = ref([])
const installedPlugins = ref([])
const updates = ref([])
const categories = ref([])

onMounted(async () => {
  await loadCategories()
  await loadPlugins()
  await loadInstalled()
  await checkUpdates()
})

async function loadCategories() {
  try {
    categories.value = await window.api.pluginMarket.getCategories()
  } catch (error) {
    console.error('加载分类失败:', error)
  }
}

async function loadPlugins() {
  loading.value = true
  try {
    plugins.value = await window.api.pluginMarket.search(searchQuery.value, {
      category: selectedCategory.value,
      sort: sortBy.value
    })
  } catch (error) {
    ElMessage.error('加载插件列表失败')
  } finally {
    loading.value = false
  }
}

async function loadInstalled() {
  try {
    installedPlugins.value = await window.api.pluginMarket.listInstalled()
  } catch (error) {
    console.error('加载已安装插件失败:', error)
  }
}

async function checkUpdates() {
  try {
    updates.value = await window.api.pluginMarket.checkUpdates()
  } catch (error) {
    console.error('检查更新失败:', error)
  }
}

function handleSearch() {
  loadPlugins()
}

async function handleInstall(pluginId: string) {
  loading.value = true
  try {
    await window.api.pluginMarket.install(pluginId)
    ElMessage.success('插件安装成功')
    await loadInstalled()
  } catch (error) {
    ElMessage.error('插件安装失败')
  } finally {
    loading.value = false
  }
}

async function handleUninstall(pluginId: string) {
  try {
    await window.api.pluginMarket.uninstall(pluginId)
    ElMessage.success('插件卸载成功')
    await loadInstalled()
  } catch (error) {
    ElMessage.error('插件卸载失败')
  }
}

async function handleUpdate(pluginId: string) {
  loading.value = true
  try {
    await window.api.pluginMarket.update(pluginId)
    ElMessage.success('插件更新成功')
    await loadInstalled()
    await checkUpdates()
  } catch (error) {
    ElMessage.error('插件更新失败')
  } finally {
    loading.value = false
  }
}

function handleConfigure(pluginId: string) {
  // 打开插件配置对话框
  window.api.plugins.openConfig(pluginId)
}
</script>

<style scoped>
.plugin-market {
  padding: 20px;
}

.market-header {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.market-header .el-input {
  flex: 1;
}

.plugin-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  min-height: 400px;
}

.updates-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>
