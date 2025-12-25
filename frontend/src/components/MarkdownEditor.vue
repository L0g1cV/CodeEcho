<template>
  <div class="markdown-editor-wrapper" :class="{ 'read-mode': mode === 'read' }">
    <!-- Editor with Notion-style slash commands -->
    <div class="markdown-editor" ref="editorRef" @click="focusEditor"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Editor, rootCtx, defaultValueCtx, editorViewOptionsCtx, commandsCtx } from '@milkdown/core'
import { commonmark, toggleStrongCommand, toggleEmphasisCommand, wrapInBlockquoteCommand, wrapInBulletListCommand, wrapInOrderedListCommand, insertHrCommand } from '@milkdown/preset-commonmark'
import { gfm, toggleStrikethroughCommand, insertTableCommand } from '@milkdown/preset-gfm'
import { listener, listenerCtx } from '@milkdown/plugin-listener'
import { history, undoCommand, redoCommand } from '@milkdown/plugin-history'
import { indent } from '@milkdown/plugin-indent'
import { cursor } from '@milkdown/plugin-cursor'
import { clipboard } from '@milkdown/plugin-clipboard'
import { trailing } from '@milkdown/plugin-trailing'
import { nord } from '@milkdown/theme-nord'
import '@milkdown/theme-nord/style.css'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  mode?: 'edit' | 'read'
}>(), {
  mode: 'edit'
})

// Computed for template usage
const mode = computed(() => props.mode)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const editorRef = ref<HTMLElement | null>(null)
let editorInstance: Editor | null = null

// Focus the editor when clicking anywhere in the editor container
const focusEditor = () => {
  if (props.mode === 'read') return
  const proseMirror = editorRef.value?.querySelector('.ProseMirror') as HTMLElement
  if (proseMirror) {
    proseMirror.focus()
  }
}

// Execute command helper
const executeCommand = (command: any) => {
  if (!editorInstance) return
  try {
    editorInstance.action((ctx) => {
      const commandManager = ctx.get(commandsCtx)
      commandManager.call(command.key)
    })
  } catch (e) {
    console.warn('Command execution failed:', e)
  }
}

// Keyboard shortcuts
const handleKeyDown = (e: KeyboardEvent) => {
  if (props.mode !== 'edit' || !editorInstance) return
  
  const isMod = e.metaKey || e.ctrlKey
  
  if (isMod && e.key === 'b') {
    e.preventDefault()
    executeCommand(toggleStrongCommand)
  } else if (isMod && e.key === 'i') {
    e.preventDefault()
    executeCommand(toggleEmphasisCommand)
  } else if (isMod && e.key === 'z') {
    e.preventDefault()
    if (e.shiftKey) {
      executeCommand(redoCommand)
    } else {
      executeCommand(undoCommand)
    }
  }
}

// Track last emitted value to distinguish from external updates
let lastEmittedValue = ''

onMounted(async () => {
  if (!editorRef.value) return

  try {
    // Track initial value to distinguish from external updates
    lastEmittedValue = props.modelValue || ''
    
    editorInstance = await Editor.make()
      .config((ctx) => {
        ctx.set(rootCtx, editorRef.value)
        ctx.set(defaultValueCtx, props.modelValue || '')
        ctx.get(listenerCtx).markdownUpdated((_, markdown) => {
          lastEmittedValue = markdown
          emit('update:modelValue', markdown)
        })
        // Set editable based on mode
        ctx.set(editorViewOptionsCtx, {
          editable: () => props.mode === 'edit'
        })
      })
      .config(nord)
      .use(commonmark)
      .use(gfm)
      .use(listener)
      .use(history)
      .use(indent)
      .use(cursor)
      .use(clipboard)
      .use(trailing)
      .create()
    
    // Add keyboard shortcut listener
    document.addEventListener('keydown', handleKeyDown)
    
  } catch (error) {
    console.error('Failed to create editor:', error)
  }
})

// Watch for mode changes to update editable state
watch(() => props.mode, () => {
  // The editable function reads props.mode dynamically
  // Force a view update if needed
  if (editorInstance) {
    const view = (editorInstance as any).ctx?.get?.('editorView')
    if (view) {
      view.updateState(view.state)
    }
  }
})

