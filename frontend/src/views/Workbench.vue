<template>
  <div class="workbench" :class="{ 'zen-mode': uiStore.zenMode }">
    <!-- Sidebar -->
    <aside class="sidebar" v-if="!uiStore.zenMode">
      <div class="sidebar-header">
        <router-link to="/" class="back-link">← Back to Dashboard</router-link>
      </div>
      <div class="sidebar-meta">
        <div class="meta-item">
          <span class="meta-label">Notebook</span>
          <span class="meta-value">{{ notebookName }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">Updated</span>
          <span class="meta-value">{{ lastUpdated }}</span>
        </div>
      </div>
    </aside>

    <!-- Main Editor Area -->
    <main class="editor-area">
      <header class="editor-header">
        <input
          type="text"
          class="note-title-input"
          v-model="noteTitle"
          placeholder="Please enter the title"
          @blur="autoSave"
        />
        <div class="header-actions">
          <span class="save-status">{{ saveStatus }}</span>
          <button class="action-btn" @click="saveNote" :disabled="isSaving">
            {{ isSaving ? 'Saving...' : 'Save' }}
          </button>
          <button class="action-btn" @click="uiStore.toggleZenMode">
            {{ uiStore.zenMode ? 'Exit Zen' : 'Zen' }}
          </button>
          <button class="action-btn echo-btn" @click="toggleEchoPanel" v-if="!uiStore.zenMode">
            Echo
          </button>
        </div>
      </header>

      <div class="editor-container">
        <MarkdownEditor
          v-model="content"
          @update:modelValue="onContentChange"
          placeholder="Start writing in Markdown..."
        />
      </div>
    </main>

    <!-- AI Echo Panel (Collapsible) -->
    <aside class="ai-panel" v-if="showEchoPanel && !uiStore.zenMode">
      <header class="ai-header">
        <h3>Echo</h3>
        <button class="close-btn" @click="showEchoPanel = false">×</button>
      </header>

      <div class="ai-chat">
        <div class="chat-messages" ref="chatMessagesRef">
          <div
            class="message"
            :class="msg.role === 'user' ? 'user-message' : 'ai-message'"
            v-for="(msg, idx) in chatMessages"
            :key="idx"
          >
            <div class="message-role">{{ msg.role === 'user' ? 'You' : 'Echo' }}</div>
            <div class="message-content">{{ msg.content }}</div>
          </div>
          <div class="empty-chat" v-if="chatMessages.length === 0">
            <p>💡 Describe your algorithm in English</p>
            <p class="hint">I'll help you express it professionally.</p>
          </div>
        </div>

        <div class="chat-input-area">
          <textarea
            v-model="userDraft"
            placeholder="Try describing your solution..."
            class="chat-input"
            @keydown.enter.ctrl="sendToEcho"
          ></textarea>
          <button
            class="send-btn"
            @click="sendToEcho"
            :disabled="!userDraft.trim() || isEchoLoading"
          >
            {{ isEchoLoading ? '...' : 'Send' }}
          </button>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUIStore } from '../stores/ui'
import { noteApi } from '../api/notes'
import MarkdownEditor from '../components/MarkdownEditor.vue'

const route = useRoute()
const router = useRouter()
const uiStore = useUIStore()

const noteId = computed(() => route.params.id as string)
const isNewNote = computed(() => noteId.value === 'new')

const noteTitle = ref('')
const content = ref('')
const notebookId = ref('')
const notebookName = ref('Loading...')
const lastUpdated = ref('...')

const isSaving = ref(false)
const isDirty = ref(false)
const saveStatus = ref('')
const showEchoPanel = ref(false)
const isEchoLoading = ref(false)
const userDraft = ref('')

const chatMessagesRef = ref<HTMLElement | null>(null)

interface ChatMessage {
  role: 'user' | 'ai'
  content: string
}

const chatMessages = ref<ChatMessage[]>([])

let saveTimeout: ReturnType<typeof setTimeout> | null = null
let currentNoteId: string | null = null

const toggleEchoPanel = () => {
  showEchoPanel.value = !showEchoPanel.value
}

const onContentChange = () => {
  isDirty.value = true
  saveStatus.value = 'Unsaved'
  
  // Auto-save after 2 seconds of inactivity
  if (saveTimeout) clearTimeout(saveTimeout)
  saveTimeout = setTimeout(() => {
    autoSave()
  }, 2000)
}

const autoSave = async () => {
  if (!isDirty.value || isSaving.value) return
  if (!noteTitle.value.trim() && !content.value.trim()) return
  
  await saveNote()
}

const saveNote = async () => {
  if (isSaving.value) return

  isSaving.value = true
  saveStatus.value = 'Saving...'

  try {
    if (isNewNote.value || !currentNoteId) {
      // Get notebookId from query or use first available
      const queryNotebookId = route.query.notebookId as string
      let targetNotebookId = queryNotebookId || notebookId.value

      if (!targetNotebookId) {
        // Create a default notebook if none exists
        const notebooks = await noteApi.getNotebooks()
        if (notebooks.length === 0) {
          const newNotebook = await noteApi.createNotebook('My Notebook')
          targetNotebookId = newNotebook.id
        } else {
          targetNotebookId = notebooks[0]?.id || ''
        }
      }

      // Create new note
      const newNote = await noteApi.create({
        type: 'note',
        title: noteTitle.value || 'Untitled',
        parentId: targetNotebookId,
        content: content.value,
      })

      currentNoteId = newNote.id
      notebookId.value = newNote.parentId || ''

      // Update URL without reloading
      router.replace(`/note/${newNote.id}`)
    } else {
      // Update existing note
      await noteApi.update(currentNoteId, {
        title: noteTitle.value,
        content: content.value,
      })
    }

    isDirty.value = false
    saveStatus.value = 'Saved'
    lastUpdated.value = 'Just now'

    // Clear status after 2 seconds
    setTimeout(() => {
      if (saveStatus.value === 'Saved') {
        saveStatus.value = ''
      }
    }, 2000)
  } catch (error) {
    console.error('Failed to save note:', error)
    saveStatus.value = 'Failed to save'
  } finally {
    isSaving.value = false
  }
}

