<template>
  <div class="notebook-detail">
    <!-- Left Sidebar -->
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <button class="collapse-toggle" @click="sidebarCollapsed = !sidebarCollapsed">
        <PanelLeftOpen v-if="sidebarCollapsed" :size="14" />
        <PanelLeftClose v-else :size="14" />
      </button>

      <div class="sidebar-content" v-show="!sidebarCollapsed">
        <!-- Notebook Header -->
        <div class="notebook-header">
          <div class="breadcrumb">
            <router-link to="/" class="home-link"><Library :size="14" class="inline-icon" /> My Books</router-link>
          </div>
          <div class="notebook-selector">
            <Folder :size="18" class="notebook-icon" />
            <span class="notebook-name" @click="openNotebook" title="Click to edit notebook">{{ currentNotebook?.title || 'Loading...' }}</span>
            <button class="dropdown-btn" @click="showNotebookSwitcher = true" title="Switch Notebook"><ChevronDown :size="12" /></button>
          </div>
        </div>

        <!-- Search Bar -->
        <div class="search-bar" @click="showSearchModal = true">
          <Search :size="14" class="search-icon" />
          <span class="search-placeholder">Search</span>
          <span class="search-shortcut">⌘J</span>
        </div>

        <!-- Quick Actions -->
        <div class="quick-actions">
          <button class="quick-btn" @click="createNewNote" title="New Note"><FilePlus :size="16" /></button>
          <button class="quick-btn" @click="createNewFolderInline" title="New Folder"><FolderPlus :size="16" /></button>
          <button class="quick-btn" @click="refreshData" title="Refresh"><RefreshCw :size="16" /></button>
          <button class="quick-btn" @click="collapseAllFolders" title="Collapse All"><FoldVertical :size="16" /></button>
        </div>

        <!-- File Tree -->
        <nav 
          class="notebook-sidebar"
          id="file-library-tree"
          @click="deselectAll"
          @dragover.prevent="onDragOverRoot"
          @dragleave="onDragLeaveRoot"
          @drop="onDropRoot"
          :class="{ 'drag-over-root': dragOverRoot }"
        >
          <!-- New Folder Input (Root) -->
          <div class="file-tree-node" v-if="isCreatingFolder && !selectedItemId">
            <div class="file-node-content">
              <span class="file-node-icon"><Folder :size="14" /></span>
              <input
                ref="newFolderInput"
                type="text"
                class="inline-input"
                v-model="newFolderName"
                @keyup.enter="confirmNewFolder"
                @keyup.esc="cancelNewFolder"
                @blur="confirmNewFolder"
                placeholder="folder name"
              />
            </div>
          </div>

          <!-- Folders -->
          <div
            class="file-tree-node"
            v-for="folder in folders"
            :key="folder.id"
            :class="{ 
              'file-node-expanded': expandedFolders.has(folder.id),
              'file-node-collapsed': !expandedFolders.has(folder.id)
            }"
          >
            <div
              class="file-node-content"
              :class="{ 'active': selectedItemId === folder.id && selectedItemType === 'folder' }"
              @click.stop="onFolderClick(folder.id)"
              @dragover.prevent="onDragOver($event, folder.id)"
              @dragleave="onDragLeave"
              @drop.stop="onDrop($event, folder.id)"
            >
              <span class="file-node-icon" @click.stop="onFolderClick(folder.id)">
                <ChevronDown v-if="expandedFolders.has(folder.id)" :size="12" /><ChevronRight v-else :size="12" /> <Folder :size="14" />
              </span>
              <input
                v-if="editingFolderId === folder.id"
                type="text"
                class="inline-input"
                v-model="editingFolderName"
                @keyup.enter="confirmRenameFolder(folder.id)"
                @keyup.esc="cancelRenameFolder"
                @blur="confirmRenameFolder(folder.id)"
              />
              <span v-else class="file-node-title">{{ folder.title }}</span>
              <button class="file-node-delete" @click.stop="promptDeleteFolder(folder)" title="Delete folder">
                <Trash2 :size="12" />
              </button>
              <div class="file-node-background"></div>
            </div>
            
            <!-- Folder children -->
            <div class="file-node-children" v-show="expandedFolders.has(folder.id)">
              <!-- New folder inside this folder -->
              <div class="file-tree-node" v-if="isCreatingFolder && selectedItemId === folder.id && selectedItemType === 'folder'">
                <div class="file-node-content">
                  <span class="file-node-icon"><Folder :size="14" /></span>
                  <input
                    ref="newFolderInput"
                    type="text"
                    class="inline-input"
                    v-model="newFolderName"
                    @keyup.enter="confirmNewFolder"
                    @keyup.esc="cancelNewFolder"
                    @blur="confirmNewFolder"
                    placeholder="folder name"
                  />
                </div>
              </div>
              
              <!-- Notes in folder -->
              <div
                class="file-tree-node"
                v-for="note in getNotesByFolder(folder.id)"
                :key="note.id"
                draggable="true"
                @dragstart="onDragStart($event, note)"
                @dragend="onDragEnd"
              >
                <div 
                  class="file-node-content"
                  :class="{ 'active': currentNoteId === note.id }"
                  @click.stop="onNoteClick(note.id)"
                >
                  <span class="file-node-icon"><FileText :size="14" /></span>
                  <span class="file-node-title">{{ note.title || 'Untitled' }}</span>
                  <button class="file-node-delete" @click.stop="promptDeleteNote(note)" title="Delete note">
                    <Trash2 :size="12" />
                  </button>
                  <div class="file-node-background"></div>
                </div>
              </div>
            </div>
          </div>

          <!-- Ungrouped Notes (root level) -->
          <div
            class="file-tree-node"
            v-for="note in ungroupedNotes"
            :key="note.id"
            draggable="true"
            @dragstart="onDragStart($event, note)"
            @dragend="onDragEnd"
          >
            <div 
              class="file-node-content"
              :class="{ 'active': currentNoteId === note.id }"
              @click.stop="onNoteClick(note.id)"
            >
              <span class="file-node-icon"><FileText :size="14" /></span>
              <span class="file-node-title">{{ note.title || 'Untitled' }}</span>
              <button class="file-node-delete" @click.stop="promptDeleteNote(note)" title="Delete note">
                <Trash2 :size="12" />
              </button>
              <div class="file-node-background"></div>
            </div>
          </div>

          <div class="empty-tree" v-if="notes.length === 0 && folders.length === 0 && !isLoading">
            No notes yet
          </div>
        </nav>

        <!-- Sidebar Footer -->
        <div class="sidebar-footer">
          <router-link to="/settings" class="footer-link"><Settings :size="14" class="inline-icon" /> Settings</router-link>
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="main-content">
      <!-- Tab Bar -->
      <div class="tab-bar" v-if="openTabs.length > 0">
        <div
          class="tab"
          :class="{ active: tab.id === currentNoteId }"
          v-for="tab in openTabs"
          :key="tab.id"
          @click="openTabItem(tab)"
        >
          <span class="tab-icon"><NotebookText v-if="tab.type === 'notebook'" :size="12" /><Folder v-else-if="tab.type === 'folder'" :size="12" /><FileText v-else :size="12" /></span>
          <span class="tab-title">{{ tab.title || 'Untitled' }}</span>
          <button class="tab-close" @click.stop="closeTab(tab.id)">×</button>
        </div>
      </div>

      <!-- Note Editor Mode (always visible) -->
      <div class="editor-view" :key="currentNoteId || 'notebook'">
        <header class="editor-header">
          <input
            type="text"
            class="note-title-input"
            v-model="editorTitle"
            placeholder="Untitled"
            @input="markDirty"
            :readonly="editorMode === 'read'"
          />
          <div class="header-actions">
            <span class="save-status">{{ saveStatus }}</span>
            <button 
              class="action-btn mode-toggle" 
              :class="editorMode"
              @click="toggleEditorMode"
            >
              {{ editorMode === 'edit' ? 'Read' : 'Edit' }}
            </button>
            <button class="action-btn secondary" @click="saveCurrentNote" :disabled="isSaving" v-if="editorMode === 'edit'">Save</button>
            <button class="action-btn secondary" @click="closeCurrentTab">Close</button>
          </div>
        </header>

        <div class="editor-container">
          <LakeEditor
            :key="`${currentNoteId || 'notebook'}-${editorVersion}`"
            v-model="editorContent"
            :mode="editorMode"
            @update:modelValue="markDirty"
            placeholder="Start writing..."
          />
        </div>
      </div>
    </main>

    <!-- Notebook Switcher Modal -->
    <SearchModal
      :isOpen="showNotebookSwitcher"
      placeholder="Search notebooks..."
      :items="notebookSearchItems"
      :recentItems="recentNotebooks"
      @close="showNotebookSwitcher = false"
      @select="handleNotebookSelect"
    />

    <!-- Global Search Modal -->
    <SearchModal
      :isOpen="showSearchModal"
      :scope="currentNotebook?.title"
      placeholder="Search notes..."
      :items="noteSearchItems"
      :recentItems="recentNotes"
      @close="showSearchModal = false"
      @select="handleSearchSelect"
    />

    <!-- Delete Confirmation Modal -->
    <div class="modal-overlay" v-if="showDeleteModal" @click="showDeleteModal = false">
      <div class="modal" @click.stop>
        <h3>Confirm Delete</h3>
        <p>Are you sure you want to delete "{{ deleteTargetName }}"?</p>
        <div class="modal-actions">
          <button class="cancel-btn" @click="showDeleteModal = false">Cancel</button>
          <button class="confirm-btn danger" @click="confirmDelete">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { noteApi, type Note } from '../api/notes'
