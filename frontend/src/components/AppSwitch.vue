<template>
  <div class="toggle-container">
    <div v-if="label" class="toggle-label-container">
      <label class="toggle-label">{{ label }}</label>
      <button
        v-if="showHelpIcon"
        class="toggle-help-icon"
        @click="emitHelpClick"
        aria-label="Help"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 17.5228 6.47715 22 12 22Z" stroke="#94A3B8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M12 16V12" stroke="#94A3B8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M12 8H12.01" stroke="#94A3B8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </button>
    </div>
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
  label?: string;
  showHelpIcon?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  label: '',
  showHelpIcon: false
});

interface Emits {
  (e: 'update:modelValue', value: boolean): void;
  (e: 'help-click'): void;
}

const emit = defineEmits<Emits>();

function toggle() {
  emit('update:modelValue', !props.modelValue);
}

function emitHelpClick() {
  emit('help-click');
}
</script>

<style scoped>
.toggle-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.toggle-label-container {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toggle-label {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  font-size: 0.875rem;
  font-weight: 500;
  color: #e2e8f0;
  user-select: none;
}

.toggle-help-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  color: #94A3B8;
  transition: color 0.2s;
}

.toggle-help-icon:hover {
  color: #CBD5E1;
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
