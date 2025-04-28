<template>
  <div class="dropdown-container" v-on-click-outside="handleClickOutside">
    <label v-if="label" class="dropdown-label">
      {{ label }}
      <span v-if="required" class="dropdown-required">*</span>
    </label>

    <div
      class="dropdown"
      :class="{
        'dropdown-open': isOpen,
        'dropdown-disabled': disabled,
        'dropdown-error': error
      }"
      @click="toggleDropdown"
    >
      <div class="dropdown-selection">
        <span class="dropdown-selection-text">
          {{ selectedOption || placeholder }}
        </span>
        <span class="dropdown-selection-icon">
          <svg
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M4 6L8 10L12 6"
              stroke="currentColor"
              stroke-width="1.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </span>
      </div>

      <div
        v-show="isOpen"
        class="dropdown-options"
        ref="optionsRef"
      >
        <div
          v-for="(option, index) in options"
          :key="index"
          class="dropdown-option"
          :class="{
            'dropdown-option-selected': option === selectedOption,
            'dropdown-option-highlighted': highlightedIndex === index
          }"
          @click.stop="selectOption(option)"
        >
          {{ option }}
        </div>
      </div>
    </div>

    <div v-if="error" class="dropdown-error-message">
      {{ error }}
    </div>
    <div v-if="helperText" class="dropdown-helper-text">
      {{ helperText }}
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { vOnClickOutside } from '@vueuse/components'

interface Props {
  modelValue?: string | null
  options: string[]
  label?: string
  placeholder?: string
  disabled?: boolean
  required?: boolean
  error?: string
  helperText?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: null,
  placeholder: 'Select an option',
  disabled: false,
  required: false
})

const emit = defineEmits(['update:modelValue'])

const isOpen = ref(false)
const selectedOption = ref<string | null>(props.modelValue || null)
const highlightedIndex = ref(-1)
const optionsRef = ref<HTMLElement | null>(null)

// Handle outside clicks
const handleClickOutside = (event: MouseEvent) => {
  if (optionsRef.value && !optionsRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

// Handle keyboard navigation
const handleKeyDown = (event: KeyboardEvent) => {
  if (!isOpen.value) return

  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      highlightedIndex.value = Math.min(highlightedIndex.value + 1, props.options.length - 1)
      scrollToHighlighted()
      break
    case 'ArrowUp':
      event.preventDefault()
      highlightedIndex.value = Math.max(highlightedIndex.value - 1, 0)
      scrollToHighlighted()
      break
    case 'Enter':
      event.preventDefault()
      if (highlightedIndex.value >= 0) {
        selectOption(props.options[highlightedIndex.value])
      }
      break
    case 'Escape':
      event.preventDefault()
      isOpen.value = false
      break
  }
}

const scrollToHighlighted = () => {
  if (optionsRef.value && highlightedIndex.value >= 0) {
    const options = optionsRef.value.children
    if (options && options[highlightedIndex.value]) {
      options[highlightedIndex.value].scrollIntoView({
        block: 'nearest'
      })
    }
  }
}

const toggleDropdown = () => {
  if (props.disabled) return

  isOpen.value = !isOpen.value
  if (isOpen.value) {
    highlightedIndex.value = props.options.findIndex(opt => opt === selectedOption.value)
  }
}

const selectOption = (option: string) => {
  selectedOption.value = option
  emit('update:modelValue', option)
  isOpen.value = false
}

watch(() => props.modelValue, (newValue) => {
  selectedOption.value = newValue || null
})

onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
})
</script>

<style scoped>
/* Base styles */
.dropdown-container {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  width: 275px;
}

.dropdown-label {
  display: block;
  margin-bottom: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--dropdown-label-color, #e2e8f0);
}

.dropdown-required {
  color: var(--dropdown-error-color, #f87171);
  margin-left: 2px;
}

/* Dropdown main container */
.dropdown {
  position: relative;
  width: 100%;
  min-height: 40px;
  border-radius: 6px;
  background-color: var(--dropdown-bg, #1e293b);
  border: 1px solid var(--dropdown-border, #334155);
  cursor: pointer;
  transition: all 0.2s ease;
}

.dropdown:hover:not(.dropdown-disabled) {
  border-color: var(--dropdown-hover-border, #475569);
}

.dropdown-open {
  border-color: var(--dropdown-active-border, #7c3aed);
  box-shadow: 0 0 0 1px var(--dropdown-active-border, #7c3aed);
}

.dropdown-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.dropdown-error {
  border-color: var(--dropdown-error-color, #f87171);
}

/* Selection area */
.dropdown-selection {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  height: 100%;
}

.dropdown-selection-text {
  font-size: 0.875rem;
  color: var(--dropdown-text-color, #f8fafc);
}

.dropdown-selection-icon {
  color: var(--dropdown-icon-color, #94a3b8);
  transition: transform 0.2s ease;
}

.dropdown-open .dropdown-selection-icon {
  transform: rotate(180deg);
}

/* Options list */
.dropdown-options {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  width: 100%;
  max-height: 240px;
  overflow-y: auto;
  background-color: var(--dropdown-bg, #1e293b);
  border: 1px solid var(--dropdown-border, #334155);
  border-radius: 6px;
  z-index: 50;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

/* Individual options */
.dropdown-option {
  padding: 8px 12px;
  font-size: 0.875rem;
  color: var(--dropdown-text-color, #f8fafc);
  transition: all 0.1s ease;
}

.dropdown-option:hover:not(.dropdown-option-selected) {
  background-color: var(--dropdown-option-hover-bg, #334155);
  cursor: pointer;
}

.dropdown-option-highlighted {
  background-color: var(--dropdown-option-highlight-bg, #475569);
}

.dropdown-option-selected {
  background-color: var(--dropdown-option-selected-bg, #7c3aed);
  color: white;
}

/* Helper text and error messages */
.dropdown-helper-text {
  margin-top: 6px;
  font-size: 0.75rem;
  color: var(--dropdown-helper-color, #94a3b8);
}

.dropdown-error-message {
  margin-top: 6px;
  font-size: 0.75rem;
  color: var(--dropdown-error-color, #f87171);
}
</style>