// Watch for external value changes - recreate editor when content changes externally
watch(() => props.modelValue, async (newValue) => {
  // Skip if this is our own update (from the editor)
  if (newValue === lastEmittedValue) return
  
  // External update detected - recreate the editor with new content
  if (!editorRef.value) return
  
  // Destroy old editor
  editorInstance?.destroy()
  editorInstance = null
  
  try {
    editorInstance = await Editor.make()
      .config((ctx) => {
        ctx.set(rootCtx, editorRef.value)
        ctx.set(defaultValueCtx, newValue || '')
        ctx.get(listenerCtx).markdownUpdated((_, markdown) => {
          lastEmittedValue = markdown
          emit('update:modelValue', markdown)
        })
        ctx.set(editorViewOptionsCtx, {
          editable: () => props.mode === 'edit'
        })
      })
      .config(nord)
      .use(commonmark)
      .use(gfm)
      .use(listener)
      .use(history)
      .use(indent)
      .use(cursor)
      .use(clipboard)
      .use(trailing)
      .create()
  } catch (error) {
    console.error('Failed to recreate editor:', error)
  }
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
  editorInstance?.destroy()
})

// Expose commands for external usage if needed
defineExpose({
  executeCommand,
  toggleBold: () => executeCommand(toggleStrongCommand),
  toggleItalic: () => executeCommand(toggleEmphasisCommand),
  toggleStrikethrough: () => executeCommand(toggleStrikethroughCommand),
  insertBlockquote: () => executeCommand(wrapInBlockquoteCommand),
  insertBulletList: () => executeCommand(wrapInBulletListCommand),
  insertOrderedList: () => executeCommand(wrapInOrderedListCommand),
  insertHr: () => executeCommand(insertHrCommand),
  insertTable: () => executeCommand(insertTableCommand),
  undo: () => executeCommand(undoCommand),
  redo: () => executeCommand(redoCommand),
})
</script>

<style scoped>
/* Notion-style CSS variables */
.markdown-editor-wrapper {
  --notion-text-color: #37352f;
  --notion-bg-color: #ffffff;
  --notion-border-color: #e1e7e8;
  --notion-link-color: #2eaadc;
  --notion-code-bg: #f7f6f3;
  --notion-code-color: #eb5757;
  --notion-quote-border: #e1e7e8;
  
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  overflow: hidden;
}

/* Editor - Notion-like styling */
.markdown-editor {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior: contain;
  cursor: text;
  background: #ffffff;
}

.markdown-editor :deep(.milkdown) {
  padding: 20px 32px;
  min-height: 100%;
}

.markdown-editor :deep(.milkdown .editor) {
  outline: none;
  min-height: 100%;
}

/* ProseMirror base - Notion style */
.markdown-editor :deep(.ProseMirror) {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans SC", sans-serif;
  font-size: 16px;
  line-height: 1.6;
  color: #37352f;
  caret-color: #37352f;
  outline: none;
}

.markdown-editor :deep(.ProseMirror > *:first-child) {
  margin-top: 0;
}

/* Paragraphs */
.markdown-editor :deep(.ProseMirror p) {
  margin: 0 0 1px;
  padding: 3px 2px;
}

.markdown-editor :deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: #c1c1c1;
  pointer-events: none;
  height: 0;
}

/* Headings - Notion style */
.markdown-editor :deep(.ProseMirror h1) {
  font-size: 1.875rem;
  font-weight: 700;
  margin-top: 2rem;
  margin-bottom: 4px;
  color: #37352f;
  line-height: 1.3;
}

.markdown-editor :deep(.ProseMirror h2) {
  font-size: 1.5rem;
  font-weight: 600;
  margin-top: 1.4rem;
  margin-bottom: 4px;
  color: #37352f;
  line-height: 1.3;
}

.markdown-editor :deep(.ProseMirror h3) {
  font-size: 1.25rem;
  font-weight: 600;
  margin-top: 1rem;
  margin-bottom: 4px;
  color: #37352f;
  line-height: 1.3;
}

