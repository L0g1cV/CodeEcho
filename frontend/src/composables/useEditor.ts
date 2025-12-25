import { ref, type Ref, type ComputedRef } from 'vue'
import { noteApi } from '../api/notes'

export interface EditorState {
    title: Ref<string>
    content: Ref<string>
    mode: Ref<'edit' | 'read'>
    version: Ref<number>
    isSaving: Ref<boolean>
    isDirty: Ref<boolean>
    saveStatus: Ref<string>
}

export interface UseEditorOptions {
    currentItemId: ComputedRef<string | null>
    onSaveSuccess?: (id: string, title: string, content: string) => void
}


/**
 * Composable for managing the Markdown editor state.
 * Handles editing, saving, auto-save, and mode switching.
 */
export function useEditor(options: UseEditorOptions) {
    const { currentItemId, onSaveSuccess } = options

    // Editor state
    const editorTitle = ref('')
    const editorContent = ref('')
    const editorMode = ref<'edit' | 'read'>('edit')
    const editorVersion = ref(0) // Increment to force editor remount
    const isSaving = ref(false)
    const isDirty = ref(false)
    const saveStatus = ref('')

    let autoSaveTimer: ReturnType<typeof setTimeout> | null = null

    /**
     * Load content into the editor
     */
    const loadContent = (title: string, content: string) => {
        editorTitle.value = title
        editorContent.value = content
        isDirty.value = false
        saveStatus.value = ''
        editorVersion.value++
    }

    /**
     * Reset editor to empty state
     */
    const resetEditor = () => {
        editorTitle.value = ''
        editorContent.value = ''
        isDirty.value = false
        saveStatus.value = ''
    }

    /**
     * Mark editor content as dirty (unsaved changes)
     */
    const markDirty = () => {
        isDirty.value = true
        saveStatus.value = 'Unsaved'

        // Auto-save after 2 seconds
        if (autoSaveTimer) clearTimeout(autoSaveTimer)
        autoSaveTimer = setTimeout(() => {
            saveCurrentItem()
        }, 2000)
    }

    /**
     * Toggle between edit and read mode
     */
    const toggleEditorMode = async () => {
        // Save before switching to read mode
        if (editorMode.value === 'edit' && isDirty.value) {
            await saveCurrentItem()
        }
        editorMode.value = editorMode.value === 'edit' ? 'read' : 'edit'
    }

    /**
     * Save the current item
     */
    const saveCurrentItem = async () => {
        if (!currentItemId.value || isSaving.value || !isDirty.value) return

        isSaving.value = true
        saveStatus.value = 'Saving...'

        try {
            await noteApi.update(currentItemId.value, {
                title: editorTitle.value || 'Untitled',
                content: editorContent.value,
            })

            // Notify parent of successful save
            onSaveSuccess?.(
                currentItemId.value,
                editorTitle.value || 'Untitled',
                editorContent.value
            )

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

    /**
     * Cleanup function for unmount
     */
    const cleanup = () => {
        if (autoSaveTimer) clearTimeout(autoSaveTimer)
    }

    return {
        // State
        editorTitle,
        editorContent,
        editorMode,
        editorVersion,
        isSaving,
        isDirty,
        saveStatus,

        // Actions
        loadContent,
        resetEditor,
        markDirty,
        toggleEditorMode,
        saveCurrentItem,
        cleanup,
    }
}
