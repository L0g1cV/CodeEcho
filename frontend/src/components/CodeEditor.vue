<template>
  <div class="code-editor">
    <div class="code-header">
      <select v-model="language" class="language-select">
        <option value="javascript">JavaScript</option>
        <option value="typescript">TypeScript</option>
        <option value="python">Python</option>
        <option value="go">Go</option>
        <option value="java">Java</option>
        <option value="cpp">C++</option>
        <option value="rust">Rust</option>
      </select>
    </div>
    <textarea
      :value="modelValue"
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
      class="code-textarea"
      :placeholder="placeholder"
      spellcheck="false"
    ></textarea>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const language = ref('python')
</script>

<style scoped>
.code-editor {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-secondary);
  overflow: hidden;
}

.code-header {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-hover);
}

.language-select {
  padding: 4px 8px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-card);
  color: var(--text-primary);
  font-size: 0.75rem;
  outline: none;
}

.code-textarea {
  width: 100%;
  min-height: 200px;
  padding: 16px;
  border: none;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 0.875rem;
  line-height: 1.6;
  resize: vertical;
  outline: none;
}

.code-textarea::placeholder {
  color: var(--text-tertiary);
}
</style>