import SearchModal, { type SearchItem } from '../components/SearchModal.vue'
import LakeEditor from '../components/LakeEditor.vue'
import { useTabs, useFileExplorer, useDragDrop, type Tab } from '../composables'
import {
  PanelLeftOpen, PanelLeftClose, Library, Folder, ChevronDown, ChevronRight,
  Search, FilePlus, FolderPlus, RefreshCw, FoldVertical, FileText, Settings, NotebookText, Trash2
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()

// Route params - unified route uses single :id for the opened item
const routeId = computed(() => route.params.id as string)
// The routeId IS the currentNoteId in the new unified route
const currentNoteId = computed(() => routeId.value || null)
// For backward compatibility, notebookId will be determined dynamically
const notebookId = ref<string>('')

// Data
const notebooks = ref<Note[]>([]) // All notebooks (type='notebook')
const notes = ref<Note[]>([]) // Notes in current notebook
const currentNotebook = ref<Note | null>(null)
const isLoading = ref(true)
const sidebarCollapsed = ref(false)

// Modals
const showNotebookSwitcher = ref(false)
const showSearchModal = ref(false)

// Folder state - now folders are just Notes with type='folder'
const folders = ref<Note[]>([])

// Use composables for cleaner state management
const fileExplorer = useFileExplorer({ notes, folders, notebookId })
const {
  selectedItemId,
  selectedItemType,
  selectFolder,
  selectNote,
  deselectAll,
  expandedFolders,
  toggleFolder,
  expandFolder,
  collapseAllFolders,
  ungroupedNotes,
  getNotesByFolder,
  isCreatingFolder,
  newFolderName,
  editingFolderId,
  editingFolderName,
  cancelRenameFolder,
} = fileExplorer

// Template ref for new folder input (needs to be in component)
const newFolderInput = ref<HTMLInputElement | null>(null)

// Tabs composable
const tabs = useTabs({
  onTabChange: (tab) => openTabItem(tab),
  onLastTabClosed: () => openNotebook(),
})
const { openTabs, closeTab: closeTabById } = tabs

// Current item type for editor
const currentItemType = ref<'note' | 'folder' | 'notebook' | null>(null)

// Drag and drop composable
const dragDrop = useDragDrop({ notes, notebookId, expandFolder })
const {
  dragOverRoot,
  onDragStart,
  onDragEnd,
  onDragOver,
  onDragLeave,
  onDragOverRoot,
  onDragLeaveRoot,
  onDrop,
  onDropRoot,
} = dragDrop

// Editor state (kept inline for now due to complex save logic dependencies)
const editorTitle = ref('')
const editorContent = ref('')
const editorMode = ref<'edit' | 'read'>('edit')
const editorVersion = ref(0)
const isSaving = ref(false)
const isDirty = ref(false)
const saveStatus = ref('')
let autoSaveTimer: ReturnType<typeof setTimeout> | null = null

// Delete modal
const showDeleteModal = ref(false)
const deleteTargetId = ref<string | null>(null)
const deleteTargetType = ref<'note' | 'folder'>('note')
const deleteTargetName = ref('')

// Prompt delete functions
const promptDeleteFolder = (folder: Note) => {
  deleteTargetId.value = folder.id
  deleteTargetType.value = 'folder'
  deleteTargetName.value = folder.title || 'Untitled Folder'
  showDeleteModal.value = true
}

const promptDeleteNote = (note: Note) => {
  deleteTargetId.value = note.id
  deleteTargetType.value = 'note'
  deleteTargetName.value = note.title || 'Untitled'
  showDeleteModal.value = true
}

// Tab wrapper function to match old interface
const closeTab = (tabId: string) => {
  closeTabById(tabId, currentNoteId.value)
}


// Click handlers - use composable functions
const onFolderClick = (folderId: string) => {
  // VSCode behavior: select + toggle expand/collapse
  selectFolder(folderId)
  toggleFolder(folderId)
  // Also open folder content in editor
  openFolder(folderId)
}

const onNoteClick = (noteId: string) => {
  selectNote(noteId)
  openNote(noteId)
}

// Open folder in editor (folders can have editable content now)
const openFolder = async (folderId: string) => {
  // Save current item if dirty
  if (isDirty.value && currentNoteId.value) {
    await saveCurrentNote()
  }
  
  // Find folder in local cache
  const folder = folders.value.find(f => f.id === folderId)
  if (!folder) return
  
  // Add to tabs
  if (!openTabs.value.find(t => t.id === folderId)) {
    openTabs.value.push({ id: folder.id, title: folder.title || 'Untitled Folder', type: 'folder' })
  }
  
  // Track current item type
  currentItemType.value = 'folder'
  
  // Load folder content into editor
  editorTitle.value = folder.title || ''
  editorContent.value = folder.content || ''
  isDirty.value = false
  saveStatus.value = ''
  
  // Navigate to folder edit view (reuse note route for now)
  router.push(`/note/${folderId}`)
}

// Open notebook in editor (notebooks can have editable content now)
const openNotebook = async () => {
  // Save current item if dirty
  if (isDirty.value && currentNoteId.value) {
    await saveCurrentNote()
  }
  
  // Use current notebook
  const notebook = currentNotebook.value
  if (!notebook) return
  
  // Add to tabs
  if (!openTabs.value.find(t => t.id === notebook.id)) {
    openTabs.value.push({ id: notebook.id, title: notebook.title || 'Untitled Notebook', type: 'notebook' })
  }
  
  // Track current item type
  currentItemType.value = 'notebook'
  
  // Load notebook content into editor
  editorTitle.value = notebook.title || ''
  editorContent.value = notebook.content || ''
  isDirty.value = false
  saveStatus.value = ''
  
  // Navigate to notebook edit view
  router.push(`/note/${notebook.id}`)
}


const createNewFolderInline = async () => {
  isCreatingFolder.value = true
  newFolderName.value = ''
  // If a folder is selected, expand it
  if (selectedItemType.value === 'folder' && selectedItemId.value) {
    expandedFolders.value.add(selectedItemId.value)
  }
  await nextTick()
  newFolderInput.value?.focus()
}

const confirmNewFolder = async () => {
  if (!notebookId.value) {
    console.warn('Cannot create folder: notebookId not set')
    cancelNewFolder()
    return
  }
  const name = newFolderName.value.trim()
  if (name) {
    try {
      // Determine parent folder
      let parentId: string | undefined
      if (selectedItemType.value === 'folder' && selectedItemId.value) {
        parentId = selectedItemId.value
      } else if (selectedItemType.value === 'note' && selectedItemId.value) {
        // If a note is selected, create folder in the same parent
        const selectedNote = notes.value.find(n => n.id === selectedItemId.value)
        parentId = selectedNote?.parentId || undefined
      }
      
      const newFolder = await noteApi.create({
        type: 'folder',
        title: name,
        parentId: parentId || notebookId.value,
      })
      folders.value.push(newFolder)
      expandedFolders.value.add(newFolder.id)
    } catch (error) {
      console.error('Failed to create folder:', error)
    }
  }
  cancelNewFolder()
}

const cancelNewFolder = () => {
  isCreatingFolder.value = false
  newFolderName.value = ''
}

const confirmRenameFolder = async (folderId: string) => {
  const name = editingFolderName.value.trim()
  if (name) {
    try {
      await noteApi.update(folderId, { title: name })
      const folder = folders.value.find(f => f.id === folderId)
      if (folder) folder.title = name
    } catch (error) {
      console.error('Failed to rename folder:', error)
    }
  }
  cancelRenameFolder()
}



// Note operations
const openNote = async (noteId: string) => {
  // Save current item if dirty
  if (isDirty.value && currentNoteId.value) {
    await saveCurrentNote()
  }
  
  // Add to tabs
  const note = notes.value.find(n => n.id === noteId)
  if (note && !openTabs.value.find(t => t.id === noteId)) {
    openTabs.value.push({ id: note.id, title: note.title || 'Untitled', type: 'note' })
  }
  
  // Track current item type
  currentItemType.value = 'note'
  
  // Navigate
  router.push(`/note/${noteId}`)
}


const createNewNote = async () => {
  if (!notebookId.value) {
    console.warn('Cannot create note: notebookId not set')
    return
  }
  try {
    // Determine target folder based on selection
    let targetFolderId: string | undefined
    
    if (selectedItemType.value === 'folder' && selectedItemId.value) {
      // If a folder is selected, create note in that folder
      targetFolderId = selectedItemId.value
    } else if (selectedItemType.value === 'note' && selectedItemId.value) {
      // If a note is selected, create note in the same folder as that note
      const selectedNote = notes.value.find(n => n.id === selectedItemId.value)
      targetFolderId = selectedNote?.parentId || undefined
    }
    // else: no selection or root selected, create at root
    
    const newNote = await noteApi.create({
      type: 'note',
      title: 'Untitled',
      parentId: targetFolderId || notebookId.value,
      content: '',
    })
    notes.value.unshift(newNote)
    
    if (targetFolderId) {
      expandedFolders.value.add(targetFolderId)
    }
    
    // Select the new note in the tree
    selectNote(newNote.id)
    
    // Open in editor directly (without navigation to avoid loadData)
    if (!openTabs.value.find(t => t.id === newNote.id)) {
      openTabs.value.push({ id: newNote.id, title: newNote.title || 'Untitled', type: 'note' })
    }
    editorTitle.value = newNote.title || ''
    editorContent.value = newNote.content || ''
    currentItemType.value = 'note'
    isDirty.value = false
    saveStatus.value = ''
    editorVersion.value++
    
  } catch (error) {
    console.error('Failed to create note:', error)
  }
}

const markDirty = () => {
  isDirty.value = true
  saveStatus.value = 'Unsaved'
  
  // Auto-save after 2 seconds
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(() => {
    saveCurrentNote()
  }, 2000)
}

const toggleEditorMode = async () => {
  // Save before switching to read mode
  if (editorMode.value === 'edit' && isDirty.value) {
    await saveCurrentNote()
  }
  editorMode.value = editorMode.value === 'edit' ? 'read' : 'edit'
}

const saveCurrentNote = async () => {
  if (!currentNoteId.value || isSaving.value || !isDirty.value) return
  
  isSaving.value = true
  saveStatus.value = 'Saving...'
  
  try {
    await noteApi.update(currentNoteId.value, {
      title: editorTitle.value || 'Untitled',
      content: editorContent.value,
    })
    
    // Update local item (note, folder, or notebook)
    if (currentItemType.value === 'notebook') {
      // Update notebook
      if (currentNotebook.value) {
        currentNotebook.value.title = editorTitle.value || 'Untitled'
        currentNotebook.value.content = editorContent.value
      }
    } else if (currentItemType.value === 'folder') {
      const folderIndex = folders.value.findIndex(f => f.id === currentNoteId.value)
      if (folderIndex !== -1) {
        const folderToUpdate = folders.value[folderIndex]
        if (folderToUpdate) {
          folderToUpdate.title = editorTitle.value || 'Untitled'
          folderToUpdate.content = editorContent.value
        }
      }
    } else {
      const noteIndex = notes.value.findIndex(n => n.id === currentNoteId.value)
      if (noteIndex !== -1) {
        const noteToUpdate = notes.value[noteIndex]
        if (noteToUpdate) {
          noteToUpdate.title = editorTitle.value || 'Untitled'
          noteToUpdate.content = editorContent.value
        }
      }
    }
    
    // Update tab title
    const tab = openTabs.value.find(t => t.id === currentNoteId.value)
    if (tab) tab.title = editorTitle.value || 'Untitled'
    
    isDirty.value = false
    saveStatus.value = 'Saved'
    setTimeout(() => {
      if (saveStatus.value === 'Saved') saveStatus.value = ''
    }, 2000)
  } catch (error) {
    console.error('Failed to save:', error)
    saveStatus.value = 'Error'
  } finally {
    isSaving.value = false
  }
}

// Tab operations
const openTabItem = (tab: Tab) => {
  if (tab.type === 'notebook') {
    openNotebook()
  } else if (tab.type === 'folder') {
    openFolder(tab.id)
  } else {
    openNote(tab.id)
  }
}

const closeCurrentTab = () => {
  if (currentNoteId.value) {
    closeTab(currentNoteId.value)
  }
}


const confirmDelete = async () => {
  if (!deleteTargetId.value) return
  
  if (deleteTargetType.value === 'note') {
    try {
      await noteApi.delete(deleteTargetId.value)
      notes.value = notes.value.filter(n => n.id !== deleteTargetId.value)
      openTabs.value = openTabs.value.filter(t => t.id !== deleteTargetId.value)
      if (currentNoteId.value === deleteTargetId.value) {
        openNotebook()
      }
    } catch (error) {
      console.error('Failed to delete note:', error)
    }
  } else {
    try {
      await noteApi.delete(deleteTargetId.value)
      folders.value = folders.value.filter(f => f.id !== deleteTargetId.value)
      // Notes are moved to root by backend
      notes.value.forEach(note => {
        if (note.parentId === deleteTargetId.value) {
          note.parentId = notebookId.value
        }
      })
    } catch (error) {
      console.error('Failed to delete folder:', error)
    }
  }
  
  showDeleteModal.value = false
  deleteTargetId.value = null
}

// Search
const notebookSearchItems = computed<SearchItem[]>(() => {
  return notebooks.value.map(nb => ({
    id: nb.id,
    type: 'notebook' as const,
    title: nb.title,
    meta: ''
  }))
})

const recentNotebooks = computed<SearchItem[]>(() => {
  return notebooks.value.slice(0, 5).map(nb => ({
    id: nb.id,
    type: 'notebook' as const,
    title: nb.title,
    meta: ''
  }))
})

const noteSearchItems = computed<SearchItem[]>(() => {
  return notes.value.map(note => ({
    id: note.id,
    type: 'note' as const,
    title: note.title || 'Untitled',
    preview: note.content?.slice(0, 100),
    meta: formatDate(note.updatedAt)
  }))
})

const recentNotes = computed<SearchItem[]>(() => {
  return notes.value.slice(0, 5).map(note => ({
    id: note.id,
    type: 'note' as const,
    title: note.title || 'Untitled',
    meta: formatDate(note.updatedAt)
  }))
})

const handleNotebookSelect = (item: SearchItem) => {
  router.push(`/notebook/${item.id}`)
}

const handleSearchSelect = (item: SearchItem) => {
  if (item.type === 'note') {
    openNote(item.id)
  }
}

// Utilities
const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const noteDate = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  
  if (noteDate.getTime() === today.getTime()) {
    return `Today ${date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}`
  }
  return date.toLocaleDateString('en-US', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

const refreshData = () => {
  loadData()
}

const loadData = async () => {
  isLoading.value = true
  try {
    // First, get all notebooks
    const notebooksData = await noteApi.getNotebooks()
    notebooks.value = notebooksData
    
    // Get the item we're opening
    const openedItem = await noteApi.getById(routeId.value)
    
    if (!openedItem) {
      console.error('Item not found:', routeId.value)
      return
    }
    
    // Determine the notebook context
    if (openedItem.type === 'notebook') {
      // Opening a notebook directly
      notebookId.value = openedItem.id
      currentNotebook.value = openedItem
      currentItemType.value = 'notebook'
    } else {
      // Opening a note or folder - find its parent notebook
      let parentId = openedItem.parentId
      let parentItem = parentId ? await noteApi.getById(parentId) : null
      
      // Traverse up to find the notebook
      while (parentItem && parentItem.type !== 'notebook') {
        parentId = parentItem.parentId
        parentItem = parentId ? await noteApi.getById(parentId) : null
      }
      
      if (parentItem && parentItem.type === 'notebook') {
        notebookId.value = parentItem.id
        currentNotebook.value = parentItem
      } else {
        // Fallback: use the first notebook or the item's direct parent
        notebookId.value = openedItem.parentId || notebooksData[0]?.id || ''
        currentNotebook.value = notebooksData.find(n => n.id === notebookId.value) || null
      }
      
      currentItemType.value = openedItem.type as 'note' | 'folder'
    }
    
    // Load children of the notebook (including nested children in folders)
    if (notebookId.value) {
      const directChildren = await noteApi.getChildren(notebookId.value)
      folders.value = directChildren.filter(n => n.type === 'folder')
      const directNotes = directChildren.filter(n => n.type === 'note')
      
      // Also load children of each folder
      const allFolderNotes: Note[] = []
      for (const folder of folders.value) {
        try {
          const folderChildren = await noteApi.getChildren(folder.id)
          allFolderNotes.push(...folderChildren.filter(n => n.type === 'note'))
        } catch (error) {
          console.error(`Failed to load children for folder ${folder.id}:`, error)
        }
      }
      
      // Combine direct notes and folder notes
      notes.value = [...directNotes, ...allFolderNotes]
    }
    
    // Auto-open the item if not a notebook or if the notebook is being opened
    if (openedItem.type === 'notebook' && openTabs.value.length === 0) {
      openNotebook()
    } else if (openedItem.type !== 'notebook') {
      // Load the opened note/folder content
      editorTitle.value = openedItem.title || ''
      editorContent.value = openedItem.content || ''
      isDirty.value = false
      saveStatus.value = ''
      editorVersion.value++ // Force editor remount with new content
      
      // Add to tabs if not already there
      if (!openTabs.value.find(t => t.id === openedItem.id)) {
        openTabs.value.push({
          id: openedItem.id,
          title: openedItem.title || 'Untitled',
          type: openedItem.type as 'note' | 'folder',
        })
      }
      
      // Restore selection state for file explorer highlighting
      selectedItemId.value = openedItem.id
      selectedItemType.value = openedItem.type as 'folder' | 'note'
      
      // Auto-expand parent folder if the item is inside a folder
      if (openedItem.parentId && openedItem.parentId !== notebookId.value) {
        expandedFolders.value.add(openedItem.parentId)
      }
    }
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    isLoading.value = false
  }
}

// Keyboard shortcuts
const handleKeydown = (e: KeyboardEvent) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'j') {
    e.preventDefault()
    showSearchModal.value = true
  }
  if ((e.metaKey || e.ctrlKey) && e.key === 's') {
    e.preventDefault()
    if (currentNoteId.value) {
      saveCurrentNote()
    }
  }
}

