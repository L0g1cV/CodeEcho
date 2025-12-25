import { ref, type Ref } from 'vue'
import { noteApi, type Note } from '../api/notes'

export interface UseDragDropOptions {
    notes: Ref<Note[]>
    notebookId: Ref<string>
    expandFolder: (folderId: string) => void
}

/**
 * Composable for managing drag and drop functionality in the file tree.
 * Handles dragging notes between folders and to root level.
 */
export function useDragDrop(options: UseDragDropOptions) {
    const { notes, notebookId, expandFolder } = options

    const draggedNoteId = ref<string | null>(null)
    const dragOverFolderId = ref<string | null>(null)
    const dragOverRoot = ref(false)

    /**
     * Start dragging a note
     */
    const onDragStart = (e: DragEvent, note: Note) => {
        draggedNoteId.value = note.id
        if (e.dataTransfer) {
            e.dataTransfer.effectAllowed = 'move'
            e.dataTransfer.setData('text/plain', note.id)
        }
    }

    /**
     * End dragging
     */
    const onDragEnd = () => {
        draggedNoteId.value = null
        dragOverFolderId.value = null
        dragOverRoot.value = false
    }

    /**
     * Drag over a folder
     */
    const onDragOver = (_e: DragEvent, folderId: string) => {
        dragOverFolderId.value = folderId
        dragOverRoot.value = false
    }

    /**
     * Leave a folder
     */
    const onDragLeave = () => {
        dragOverFolderId.value = null
    }

    /**
     * Drag over root area
     */
    const onDragOverRoot = (_e: DragEvent) => {
        if (draggedNoteId.value && !dragOverFolderId.value) {
            dragOverRoot.value = true
        }
    }

    /**
     * Leave root area
     */
    const onDragLeaveRoot = () => {
        dragOverRoot.value = false
    }

    /**
     * Drop note into a folder
     */
    const onDrop = async (e: DragEvent, folderId: string) => {
        e.preventDefault()
        e.stopPropagation()

        if (draggedNoteId.value) {
            const noteId = draggedNoteId.value
            const noteIndex = notes.value.findIndex(n => n.id === noteId)
            const originalNote = noteIndex !== -1 ? notes.value[noteIndex] : null

            // Update UI immediately for responsiveness
            if (originalNote) {
                notes.value[noteIndex] = { ...originalNote, parentId: folderId }
            }
            expandFolder(folderId)

            try {
                await noteApi.update(noteId, { parentId: folderId })
            } catch (error) {
                console.error('Failed to move note:', error)
                // Revert on error
                if (originalNote) {
                    notes.value[noteIndex] = originalNote
                }
            }
        }
        onDragEnd()
    }

    /**
     * Drop note to root level
     */
    const onDropRoot = async (e: DragEvent) => {
        e.preventDefault()

        if (!notebookId.value) {
            console.warn('Cannot drop to root: notebookId not set')
            onDragEnd()
            return
        }

        if (draggedNoteId.value) {
            const noteId = draggedNoteId.value
            const noteIndex = notes.value.findIndex(n => n.id === noteId)
            const originalNote = noteIndex !== -1 ? notes.value[noteIndex] : null

            // Update UI immediately
            if (originalNote) {
                notes.value[noteIndex] = { ...originalNote, parentId: notebookId.value }
            }

            try {
                await noteApi.update(noteId, { parentId: notebookId.value })
            } catch (error) {
                console.error('Failed to move note:', error)
                // Revert on error
                if (originalNote) {
                    notes.value[noteIndex] = originalNote
                }
            }
        }
        onDragEnd()
    }

    return {
        // State
        draggedNoteId,
        dragOverFolderId,
        dragOverRoot,

        // Event handlers
        onDragStart,
        onDragEnd,
        onDragOver,
        onDragLeave,
        onDragOverRoot,
        onDragLeaveRoot,
        onDrop,
        onDropRoot,
    }
}
