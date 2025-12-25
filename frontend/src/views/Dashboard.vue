<template>
  <div class="dashboard">
    <aside class="sidebar">
      <div class="sidebar-header">
        <h1 class="logo">CodeEcho</h1>
        <p class="tagline">Whisper code. Echo clarity.</p>
      </div>
      
      <!-- Search Bar -->
      <div class="search-bar" @click="showSearchModal = true">
        <Search :size="14" class="search-icon" />
        <span class="search-placeholder">Search</span>
        <span class="search-shortcut">⌘J</span>
      </div>

      <nav class="notebook-list">
        <div
          class="notebook-item"
          :class="{ active: selectedNotebookId === null }"
          @click="selectedNotebookId = null"
        >
          <span class="notebook-name"><Library :size="14" class="inline-icon" /> All Notes</span>
        </div>
        <div
          class="notebook-item"
          :class="{ active: selectedNotebookId === notebook.id }"
          v-for="notebook in notebooks"
          :key="notebook.id"
          @click="openNotebook(notebook.id)"
        >
          <span class="notebook-name"><Folder :size="14" class="inline-icon" /> {{ notebook.title }}</span>
        </div>
        <button class="add-notebook-btn" @click="showNewNotebook = true">+ New Notebook</button>
      </nav>
      <div class="sidebar-footer">
        <router-link to="/settings" class="settings-link"><Settings :size="14" class="inline-icon" /> Settings</router-link>
        <button class="logout-btn" @click="handleLogout">Logout</button>
      </div>
    </aside>

    <main class="main-content">
      <header class="main-header">
        <h2>{{ selectedNotebook?.title || 'All Notes' }}</h2>
        <button class="new-note-btn" @click="createNewNote">+ New Note</button>
      </header>

      <div class="notes-grid" v-if="!isLoading">
        <div
          class="note-card"
          v-for="note in filteredNotes"
          :key="note.id"
          @click="openNote(note.id)"
        >
          <h3 class="note-title">{{ note.title || 'Untitled' }}</h3>
          <p class="note-preview">{{ getPreview(note.content) }}</p>
          <span class="note-date">{{ formatDate(note.updatedAt) }}</span>
        </div>
        <div class="empty-state" v-if="filteredNotes.length === 0">
          <p>No notes yet. Create your first note!</p>
        </div>
      </div>

      <div class="loading" v-else>Loading...</div>
    </main>

    <!-- New Notebook Modal -->
    <div class="modal-overlay" v-if="showNewNotebook" @click="showNewNotebook = false">
      <div class="modal" @click.stop>
        <h3>New Notebook</h3>
        <input
          type="text"
          v-model="newNotebookName"
          placeholder="Notebook name"
          @keyup.enter="createNotebook"
        />
        <div class="modal-actions">
          <button class="cancel-btn" @click="showNewNotebook = false">Cancel</button>
          <button class="confirm-btn" @click="createNotebook">Create</button>
        </div>
      </div>
    </div>

    <!-- Global Search Modal -->
    <SearchModal
      :isOpen="showSearchModal"
      placeholder="Search content, or type > to get all commands"
      :items="searchItems"
      :recentItems="recentItems"
      @close="showSearchModal = false"
      @select="handleSearchSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { noteApi, type Note } from '../api/notes'
import SearchModal, { type SearchItem } from '../components/SearchModal.vue'
import { Search, Library, Folder, Settings } from 'lucide-vue-next'

const router = useRouter()
const userStore = useUserStore()

const notebooks = ref<Note[]>([]) // Type = 'notebook'
const notes = ref<Note[]>([]) // Type = 'note'
const selectedNotebookId = ref<string | null>(null)
const isLoading = ref(true)
const showNewNotebook = ref(false)
const newNotebookName = ref('')
const showSearchModal = ref(false)

const selectedNotebook = computed(() => {
  if (!selectedNotebookId.value) return null
  return notebooks.value.find(n => n.id === selectedNotebookId.value)
})

const filteredNotes = computed(() => {
  if (!selectedNotebookId.value) return notes.value
  return notes.value.filter(n => n.parentId === selectedNotebookId.value)
})

const searchItems = computed<SearchItem[]>(() => {
  const items: SearchItem[] = []
  
  // Add notebooks
  notebooks.value.forEach(nb => {
    items.push({
      id: nb.id,
      type: 'notebook',
      title: nb.title,
      meta: 'notebook'
    })
  })
  
  // Add notes
  notes.value.forEach(note => {
    const notebook = notebooks.value.find(nb => nb.id === note.parentId)
    items.push({
      id: note.id,
      type: 'note',
      title: note.title || 'Untitled',
      preview: note.content?.slice(0, 100),
      meta: notebook?.title || ''
    })
  })
  
  return items
})