// Watchers
// Watch route changes to reload data
watch(routeId, (newId, oldId) => {
  if (newId && newId !== oldId) {
    loadData()
  }
})

// Lifecycle
onMounted(() => {
  loadData()
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
})
</script>

<style scoped>
.notebook-detail {
  display: flex;
  height: 100vh;
  background: var(--bg-primary);
}

/* Sidebar */
.sidebar {
  position: relative;
  width: 260px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  transition: width 0.2s;
}

.sidebar.collapsed {
  width: 40px;
}

.collapse-toggle {
  position: absolute;
  right: -12px;
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 24px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 50%;
  font-size: 0.625rem;
  color: var(--text-tertiary);
  cursor: pointer;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
}

.collapse-toggle:hover {
  color: var(--text-primary);
  border-color: var(--accent-color);
}

.sidebar-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.notebook-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  /* Fixed height: breadcrumb(~18px) + gap(8px) + selector(~24px) + padding(24px) = ~74px total */
  min-height: 74px;
  box-sizing: border-box;
}

.breadcrumb {
  margin-bottom: 8px;
}

.home-link {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  text-decoration: none;
}

.home-link:hover {
  color: var(--text-secondary);
}

.notebook-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}

.notebook-icon {
  font-size: 1.125rem;
}

.notebook-name {
  flex: 1;
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.9375rem;
  cursor: pointer;
}

