import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface LLMConfig {
    provider: 'openai' | 'claude' | 'deepseek' | 'custom'
    apiKey: string
    baseUrl?: string
    model: string
}

export interface User {
    id: string
    username: string
    createdAt: string
}

export const useUserStore = defineStore('user', () => {
    const user = ref<User | null>(null)
    const token = ref<string | null>(localStorage.getItem('token'))
    const llmConfig = ref<LLMConfig | null>(null)

    const isLoggedIn = () => !!token.value

    const setUser = (newUser: User) => {
        user.value = newUser
    }

    const setToken = (newToken: string) => {
        token.value = newToken
        localStorage.setItem('token', newToken)
    }

    const setLLMConfig = (config: LLMConfig) => {
        llmConfig.value = config
    }

    const logout = () => {
        user.value = null
        token.value = null
        llmConfig.value = null
        localStorage.removeItem('token')
    }

    return {
        user,
        token,
        llmConfig,
        isLoggedIn,
        setUser,
        setToken,
        setLLMConfig,
        logout,
    }
})