const recentItems = computed<SearchItem[]>(() => {
  return notes.value.slice(0, 5).map(note => {
    const notebook = notebooks.value.find(nb => nb.id === note.parentId)
    return {
      id: note.id,
      type: 'note' as const,
      title: note.title || 'Untitled',
      meta: notebook?.title || ''
    }
  })
})

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) return 'Today'
  if (days === 1) return 'Yesterday'
  if (days < 7) return `${days} days ago`
  return date.toLocaleDateString()
}

const getPreview = (content: string) => {
  if (!content) return 'No content'
  return content.replace(/[#*`\[\]]/g, '').slice(0, 100) + (content.length > 100 ? '...' : '')
}

const openNote = (id: string) => {
  router.push(`/note/${id}`)
}

const openNotebook = (id: string) => {
  router.push(`/note/${id}`)
}

const handleSearchSelect = (item: SearchItem) => {
  if (item.type === 'notebook') {
    router.push(`/note/${item.id}`)
  } else {
    router.push(`/note/${item.id}`)
  }
}

const createNewNote = async () => {
  try {
    const defaultNotebookId = selectedNotebookId.value || notebooks.value[0]?.id
    if (!defaultNotebookId) {
      const notebook = await noteApi.createNotebook('My Notebook')
      notebooks.value.push(notebook)
      router.push(`/note/new?notebookId=${notebook.id}`)
    } else {
      router.push(`/note/new?notebookId=${defaultNotebookId}`)
    }
  } catch (error) {
    console.error('Failed to create note:', error)
  }
}

const createNotebook = async () => {
  if (!newNotebookName.value.trim()) return

  try {
    const notebook = await noteApi.createNotebook(newNotebookName.value)
    notebooks.value.push(notebook)
    newNotebookName.value = ''
    showNewNotebook.value = false
  } catch (error) {
    console.error('Failed to create notebook:', error)
  }
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}

// Keyboard shortcut for search
const handleKeydown = (e: KeyboardEvent) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'j') {
    e.preventDefault()
    showSearchModal.value = true
  }
}

onMounted(async () => {
  document.addEventListener('keydown', handleKeydown)
  try {
    const [notebooksData, notesData] = await Promise.all([
      noteApi.getNotebooks(),
      noteApi.getAll({ type: 'note' }),
    ])
    notebooks.value = notebooksData
    notes.value = notesData.items
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    isLoading.value = false
  }
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.dashboard {
  display: flex;
  height: 100vh;
  background: var(--bg-primary);
}

.sidebar {
  width: 280px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 24px;
  border-bottom: 1px solid var(--border-color);
}

.logo {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.tagline {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 4px 0 0;
}

/* Search Bar */
.search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 12px 16px;
  padding: 8px 12px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.2s;
}

.search-bar:hover {
  border-color: var(--accent-color);
}

.search-icon {
  font-size: 0.875rem;
  color: var(--text-tertiary);
}

.search-placeholder {
  flex: 1;
  font-size: 0.875rem;
  color: var(--text-tertiary);
}

.search-shortcut {
  font-size: 0.6875rem;
  color: var(--text-tertiary);
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
}

.notebook-list {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.notebook-item {
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.notebook-item:hover {
  background: var(--bg-hover);
}

.notebook-item.active {
  background: var(--accent-color);
  color: white;
}

.add-notebook-btn {
  width: 100%;
  padding: 12px;
  margin-top: 8px;
  border: 1px dashed var(--border-color);
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.add-notebook-btn:hover {
  border-color: var(--accent-color);
  color: var(--accent-color);
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.settings-link {
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.875rem;
}

.settings-link:hover {
  color: var(--text-primary);
}

.logout-btn {
  padding: 6px 12px;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
}

.logout-btn:hover {
  border-color: var(--error-color);
  color: var(--error-color);
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.main-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 32px;
  border-bottom: 1px solid var(--border-color);
}

.main-header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: var(--text-primary);
}

.new-note-btn {
  padding: 10px 20px;
  background: var(--accent-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.new-note-btn:hover {
  opacity: 0.9;
}

.notes-grid {
  flex: 1;
  padding: 24px 32px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
  overflow-y: auto;
  align-content: start;
}

.note-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.2s;
}

.note-card:hover {
  border-color: var(--accent-color);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.note-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 8px;
}

.note-preview {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0 0 12px;
  line-height: 1.5;
}

.note-date {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px;
  color: var(--text-secondary);
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  color: var(--text-secondary);
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: var(--bg-card);
  border-radius: 12px;
  padding: 24px;
  width: 100%;
  max-width: 400px;
}

.modal h3 {
  margin: 0 0 16px;
  color: var(--text-primary);
}

.modal input {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 0.9375rem;
  outline: none;
  margin-bottom: 16px;
}

.modal input:focus {
  border-color: var(--accent-color);
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.cancel-btn {
  padding: 10px 20px;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
}

.confirm-btn {
  padding: 10px 20px;
  background: var(--accent-color);
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
}

/* Inline Icon */
.inline-icon {
  display: inline-block;
  vertical-align: middle;
  margin-right: 4px;
}
</style>