.notebook-name:hover {
  color: var(--accent-color);
}

.dropdown-btn {
  background: none;
  border: none;
  font-size: 0.625rem;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}

.dropdown-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 12px 16px 8px;
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

.quick-actions {
  display: flex;
  gap: 4px;
  padding: 0 16px 8px;
  border-bottom: 1px solid var(--border-color);
}

.quick-btn {
  flex: 1;
  padding: 6px;
  background: transparent;
  border: none;
  border-radius: 4px;
  font-size: 0.875rem;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: all 0.15s;
}

.quick-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

/* File Tree */
.file-tree {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
  min-height: 100px;
}

.file-tree.drag-over-root {
  background: rgba(99, 102, 241, 0.1);
}

.folder-item {
  transition: background 0.15s;
}

.folder-item.drag-over {
  background: rgba(99, 102, 241, 0.2);
}

.folder-item.selected > .tree-item.folder {
  background: rgba(99, 102, 241, 0.15);
  border-left: 3px solid var(--accent-color);
  margin-left: -3px;
  box-shadow: inset 0 0 0 1px rgba(99, 102, 241, 0.2);
}

.folder-item.selected > .tree-item.folder .tree-label {
  color: var(--accent-color);
  font-weight: 500;
}

.tree-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  color: var(--text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: background 0.1s;
  position: relative;
}

