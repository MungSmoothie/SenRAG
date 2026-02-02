<template>
  <div class="config">
    <div class="card">
      <h2>⚙️ 系统配置</h2>
      
      <div class="config-section">
        <h3>LLM 配置</h3>
        <div class="form-group">
          <label>Base URL</label>
          <input v-model="config.llm.base_url" placeholder="https://api.openai.com/v1">
        </div>
        <div class="form-group">
          <label>API Key</label>
          <input v-model="config.llm.api_key" type="password" placeholder="sk-...">
        </div>
        <div class="form-group">
          <label>模型</label>
          <input v-model="config.llm.model" placeholder="gpt-3.5-turbo">
        </div>
      </div>

      <div class="config-section">
        <h3>Qdrant 配置</h3>
        <div class="form-group">
          <label>Host</label>
          <input v-model="config.qdrant.host" placeholder="localhost">
        </div>
        <div class="form-group">
          <label>Port</label>
          <input v-model.number="config.qdrant.port" type="number">
        </div>
      </div>

      <div class="actions">
        <button class="btn-secondary" @click="loadConfig">刷新</button>
        <button class="btn-primary" @click="saveConfig">保存</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const config = ref({
  llm: { base_url: '', api_key: '', model: 'gpt-3.5-turbo' },
  qdrant: { host: 'localhost', port: 6333 }
})

const loadConfig = async () => {
  try {
    const res = await axios.get('/api/config')
    config.value = res.data
  } catch (err) {
    console.error('加载配置失败:', err)
  }
}

const saveConfig = async () => {
  try {
    await axios.put('/api/config', config.value)
    alert('配置已保存')
  } catch (err) {
    console.error('保存配置失败:', err)
  }
}

onMounted(loadConfig)
</script>

<style scoped>
.config {
  max-width: 600px;
  margin: 0 auto;
}

.config h2 {
  margin-bottom: 1.5rem;
}

.config-section {
  margin-bottom: 1.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid #e9ecef;
}

.config-section:last-of-type {
  border-bottom: none;
}

.config-section h3 {
  marginrem;
  color-bottom: 1: #667eea;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #666;
  font-size: 0.9rem;
}

.actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}
</style>
