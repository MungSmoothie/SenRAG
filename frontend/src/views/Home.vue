<template>
  <div class="home">
    <!-- 上传区域 -->
    <div class="card upload-section">
      <h2>📤 上传文档</h2>
      <div class="upload-area" @drop.prevent="handleDrop" @dragover.prevent="isDragging = true"
           @dragleave="isDragging = false" :class="{ dragging: isDragging }">
        <input type="file" ref="fileInput" @change="handleFileSelect" multiple accept=".txt,.md,.pdf,.json,.yaml,.html" hidden>
        <div class="upload-content" @click="$refs.fileInput.click()">
          <p v-if="!uploading">📁 拖拽文件到此处，或点击选择文件</p>
          <p v-else>⏳ 上传中...</p>
          <small>支持 txt, md, pdf, json, yaml, html</small>
        </div>
      </div>
      <div v-if="uploadStatus" :class="['status', uploadStatus.type]">
        {{ uploadStatus.message }}
      </div>
    </div>

    <!-- 问答区域 -->
    <div class="card chat-section">
      <h2>💬 智能问答</h2>
      
      <div class="chat-history" ref="chatHistory">
        <div v-for="(msg, idx) in messages" :key="idx" :class="['message', msg.role]">
          <div class="message-content" v-html="formatMessage(msg.content)"></div>
        </div>
        <div v-if="streaming" class="message assistant streaming">
          <span class="typing-indicator">正在输入</span>
        </div>
      </div>

      <div class="chat-input">
        <textarea v-model="question" @keydown.enter.exact.prevent="sendQuestion" 
                  placeholder="输入问题，按 Enter 发送..." rows="2"></textarea>
        <button class="btn-primary" @click="sendQuestion" :disabled="!question.trim() || streaming">
          发送
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import axios from 'axios'

const fileInput = ref(null)
const isDragging = ref(false)
const uploading = ref(false)
const uploadStatus = ref(null)
const question = ref('')
const messages = reactive([
  { role: 'assistant', content: '你好！我是 SenRAG 助手。上传文档后，你可以问我关于文档内容的问题。' }
])
const streaming = ref(false)
const chatHistory = ref(null)

const isDragging = ref(false)

const handleDrop = async (e) => {
  isDragging.value = false
  const files = Array.from(e.dataTransfer.files)
  if (files.length > 0) {
    await uploadFiles(files)
  }
}

const handleFileSelect = async (e) => {
  const files = Array.from(e.target.files)
  if (files.length > 0) {
    await uploadFiles(files)
  }
}

const uploadFiles = async (files) => {
  uploading.value = true
  uploadStatus.value = null
  
  for (const file of files) {
    const formData = new FormData()
    formData.append('file', file)
    
    try {
      await axios.post('/api/upload', formData)
      uploadStatus.value = { type: 'success', message: `✅ ${file.name} 上传成功！` }
    } catch (err) {
      uploadStatus.value = { type: 'error', message: `❌ ${file.name} 上传失败: ${err.message}` }
    }
  }
  
  uploading.value = false
}

const sendQuestion = async () => {
  if (!question.value.trim() || streaming.value) return
  
  const q = question.value
  question.value = ''
  messages.push({ role: 'user', content: q })
  
  scrollToBottom()
  streaming.value = true
  
  try {
    const response = await axios.post('/api/query', { question: q })
    messages.push({ role: 'assistant', content: response.data.answer })
  } catch (err) {
    messages.push({ role: 'assistant', content: `❌ 发生错误: ${err.message}` })
  }
  
  streaming.value = false
  scrollToBottom()
}

const scrollToBottom = () => {
  setTimeout(() => {
    if (chatHistory.value) {
      chatHistory.value.scrollTop = chatHistory.value.scrollHeight
    }
  }, 100)
}

const formatMessage = (content) => {
  return content.replace(/\n/g, '<br>')
}
</script>

<style scoped>
.home {
  display: grid;
  gap: 1.5rem;
}

.upload-section h2,
.chat-section h2 {
  margin-bottom: 1rem;
  color: #333;
}

.upload-area {
  border: 2px dashed #ccc;
  border-radius: 12px;
  padding: 2rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.upload-area.dragging {
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.1);
}

.upload-area:hover {
  border-color: #667eea;
}

.upload-content {
  pointer-events: none;
}

.status {
  margin-top: 1rem;
  padding: 0.75rem;
  border-radius: 8px;
}

.status.success {
  background: #d4edda;
  color: #155724;
}

.status.error {
  background: #f8d7da;
  color: #721c24;
}

.chat-history {
  max-height: 400px;
  overflow-y: auto;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.message {
  margin-bottom: 1rem;
  padding: 0.75rem 1rem;
  border-radius: 12px;
  max-width: 85%;
}

.message.user {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  margin-left: auto;
}

.message.assistant {
  background: white;
  border: 1px solid #e9ecef;
}

.message-content {
  white-space: pre-wrap;
}

.typing-indicator {
  color: #666;
  font-style: italic;
}

.chat-input {
  display: flex;
  gap: 0.5rem;
}

.chat-input textarea {
  flex: 1;
  resize: none;
}
</style>