.tree-item:hover {
  background: var(--bg-hover);
}

.tree-item.active {
  background: var(--accent-color);
  color: white;
}

.tree-item.note.selected {
  background: rgba(99, 102, 241, 0.15);
  border-left: 3px solid var(--accent-color);
  margin-left: -3px;
}

.tree-item.note.selected .tree-label {
  color: var(--accent-color);
  font-weight: 500;
}

.tree-item.editing {
  background: var(--bg-hover);
}

.tree-item.folder {
  padding-left: 12px;
}

.tree-item.note {
  padding-left: 24px;
}

.folder-children .tree-item.note {
  padding-left: 36px;
}

.expand-icon {
  font-size: 0.5rem;
  width: 12px;
  text-align: center;
}

.tree-icon {
  font-size: 0.875rem;
  flex-shrink: 0;
}

.tree-label {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.inline-input {
  flex: 1;
  background: var(--bg-primary);
  border: 1px solid var(--accent-color);
  border-radius: 3px;
  padding: 2px 6px;
  font-size: 0.8125rem;
  color: var(--text-primary);
  outline: none;
}

.tree-action {
  opacity: 0;
  background: none;
  border: none;
  font-size: 0.75rem;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 2px;
}

.tree-item:hover .tree-action {
  opacity: 1;
}

.tree-action:hover {
  color: var(--error-color);
}

.empty-tree {
  padding: 16px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 0.8125rem;
}

.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
}

