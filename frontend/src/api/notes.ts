import api from './client'

// Types
export type NoteType = 'notebook' | 'folder' | 'note'

export interface Note {
    id: string
    userId: string
    parentId?: string
    type: NoteType
    title: string
    content: string
    problemUrl?: string
    solutionCode?: string
    createdAt: string
    updatedAt: string
    children?: Note[]
}

export interface NotesResponse {
    total: number
    items: Note[]
}

export interface CreateNoteRequest {
    type: NoteType
    title: string
    parentId?: string
    content?: string
    problemUrl?: string
    solutionCode?: string
}

export interface UpdateNoteRequest {
    title?: string
    parentId?: string
    content?: string
    problemUrl?: string
    solutionCode?: string
}

// Note APIs (unified for notebooks, folders, and notes)
export const noteApi = {
    // Get all notes with optional filtering
    getAll: async (params?: {
        type?: NoteType
        parentId?: string | 'null' | 'root'
        search?: string
        page?: number
        pageSize?: number
    }) => {
        const { data } = await api.get<NotesResponse>('/notes', { params })
        return data
    },

    // Get single note by ID
    getById: async (id: string) => {
        const { data } = await api.get<Note>(`/notes/${id}`)
        return data
    },

    // Create a new note (notebook, folder, or note)
    create: async (req: CreateNoteRequest) => {
        const { data } = await api.post<Note>('/notes', req)
        return data
    },

    // Update a note
    update: async (id: string, req: UpdateNoteRequest) => {
        const { data } = await api.put<Note>(`/notes/${id}`, req)
        return data
    },

    // Delete a note (and children for notebooks/folders)
    delete: async (id: string) => {
        await api.delete(`/notes/${id}`)
    },

    // Helper: Get all notebooks
    getNotebooks: async () => {
        const { data } = await api.get<NotesResponse>('/notes', {
            params: { type: 'notebook', parentId: 'null' }
        })
        return data.items
    },

    // Helper: Get items inside a notebook or folder
    getChildren: async (parentId: string) => {
        const { data } = await api.get<NotesResponse>('/notes', {
            params: { parentId }
        })
        return data.items
    },

    // Helper: Create a notebook
    createNotebook: async (title: string, content?: string) => {
        return noteApi.create({ type: 'notebook', title, content })
    },

    // Helper: Create a folder inside a parent
    createFolder: async (parentId: string, title: string, content?: string) => {
        return noteApi.create({ type: 'folder', title, parentId, content })
    },

    // Helper: Create a note inside a parent
    createNote: async (parentId: string, title: string, content?: string) => {
        return noteApi.create({ type: 'note', title, parentId, content })
    },
}
