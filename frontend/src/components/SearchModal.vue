<template>
  <Teleport to="body">
    <div class="search-modal-overlay" v-if="isOpen" @click="close">
      <div class="search-modal" @click.stop>
        <div class="search-header">
          <span class="search-icon"><Search :size="18" /></span>
          <span class="search-path" v-if="scope">{{ scope }} /</span>
          <input
            ref="searchInputRef"
            type="text"
            v-model="searchQuery"
            :placeholder="placeholder"
            class="search-input"
            @keydown.esc="close"
            @keydown.enter="selectFirst"
            @keydown.down.prevent="moveSelection(1)"
            @keydown.up.prevent="moveSelection(-1)"
          />
          <button class="close-btn" @click="close"><X :size="18" /></button>
        </div>

        <div class="search-results">
          <!-- Recent / Recommended -->
          <div class="result-section" v-if="recentItems.length > 0 && !searchQuery">
            <div class="section-title">Recent</div>
            <div
              class="result-item"
              :class="{ selected: selectedIndex === idx }"
              v-for="(item, idx) in recentItems"
              :key="item.id"
              @click="selectItem(item)"
              @mouseenter="selectedIndex = idx"
            >
              <span class="item-icon"><Folder v-if="item.type === 'notebook' || item.type === 'folder'" :size="16" /><FileText v-else :size="16" /></span>
              <span class="item-title">{{ item.title }}</span>
              <span class="item-meta">{{ item.meta }}</span>
            </div>
          </div>

          <!-- Search Results -->
          <div class="result-section" v-if="searchQuery && filteredResults.length > 0">
            <div class="section-title">Search results</div>
            <div
              class="result-item"
              :class="{ selected: selectedIndex === idx }"
              v-for="(item, idx) in filteredResults"
              :key="item.id"
              @click="selectItem(item)"
              @mouseenter="selectedIndex = idx"
            >
              <span class="item-icon"><Folder v-if="item.type === 'notebook' || item.type === 'folder'" :size="16" /><FileText v-else :size="16" /></span>
              <div class="item-content">
                <span class="item-title" v-html="highlightMatch(item.title)"></span>
                <span class="item-preview" v-if="item.preview">{{ item.preview }}</span>
              </div>
              <span class="item-meta">{{ item.meta }}</span>
            </div>
          </div>

          <!-- Empty State -->
          <div class="empty-state" v-if="searchQuery && filteredResults.length === 0">
            No results for "{{ searchQuery }}"
          </div>

          <!-- Initial State -->
          <div class="empty-state" v-if="!searchQuery && recentItems.length === 0">
            Start typing to search...
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { Search, X, Folder, FileText } from 'lucide-vue-next'

export interface SearchItem {
  id: string
  type: 'notebook' | 'note' | 'folder'
  title: string
  preview?: string
  meta?: string
  notebookId?: string
}

const props = defineProps<{
  isOpen: boolean
  scope?: string
  placeholder?: string
  items: SearchItem[]
  recentItems?: SearchItem[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', item: SearchItem): void
}>()

const searchInputRef = ref<HTMLInputElement | null>(null)
const searchQuery = ref('')
const selectedIndex = ref(0)

const filteredResults = computed(() => {
  if (!searchQuery.value.trim()) return []
  const query = searchQuery.value.toLowerCase()
  return props.items.filter(item =>
    item.title.toLowerCase().includes(query) ||
    item.preview?.toLowerCase().includes(query)
  ).slice(0, 10)
})

const highlightMatch = (text: string) => {
  if (!searchQuery.value.trim()) return text
  const query = searchQuery.value.toLowerCase()
  const idx = text.toLowerCase().indexOf(query)
  if (idx === -1) return text
  return text.slice(0, idx) + 
    `<mark>${text.slice(idx, idx + query.length)}</mark>` + 
    text.slice(idx + query.length)
}

const close = () => {
  searchQuery.value = ''
  selectedIndex.value = 0
  emit('close')
}

const selectItem = (item: SearchItem) => {
  emit('select', item)
  close()
}

const selectFirst = () => {
  const items = searchQuery.value ? filteredResults.value : (props.recentItems || [])
  if (items.length > 0 && selectedIndex.value < items.length) {
    selectItem(items[selectedIndex.value])
  }
}

const moveSelection = (delta: number) => {
  const items = searchQuery.value ? filteredResults.value : (props.recentItems || [])
  const newIndex = selectedIndex.value + delta
  if (newIndex >= 0 && newIndex < items.length) {
    selectedIndex.value = newIndex
  }
}

watch(() => props.isOpen, async (isOpen) => {
  if (isOpen) {
    await nextTick()
    searchInputRef.value?.focus()
  }
})

watch(searchQuery, () => {
  selectedIndex.value = 0
})
</script>

<style scoped>
.search-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 10vh;
  z-index: 1000;
}

.search-modal {
  width: 100%;
  max-width: 600px;
  background: var(--bg-card);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  overflow: hidden;
}

.search-header {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  gap: 8px;
}

.search-icon {
  font-size: 1.125rem;
  color: var(--text-tertiary);
}

.search-path {
  font-size: 0.875rem;
  color: var(--text-secondary);
  white-space: nowrap;
}

.search-input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 1rem;
  color: var(--text-primary);
  outline: none;
}

.search-input::placeholder {
  color: var(--text-tertiary);
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-tertiary);
  font-size: 1rem;
  cursor: pointer;
  padding: 4px;
}

.close-btn:hover {
  color: var(--text-primary);
}

.search-results {
  max-height: 400px;
  overflow-y: auto;
}

.result-section {
  padding: 8px 0;
}

.section-title {
  padding: 8px 16px;
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.result-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.1s;
}

.result-item:hover,
.result-item.selected {
  background: var(--bg-hover);
}

.item-icon {
  font-size: 1rem;
  flex-shrink: 0;
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-title {
  font-size: 0.9375rem;
  color: var(--text-primary);
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-title :deep(mark) {
  background: var(--accent-color);
  color: white;
  padding: 0 2px;
  border-radius: 2px;
}

.item-preview {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-top: 2px;
}

.item-meta {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.empty-state {
  padding: 32px 16px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 0.875rem;
}
</style>