.footer-link {
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.875rem;
}

.footer-link:hover {
  color: var(--text-primary);
}

/* Tab Bar */
.tab-bar {
  display: flex;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  overflow-x: auto;
  min-height: 30px;
  max-height: 30px;
  flex-shrink: 0;
  scrollbar-width: thin;
  scrollbar-color: var(--text-tertiary) transparent;
}

.tab-bar::-webkit-scrollbar {
  height: 2px;
}

.tab-bar::-webkit-scrollbar-track {
  background: transparent;
}

.tab-bar::-webkit-scrollbar-thumb {
  background-color: var(--text-tertiary);
  border-radius: 3px;
}

.tab {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: transparent;
  border-right: 1px solid var(--border-color);
  cursor: pointer;
  white-space: nowrap;
  font-size: 0.6875rem;
  color: var(--text-secondary);
  line-height: 1;
}

.tab:hover {
  background: var(--bg-hover);
}

.tab.active {
  background: var(--bg-primary);
  color: var(--text-primary);
}

.tab-title {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  background: none;
  border: none;
  font-size: 0.875em;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 0;
  line-height: 1;
  opacity: 0.6;
}

.tab-close:hover {
  color: var(--text-primary);
  opacity: 1;
}

/* Tab Icons */
.tab-icon {
  display: flex;
  align-items: center;
  color: var(--text-tertiary);
}

