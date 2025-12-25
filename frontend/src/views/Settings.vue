<template>
  <div class="settings-page">
    <header class="settings-header">
      <router-link to="/" class="back-link">← Back to Dashboard</router-link>
      <h1>Settings</h1>
    </header>

    <main class="settings-content">
      <!-- LLM Configuration -->
      <section class="settings-section">
        <h2>AI Configuration (BYOK)</h2>
        <p class="section-desc">Configure your own LLM API key to power the Echo feature.</p>

        <div class="form-group">
          <label>Provider</label>
          <select v-model="llmConfig.provider">
            <option value="openai">OpenAI</option>
            <option value="claude">Claude (Anthropic)</option>
            <option value="deepseek">DeepSeek</option>
            <option value="custom">Custom Endpoint</option>
          </select>
        </div>

        <div class="form-group">
          <label>API Key</label>
          <input
            type="password"
            v-model="llmConfig.apiKey"
            placeholder="sk-..."
          />
        </div>

        <div class="form-group" v-if="llmConfig.provider === 'custom'">
          <label>Base URL</label>
          <input
            type="text"
            v-model="llmConfig.baseUrl"
            placeholder="https://api.example.com/v1"
          />
        </div>

        <div class="form-group">
          <label>Model</label>
          <input
            type="text"
            v-model="llmConfig.model"
            :placeholder="getModelPlaceholder()"
          />
        </div>

        <button class="save-btn" @click="saveLLMConfig">Save Configuration</button>
      </section>

      <!-- Appearance -->
      <section class="settings-section">
        <h2>Appearance</h2>

        <div class="form-group">
          <label>Theme</label>
          <div class="theme-options">
            <button
              :class="['theme-btn', { active: uiStore.theme === 'light' }]"
              @click="uiStore.setTheme('light')"
            >
              ☀️ Light
            </button>
            <button
              :class="['theme-btn', { active: uiStore.theme === 'dark' }]"
              @click="uiStore.setTheme('dark')"
            >
              🌙 Dark
            </button>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useUIStore } from '../stores/ui'

const uiStore = useUIStore()

const llmConfig = reactive({
  provider: 'openai' as 'openai' | 'claude' | 'deepseek' | 'custom',
  apiKey: '',
  baseUrl: '',
  model: '',
})

const getModelPlaceholder = () => {
  switch (llmConfig.provider) {
    case 'openai':
      return 'gpt-4o'
    case 'claude':
      return 'claude-3-5-sonnet-20241022'
    case 'deepseek':
      return 'deepseek-chat'
    default:
      return 'model-name'
  }
}

const saveLLMConfig = async () => {
  // TODO: Call API to save LLM config
  console.log('Saving LLM config:', llmConfig)
  alert('Configuration saved!')
}
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
  background: var(--bg-primary);
}

.settings-header {
  padding: 24px 32px;
  border-bottom: 1px solid var(--border-color);
}

.back-link {
  display: inline-block;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.875rem;
  margin-bottom: 8px;
}

.back-link:hover {
  color: var(--text-primary);
}

.settings-header h1 {
  margin: 0;
  font-size: 1.75rem;
  color: var(--text-primary);
}

.settings-content {
  max-width: 640px;
  margin: 0 auto;
  padding: 32px;
}

.settings-section {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
}

.settings-section h2 {
  font-size: 1.25rem;
  margin: 0 0 8px;
  color: var(--text-primary);
}

.section-desc {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0 0 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 0.9375rem;
  outline: none;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  border-color: var(--accent-color);
}

.save-btn {
  padding: 12px 24px;
  background: var(--accent-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.save-btn:hover {
  opacity: 0.9;
}

.theme-options {
  display: flex;
  gap: 12px;
}

.theme-btn {
  flex: 1;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s;
}

.theme-btn:hover {
  border-color: var(--accent-color);
}

.theme-btn.active {
  border-color: var(--accent-color);
  background: var(--accent-color);
  color: white;
}
</style>
