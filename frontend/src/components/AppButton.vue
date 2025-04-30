<template>
  <button
    class="app-button"
    :disabled="props.loading"
  >
    <slot />
  </button>
</template>

<script setup lang="ts">
interface Props {
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
})
</script>

<style scoped>
.app-button {
  padding: 8px 16px;
  background-color: var(--primary);
  color: var(--primary-foreground);
  border: none;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.app-button:hover {
  background-color: var(--primary-dark);
}

.app-button:disabled {
  background-color: var(--muted);
  cursor: not-allowed;
}
</style>