.markdown-editor :deep(.ProseMirror h4),
.markdown-editor :deep(.ProseMirror h5),
.markdown-editor :deep(.ProseMirror h6) {
  font-size: 1rem;
  font-weight: 600;
  margin-top: 1rem;
  margin-bottom: 4px;
  color: #37352f;
}

/* Inline code - Notion red style */
.markdown-editor :deep(.ProseMirror code) {
  font-family: "SFMono-Regular", Menlo, Consolas, monospace;
  font-size: 85%;
  background: rgba(135, 131, 120, 0.15);
  padding: 0.2em 0.4em;
  border-radius: 3px;
  color: #eb5757;
}

/* Code blocks - Notion style */
.markdown-editor :deep(.ProseMirror pre) {
  font-family: "SFMono-Regular", Menlo, Consolas, monospace;
  font-size: 0.875rem;
  line-height: 1.5;
  background: #f7f6f3;
  color: #37352f;
  padding: 32px 16px 32px 32px;
  border-radius: 3px;
  overflow-x: auto;
  margin: 4px 0;
}

.markdown-editor :deep(.ProseMirror pre code) {
  background: none;
  padding: 0;
  color: inherit;
  font-size: inherit;
}

/* Blockquote - Notion style */
.markdown-editor :deep(.ProseMirror blockquote) {
  border-left: 3px solid currentColor;
  margin: 4px 0;
  padding: 0 0 0 14px;
  color: #37352f;
}

/* Lists - Notion style */
.markdown-editor :deep(.ProseMirror ul),
.markdown-editor :deep(.ProseMirror ol) {
  padding-left: 24px;
  margin: 0;
}

.markdown-editor :deep(.ProseMirror li) {
  margin: 0;
  padding: 3px 0;
}

.markdown-editor :deep(.ProseMirror li p) {
  margin: 0;
}

/* Task list - Notion style */
.markdown-editor :deep(.ProseMirror .task-list-item) {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.markdown-editor :deep(.ProseMirror .task-list-item input[type="checkbox"]) {
  margin-top: 6px;
  width: 16px;
  height: 16px;
  cursor: pointer;
}

/* Links */
.markdown-editor :deep(.ProseMirror a) {
  color: #37352f;
  text-decoration: underline;
  text-decoration-color: rgba(55, 53, 47, 0.4);
  text-underline-offset: 3px;
}

.markdown-editor :deep(.ProseMirror a:hover) {
  text-decoration-color: #37352f;
}

/* Tables - Notion style */
.markdown-editor :deep(.ProseMirror table) {
  border-collapse: collapse;
  margin: 4px 0;
  width: 100%;
}

.markdown-editor :deep(.ProseMirror th),
.markdown-editor :deep(.ProseMirror td) {
  border: 1px solid #e1e7e8;
  padding: 8px 10px;
  text-align: left;
}

.markdown-editor :deep(.ProseMirror th) {
  background: #f7f6f3;
  font-weight: 600;
}

/* Horizontal rule */
.markdown-editor :deep(.ProseMirror hr) {
  border: none;
  border-top: 1px solid #e1e7e8;
  margin: 16px 0;
}

/* Images */
.markdown-editor :deep(.ProseMirror img) {
  max-width: 100%;
  border-radius: 3px;
}

/* Selection */
.markdown-editor :deep(.ProseMirror ::selection) {
  background: rgba(45, 170, 219, 0.3);
}

/* Focus state */
.markdown-editor :deep(.ProseMirror-focused) {
  outline: none;
}

/* Read Mode Styles */
.markdown-editor-wrapper.read-mode .markdown-editor {
  cursor: default;
}

.markdown-editor-wrapper.read-mode :deep(.ProseMirror) {
  cursor: default;
  caret-color: transparent;
}

/* Strikethrough */
.markdown-editor :deep(.ProseMirror s),
.markdown-editor :deep(.ProseMirror del) {
  text-decoration: line-through;
}

/* Strong/Bold */
.markdown-editor :deep(.ProseMirror strong) {
  font-weight: 600;
}

/* Emphasis/Italic */
.markdown-editor :deep(.ProseMirror em) {
  font-style: italic;
}
</style>