/* Inline Icon */
.inline-icon {
  display: inline-block;
  vertical-align: middle;
  margin-right: 4px;
}

/* Main Content */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.overview-view {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 32px;
  border-bottom: 1px solid var(--border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.notebook-icon-large {
  font-size: 2.5rem;
}

.header-info h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
}

.notebook-stats {
  margin-top: 4px;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.stat-divider {
  margin: 0 8px;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.action-btn {
  padding: 10px 20px;
  background: var(--accent-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.action-btn:hover {
  opacity: 0.9;
}

.action-btn.secondary {
  background: var(--bg-tertiary, #f0f0f0);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  padding: 6px 14px;
}

.action-btn.secondary:hover {
  background: var(--bg-secondary);
}

/* Mode Toggle Button */
.action-btn.mode-toggle {
  background: var(--bg-secondary);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
  font-weight: 500;
  min-width: 60px;
  padding: 6px 14px;
}

.action-btn.mode-toggle.edit {
  background: var(--bg-tertiary, #f0f0f0);
  color: var(--text-primary);
  border-color: var(--border-color);
}

.action-btn.mode-toggle.read {
  background: var(--bg-tertiary, #f0f0f0);
  color: var(--text-primary);
  border-color: var(--border-color);
}

/* Edit Notebook Button */
.edit-notebook-btn {
  background: transparent;
  border: none;
  font-size: 0.75rem;
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 4px;
  opacity: 0.6;
  transition: opacity 0.15s;
}

.edit-notebook-btn:hover {
  opacity: 1;
  background: var(--bg-hover);
}

/* Tab Icons */
.tab-icon {
  font-size: 0.875rem;
  margin-right: 4px;
}

.content-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px 32px;
}

.notes-table {
  max-width: 800px;
}

.note-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px dashed var(--border-color);
  cursor: pointer;
  transition: background 0.2s;
}

.note-row:hover {
  background: var(--bg-hover);
  margin: 0 -16px;
  padding-left: 16px;
  padding-right: 16px;
}

.note-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.note-icon {
  font-size: 1rem;
}

.note-name {
  font-size: 0.9375rem;
  color: var(--text-primary);
}

.note-date {
  font-size: 0.8125rem;
  color: var(--text-tertiary);
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.create-btn {
  margin-top: 16px;
  padding: 10px 20px;
  background: var(--accent-color);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.editor-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 16px;
  border-bottom: 1px solid var(--border-color);
  gap: 8px;
  flex-shrink: 0;
  /* Height = 74px (sidebar header) - 30px (tab bar) = 44px (strict alignment) */
  height: 44px;
  max-height: 44px;
  box-sizing: border-box;
}

.note-title-input {
  flex: 1;
  font-size: 0.875rem;
  font-weight: 500;
  border: none;
  background: transparent;
  color: var(--text-primary);
  outline: none;
  padding: 0.25em 0;
}

.note-title-input::placeholder {
  color: var(--text-tertiary);
}

.save-status {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

.editor-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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

.modal p {
  margin: 0 0 16px;
  color: var(--text-secondary);
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

.confirm-btn.danger {
  background: var(--error-color);
}

/* Notion Sidebar Styles */
#file-library-tree {
  padding-top: 0.5714em;
  padding-left: 0.8571em;
  padding-right: 0.8571em;
  box-sizing: border-box;
  font-size: 14px;
  background: #fafafa;
  color: #37352f;
  height: 100%;
  overflow-y: auto;
}

.file-tree-node {
  position: relative;
  padding-left: 0;
  margin-left: 0;
}

.file-node-content {
  position: relative;
  display: flex;
  align-items: center;
  padding: 2px 0 2px 10px;
  cursor: pointer;
  z-index: 2;
  min-height: 28px;
  margin-left: -10px;
  padding-right: 10px;
}

.file-node-background {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  bottom: 0;
  z-index: -1;
  border-radius: 3px;
}

.file-node-content:hover .file-node-background {
  background: #ededef;
}

.file-node-content.active .file-node-background {
  background: #ededef;
  font-weight: 600;
}

/* Delete button on file nodes */
.file-node-delete {
  display: none;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 2px;
  border-radius: 3px;
  margin-left: auto;
  flex-shrink: 0;
  position: relative;
  z-index: 5;
}

.file-node-content:hover .file-node-delete {
  display: flex;
}

.file-node-delete:hover {
  color: var(--error-color);
  background: rgba(239, 68, 68, 0.1);
}

.file-node-icon {
  margin-right: 4px;
  color: #999;
  font-size: 13px;
  white-space: nowrap;
  flex-shrink: 0;
}

.file-node-title {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #37352f;
}

.file-node-children {
  margin-left: 20px;
  position: relative;
}

/* Tree connection lines */
.file-node-children::before {
  content: "";
  position: absolute;
  left: -9px;
  top: 0;
  bottom: 0;
  width: 1px;
  background: #e1e7e8;
  opacity: 0.5;
}

.inline-input {
  flex: 1;
  border: 1px solid #2eaadc;
  border-radius: 3px;
  padding: 2px 4px;
  font-size: inherit;
  outline: none;
}
</style>
