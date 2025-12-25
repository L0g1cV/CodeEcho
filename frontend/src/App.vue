<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useUIStore } from './stores/ui'

const uiStore = useUIStore()

onMounted(() => {
  // Initialize theme from store
  document.documentElement.setAttribute('data-theme', uiStore.theme)
})
</script>
