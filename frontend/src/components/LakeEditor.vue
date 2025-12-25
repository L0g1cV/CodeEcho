<template>
  <div class="lake-editor-wrapper" :class="{ 'read-mode': mode === 'read' }">
    <div class="lake-toolbar" ref="toolbarRef"></div>
    <div class="lake-content" ref="contentRef" @click="handleContentClick"></div>
  </div>
</template>

<script setup lang="ts">
import 'lakelib/lib/lake.css'
import { Editor, Toolbar } from 'lakelib'
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'

// Lazy load CodeMirror and KaTeX to avoid blocking component setup
let codeMirrorLoaded = false
let katexLoaded = false

const loadCodeMirror = async () => {
  if (codeMirrorLoaded) return
  try {
    const CodeMirror = await import('lake-codemirror')
    ;(window as any).LakeCodeMirror = CodeMirror
    codeMirrorLoaded = true
  } catch (e) {
    console.warn('Failed to load CodeMirror:', e)
  }
}

const loadKaTeX = async () => {
  if (katexLoaded) return
  try {
    await import('katex/dist/katex.css')
    const katex = await import('katex')
    ;(window as any).katex = katex.default || katex
    katexLoaded = true
  } catch (e) {
    console.warn('Failed to load KaTeX:', e)
  }
}

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  mode?: 'edit' | 'read'
}>(), {
  mode: 'edit'
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

// Computed for template usage
const mode = computed(() => props.mode)

const toolbarRef = ref<HTMLElement | null>(null)
const contentRef = ref<HTMLElement | null>(null)

let editor: Editor | null = null
let isUpdatingFromProp = false

// Get auth token for authenticated uploads
const getAuthToken = (): string | null => {
  return localStorage.getItem('token')
}

// Common editor config (called only when refs are guaranteed to exist)
const getEditorConfig = (readonly: boolean, value: string) => ({
  root: contentRef.value!,
  value,
  readonly,
  lang: 'en-US' as const,
  indentWithTab: true,
  spellcheck: false,
  
  // Slash commands configuration - all single-user features
  slash: {
    items: [
      'heading1',
      'heading2', 
      'heading3',
      'heading4',
      'heading5',
      'heading6',
      'paragraph',
      'blockQuote',
      'numberedList',
      'bulletedList',
      'checklist',
      'codeBlock',
      'equation',
      'table',
      'hr',
      'image',
      'file',
      'video',
      'infoAlert',
      'tipAlert',
      'warningAlert',
      'dangerAlert',
    ],
  },
  
  // Code block with language selector
  codeBlock: {
    langList: [
      'text',
      'bash',
      'c',
      'cpp',
      'csharp',
      'css',
      'dockerfile',
      'go',
      'html',
      'java',
      'javascript',
      'json',
      'kotlin',
      'lua',
      'markdown',
      'nginx',
      'php',
      'python',
      'ruby',
      'rust',
      'scala',
      'shell',
      'sql',
      'swift',
      'typescript',
      'vue',
      'xml',
      'yaml',
    ],
    defaultLang: 'text',
  },
  
  // Image upload configuration
  image: {
    requestMethod: 'POST',
    requestAction: '/api/upload',
    requestTypes: ['image/png', 'image/jpeg', 'image/gif', 'image/webp', 'image/svg+xml'],
    requestFieldName: 'file',
    requestHeaders: () => {
      const token = getAuthToken()
      return token ? { Authorization: `Bearer ${token}` } : {}
    },
    requestWithCredentials: false,
    parseResponse: (response: any) => {
      try {
        const data = typeof response === 'string' ? JSON.parse(response) : response
        return { url: data.url }
      } catch (e) {
        console.error('Failed to parse upload response:', e)
        return { url: '' }
      }
    },
  },
  
  // File upload configuration (for attachments)
  file: {
    requestMethod: 'POST',
    requestAction: '/api/upload',
    requestTypes: [
      'application/pdf',
      'application/zip',
      'application/x-rar-compressed',
      'application/msword',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'application/vnd.ms-excel',
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      'text/plain',
    ],
    requestFieldName: 'file',
    requestHeaders: () => {
      const token = getAuthToken()
      return token ? { Authorization: `Bearer ${token}` } : {}
    },
    requestWithCredentials: false,
    parseResponse: (response: any) => {
      try {
        const data = typeof response === 'string' ? JSON.parse(response) : response
        return { url: data.url }
      } catch (e) {
        console.error('Failed to parse upload response:', e)
        return { url: '' }
      }
    },
  },
})

onMounted(async () => {
  if (!toolbarRef.value || !contentRef.value) return

  // Load plugins before initializing editor
  await Promise.all([loadCodeMirror(), loadKaTeX()])

  const toolbar = new Toolbar({
    root: toolbarRef.value,
  })

  editor = new Editor({
    ...getEditorConfig(props.mode === 'read', props.modelValue || ''),
    toolbar,
  })

  editor.render()

  // Listen for content changes
  editor.event.on('change', (value: string) => {
    if (!isUpdatingFromProp) {
      emit('update:modelValue', value)
    }
  })
})

// Watch for external value changes
watch(() => props.modelValue, (newValue) => {
  if (editor && newValue !== editor.getValue()) {
    isUpdatingFromProp = true
    editor.setValue(newValue || '')
    isUpdatingFromProp = false
  }
})

// Watch for mode changes - recreate editor with new readonly setting
watch(() => props.mode, (newMode) => {
  if (editor && contentRef.value && toolbarRef.value) {
    const currentValue = editor.getValue()
    editor.unmount()
    
    const toolbar = new Toolbar({
      root: toolbarRef.value,
    })
    
    editor = new Editor({
      ...getEditorConfig(newMode === 'read', currentValue),
      toolbar,
    })
    
    editor.render()
    
    editor.event.on('change', (value: string) => {
      if (!isUpdatingFromProp) {
        emit('update:modelValue', value)
      }
    })
  }
})

onUnmounted(() => {
  if (editor) {
    editor.unmount()
    editor = null
  }
})

// Handle click on content area to focus editor
const handleContentClick = (event: MouseEvent) => {
  // Only focus if in edit mode and click target is the content wrapper or editor container (not specific content)
  if (props.mode === 'read' || !editor) return
  
  const target = event.target as HTMLElement
  // If clicking on empty areas (lake-content wrapper or lake-container), focus the editor
  if (target.classList.contains('lake-content') || 
      target.classList.contains('lake-container') ||
      target.classList.contains('lake-main') ||
      target.classList.contains('lake-editor')) {
    editor.focus()
  }
}

// Expose methods for parent component
defineExpose({
  focus: () => editor?.focus(),
  blur: () => editor?.blur(),
  getValue: () => editor?.getValue() || '',
  setValue: (value: string) => editor?.setValue(value),
})
</script>

<style scoped>
.lake-editor-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-secondary, #f5f5f5);
  border-radius: 8px;
  overflow: hidden;
}

.lake-toolbar {
  flex-shrink: 0;
  background: var(--bg-primary, #ffffff);
  border-bottom: 1px solid var(--border-color, #e0e0e0);
}

.lake-content {
  flex: 1;
  overflow: auto;
  background: var(--bg-primary, #ffffff);
  padding: 16px;
  cursor: text;
}

/* Ensure edit mode shows text cursor throughout the editor area */
.lake-editor-wrapper:not(.read-mode) .lake-content {
  cursor: text;
}

.lake-editor-wrapper:not(.read-mode) :deep(.lake-container),
.lake-editor-wrapper:not(.read-mode) :deep(.lake-editor) {
  cursor: text;
}

/* Hide toolbar in read mode */
.read-mode .lake-toolbar {
  display: none;
}

/* Lake editor overrides for dark mode compatibility */
:deep(.lake-container) {
  background: var(--bg-primary, #ffffff);
  color: var(--text-primary, #1a1a2e);
}

:deep(.lake-editor) {
  min-height: 200px;
}

:deep(.lake-toolbar) {
  background: var(--bg-secondary, #f5f5f5);
}
</style>
