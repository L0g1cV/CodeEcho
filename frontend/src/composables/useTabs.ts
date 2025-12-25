import { ref, type Ref } from 'vue'

export interface Tab {
    id: string
    title: string
    type: 'note' | 'folder' | 'notebook'
}

export interface UseTabsOptions {
    onTabChange?: (tab: Tab) => void
    onLastTabClosed?: () => void
}

/**
 * Composable for managing editor tabs.
 * Handles opening, closing, and switching between tabs.
 */
export function useTabs(options: UseTabsOptions = {}) {
    const { onTabChange, onLastTabClosed } = options

    const openTabs = ref<Tab[]>([])

    /**
     * Add a new tab (if not already open)
     */
    const addTab = (tab: Tab) => {
        if (!openTabs.value.find(t => t.id === tab.id)) {
            openTabs.value.push(tab)
        }
    }

    /**
     * Close a tab by ID
     */
    const closeTab = (tabId: string, currentId: string | null) => {
        openTabs.value = openTabs.value.filter(t => t.id !== tabId)

        // If closing current tab, switch to another or notify
        if (currentId === tabId) {
            const nextTab = openTabs.value[0]
            if (nextTab) {
                onTabChange?.(nextTab)
            } else {
                onLastTabClosed?.()
            }
        }
    }

    /**
     * Update tab title
     */
    const updateTabTitle = (tabId: string, title: string) => {
        const tab = openTabs.value.find(t => t.id === tabId)
        if (tab) {
            tab.title = title
        }
    }

    /**
     * Get tab by ID
     */
    const getTab = (tabId: string): Tab | undefined => {
        return openTabs.value.find(t => t.id === tabId)
    }

    /**
     * Check if tab exists
     */
    const hasTab = (tabId: string): boolean => {
        return openTabs.value.some(t => t.id === tabId)
    }

    return {
        openTabs,
        addTab,
        closeTab,
        updateTabTitle,
        getTab,
        hasTab,
    }
}
