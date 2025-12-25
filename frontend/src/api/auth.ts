import api from './client'

// Types
export interface User {
    id: string
    username: string
    createdAt: string
}

export interface AuthResponse {
    token: string
    user: User
}

export interface LLMConfig {
    provider: 'openai' | 'claude' | 'deepseek' | 'custom'
    apiKey: string
    baseUrl?: string
    model: string
}

// Auth APIs
export const authApi = {
    register: async (username: string, password: string) => {
        const { data } = await api.post<User>('/auth/register', { username, password })
        return data
    },

    login: async (username: string, password: string) => {
        const { data } = await api.post<AuthResponse>('/auth/login', { username, password })
        return data
    },

    getCurrentUser: async () => {
        const { data } = await api.get<User>('/users/me')
        return data
    },

    updateLLMConfig: async (config: LLMConfig) => {
        const { data } = await api.put('/users/me/llm-config', config)
        return data
    },
}
