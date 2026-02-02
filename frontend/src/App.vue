<template>
  <div class="app">
    <header class="header">
      <h1>🧠 SenRAG</h1>
      <p class="subtitle">个人知识库 + RAG 问答系统</p>
    </header>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
  background: #f5f7fa;
}
.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 2rem;
  text-align: center;
}
.header h1 {
  margin: 0;
  font-size: 2.5rem;
}
.subtitle {
  margin: 0.5rem 0 0;
  opacity: 0.9;
}
.main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}
</style>
