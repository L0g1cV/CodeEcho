import { ref, computed, type Ref } from 'vue'
import type { Note } from '../api/notes'

export interface UseFileExplorerOptions {
    notes: Ref<Note[]>
    folders: Ref<Note[]>
    notebookId: Ref<string>
}

/**
 * Composable for managing file explorer selection and folder state.
 * Handles VSCode-style selection highlighting and folder expand/collapse.
 */
export function useFileExplorer(options: UseFileExplorerOptions) {
    const { notes, folders, notebookId } = options

    // Selection state (VSCode-style: only one item highlighted)
    const selectedItemId = ref<string | null>(null)
    const selectedItemType = ref<'folder' | 'note' | null>(null)

    // Folder state
    const expandedFolders = ref(new Set<string>())

    // Folder editing state
    const isCreatingFolder = ref(false)
    const newFolderName = ref('')
    const editingFolderId = ref<string | null>(null)
    const editingFolderName = ref('')

    // Computed: notes at root level (not inside folders)
    const ungroupedNotes = computed(() => {
        return notes.value.filter(note => !note.parentId || note.parentId === notebookId.value)
    })

    /**
     * Get notes inside a specific folder
     */
    const getNotesByFolder = (folderId: string): Note[] => {
        return notes.value.filter(note => note.parentId === folderId)
    }

    /**
     * Select a folder (for highlighting)
     */
    const selectFolder = (folderId: string) => {
        selectedItemId.value = folderId
        selectedItemType.value = 'folder'
    }

    /**
     * Select a note (for highlighting)
     */
    const selectNote = (noteId: string) => {
        selectedItemId.value = noteId
        selectedItemType.value = 'note'
    }

    /**
     * Deselect all items (clicking empty space)
     */
    const deselectAll = () => {
        selectedItemId.value = null
        selectedItemType.value = null
    }

    /**
     * Toggle folder expand/collapse
     */
    const toggleFolder = (folderId: string) => {
        if (expandedFolders.value.has(folderId)) {
            expandedFolders.value.delete(folderId)
        } else {
            expandedFolders.value.add(folderId)
        }
    }

    /**
     * Expand a folder
     */
    const expandFolder = (folderId: string) => {
        expandedFolders.value.add(folderId)
    }

    /**
     * Collapse all folders
     */
    const collapseAllFolders = () => {
        expandedFolders.value.clear()
        deselectAll()
    }

    /**
     * Get the parent folder ID for creating new items
     * Returns the selected folder ID, or the parent of the selected note
     */
    const getTargetParentId = (): string | undefined => {
        if (selectedItemType.value === 'folder' && selectedItemId.value) {
            return selectedItemId.value
        }
        if (selectedItemType.value === 'note' && selectedItemId.value) {
            const selectedNote = notes.value.find(n => n.id === selectedItemId.value)
            return selectedNote?.parentId || undefined
        }
        return undefined
    }

    /**
     * Start creating a new folder
     */
    const startCreateFolder = () => {
        isCreatingFolder.value = true
        newFolderName.value = ''
        // If a folder is selected, expand it
        if (selectedItemType.value === 'folder' && selectedItemId.value) {
            expandedFolders.value.add(selectedItemId.value)
        }
    }

    /**
     * Cancel folder creation
     */
    const cancelCreateFolder = () => {
        isCreatingFolder.value = false
        newFolderName.value = ''
    }

    /**
     * Start renaming a folder
     */
    const startRenameFolder = (folderId: string) => {
        const folder = folders.value.find(f => f.id === folderId)
        if (folder) {
            editingFolderId.value = folderId
            editingFolderName.value = folder.title
        }
    }

    /**
     * Cancel folder rename
     */
    const cancelRenameFolder = () => {
        editingFolderId.value = null
        editingFolderName.value = ''
    }

    return {
        // Selection state
        selectedItemId,
        selectedItemType,
        selectFolder,
        selectNote,
        deselectAll,

        // Folder state
        expandedFolders,
        toggleFolder,
        expandFolder,
        collapseAllFolders,

        // Notes helpers
        ungroupedNotes,
        getNotesByFolder,
        getTargetParentId,

        // Folder editing
        isCreatingFolder,
        newFolderName,
        editingFolderId,
        editingFolderName,
        startCreateFolder,
        cancelCreateFolder,
        startRenameFolder,
        cancelRenameFolder,
    }
}
