import { defineStore } from 'pinia'
import { ref } from 'vue'
import { type Note } from '../api/notes'

// Re-export for convenience
export type { Note }

export const useNoteStore = defineStore('note', () => {
    const currentNote = ref<Note | null>(null)
    const isDirty = ref(false)

    const setCurrentNote = (note: Note) => {
        currentNote.value = note
        isDirty.value = false
    }

    const updateContent = (content: string) => {
        if (currentNote.value) {
            currentNote.value.content = content
            isDirty.value = true
        }
    }

    const updateSolutionCode = (code: string) => {
        if (currentNote.value) {
            currentNote.value.solutionCode = code
            isDirty.value = true
        }
    }

    const clearNote = () => {
        currentNote.value = null
        isDirty.value = false
    }

    return {
        currentNote,
        isDirty,
        setCurrentNote,
        updateContent,
        updateSolutionCode,
        clearNote,
    }
})