const loadNote = async (id: string) => {
  try {
    const note = await noteApi.getById(id)
    noteTitle.value = note.title
    content.value = note.content
    notebookId.value = note.parentId || ''
    currentNoteId = note.id
    lastUpdated.value = formatDate(note.updatedAt)

    // Load notebook name
    if (note.parentId) {
      const parent = await noteApi.getById(note.parentId)
      notebookName.value = parent?.title || 'Unknown'
    } else {
      notebookName.value = 'Unknown'
    }
  } catch (error) {
    console.error('Failed to load note:', error)
    router.push('/')
  }
}

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const mins = Math.floor(diff / 60000)

  if (mins < 1) return 'Just now'
  if (mins < 60) return `${mins} min ago`
  if (mins < 1440) return `${Math.floor(mins / 60)} hours ago`
  return date.toLocaleDateString()
}

const sendToEcho = async () => {
  if (!userDraft.value.trim() || isEchoLoading.value) return

  const draft = userDraft.value
  userDraft.value = ''

  chatMessages.value.push({ role: 'user', content: draft })

  await nextTick()
  if (chatMessagesRef.value) {
    chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
  }

  isEchoLoading.value = true

  try {
    await new Promise(resolve => setTimeout(resolve, 1000))
    chatMessages.value.push({
      role: 'ai',
      content: `Here's a more professional way to express that:\n\n→ Configure your API key in Settings to get real AI suggestions.`,
    })
  } catch (error) {
    chatMessages.value.push({
      role: 'ai',
      content: 'Sorry, something went wrong.',
    })
  } finally {
    isEchoLoading.value = false
    await nextTick()
    if (chatMessagesRef.value) {
      chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
    }
  }
}

onMounted(async () => {
  if (!isNewNote.value) {
    await loadNote(noteId.value)
  } else {
    notebookName.value = 'New Note'
    lastUpdated.value = 'Not saved'

    // Check if notebookId is provided
    const queryNotebookId = route.query.notebookId as string
    if (queryNotebookId) {
      notebookId.value = queryNotebookId
      const parent = await noteApi.getById(queryNotebookId)
      notebookName.value = parent?.title || 'New Note'
    }
  }
})

// Handle navigation away
watch(() => route.params.id, async (newId) => {
  if (newId && newId !== 'new') {
    await loadNote(newId as string)
  }
})
</script>

<style scoped>
.workbench {
  display: flex;
  height: 100vh;
  background: var(--bg-primary);
}

.workbench.zen-mode .editor-area {
  max-width: 800px;
  margin: 0 auto;
}

/* Sidebar */
.sidebar {
  width: 220px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
}

.back-link {
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.875rem;
}

.back-link:hover {
  color: var(--text-primary);
}

.sidebar-meta {
  padding: 16px;
}

.meta-item {
  margin-bottom: 12px;
}

.meta-label {
  display: block;
  font-size: 0.75rem;
  color: var(--text-tertiary);
  margin-bottom: 4px;
}

.meta-value {
  font-size: 0.875rem;
  color: var(--text-primary);
}

/* Editor Area */
.editor-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 24px;
  border-bottom: 1px solid var(--border-color);
  gap: 16px;
}

.note-title-input {
  flex: 1;
  font-size: 1.25rem;
  font-weight: 500;
  border: none;
  background: transparent;
  color: var(--text-primary);
  outline: none;
}

.note-title-input::placeholder {
  color: var(--text-tertiary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.save-status {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

.action-btn {
  padding: 6px 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.echo-btn {
  background: var(--accent-color);
  border-color: var(--accent-color);
  color: white;
}

.echo-btn:hover {
  opacity: 0.9;
  color: white;
}

/* Editor Container */
.editor-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px 48px;
}

/* AI Panel */
.ai-panel {
  width: 360px;
  background: var(--bg-secondary);
  border-left: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.ai-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
}

.ai-header h3 {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.25rem;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 0 4px;
}

.close-btn:hover {
  color: var(--text-primary);
}

.ai-chat {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chat-messages {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.message {
  margin-bottom: 16px;
}

.message-role {
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--text-tertiary);
  margin-bottom: 4px;
  text-transform: uppercase;
}

.user-message .message-role {
  color: var(--accent-color);
}

.message-content {
  padding: 10px 14px;
  border-radius: 10px;
  background: var(--bg-card);
  font-size: 0.875rem;
  line-height: 1.5;
  white-space: pre-wrap;
}

.user-message .message-content {
  background: var(--accent-color);
  color: white;
}

.empty-chat {
  text-align: center;
  padding: 32px 16px;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.empty-chat .hint {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.chat-input-area {
  padding: 12px;
  border-top: 1px solid var(--border-color);
}

.chat-input {
  width: 100%;
  min-height: 60px;
  padding: 10px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-card);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 0.875rem;
  resize: none;
  outline: none;
  margin-bottom: 8px;
}

.chat-input:focus {
  border-color: var(--accent-color);
}

.send-btn {
  width: 100%;
  padding: 10px;
  background: var(--accent-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  font-size: 0.875rem;
  cursor: pointer;
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
