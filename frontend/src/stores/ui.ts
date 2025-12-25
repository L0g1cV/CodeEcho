import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUIStore = defineStore('ui', () => {
    const zenMode = ref(false)
    const sidebarWidth = ref(280)
    const aiPanelWidth = ref(400)
    const theme = ref<'light' | 'dark'>('light')

    const toggleZenMode = () => {
        zenMode.value = !zenMode.value
    }

    const setTheme = (newTheme: 'light' | 'dark') => {
        theme.value = newTheme
        document.documentElement.setAttribute('data-theme', newTheme)
    }

    const setSidebarWidth = (width: number) => {
        sidebarWidth.value = Math.max(200, Math.min(400, width))
    }

    const setAIPanelWidth = (width: number) => {
        aiPanelWidth.value = Math.max(300, Math.min(600, width))
    }

    return {
        zenMode,
        sidebarWidth,
        aiPanelWidth,
        theme,
        toggleZenMode,
        setTheme,
        setSidebarWidth,
        setAIPanelWidth,
    }
})
