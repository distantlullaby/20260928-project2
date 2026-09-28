<template>
  <Transition name="fade">
    <div v-if="toast.visible" :class="['toast', toast.type]">{{ toast.text }}</div>
  </Transition>
</template>

<script setup>
import { toast } from './toast'
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
