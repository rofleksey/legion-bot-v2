<template>
  <div class="toggle-container">
    <label v-if="label" class="toggle-label">{{ label }}</label>
    <div
      class="toggle-switch"
      :class="{ 'toggle-switch--active': modelValue }"
      @click="toggle"
      role="switch"
      :aria-checked="modelValue"
    >
      <div class="toggle-switch__thumb"></div>
    </div>
  </div>
</template>

<script setup lang="ts">

interface Props {
  modelValue: boolean;
  label: string;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'update:modelValue', value: boolean): void;
}

const emit = defineEmits<Emits>();

function toggle() {
  emit('update:modelValue', !props.modelValue);
}
</script>

<style scoped>
.toggle-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
}

.toggle-label {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  font-size: 0.875rem;
  font-weight: 500;
  color: #e2e8f0;
  user-select: none;
}

.toggle-switch {
  position: relative;
  width: 3.25rem;
  height: 1.75rem;
  border-radius: 9999px;
  background-color: #3e4c5e;
  transition: background-color 0.2s ease-in-out;
  cursor: pointer;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.2);
}

.toggle-switch--active {
  background-color: #4f46e5 !important;
}

.toggle-switch__thumb {
  position: absolute;
  top: 0.1875rem;
  left: 0.1875rem;
  width: 1.375rem;
  height: 1.375rem;
  border-radius: 50%;
  background-color: #f8fafc;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  transition: transform 0.2s ease-in-out;
}

.toggle-switch--active .toggle-switch__thumb {
  transform: translateX(1.5rem);
}
</style>
