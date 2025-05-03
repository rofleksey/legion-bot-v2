<template>
  <div class="text-input">
    <div v-if="label" class="text-input__label-container">
      <label class="text-input__label">{{ label }}</label>
      <button
        v-if="showHelpIcon"
        class="text-input__help-icon"
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
    <div class="text-input__container">
      <input
        ref="inputRef"
        v-model="localValue"
        type="text"
        class="text-input__field"
        :class="{ 'text-input__field--error': hasError }"
        :placeholder="placeholder"
        @focus="handleFocus"
        @blur="handleBlur"
      />
      <div v-if="hasError" class="text-input__error">
        {{ errorMessage }}
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: string
    label?: string
    placeholder?: string
    maxLength?: number
    minLength?: number
    pattern?: RegExp
    required?: boolean
    showHelpIcon?: boolean // whether to show the help icon
  }>(),
  {
    label: '',
    placeholder: 'Enter text',
    maxLength: undefined,
    minLength: undefined,
    pattern: undefined,
    required: false,
    showHelpIcon: false
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'help-click'): void // new custom event for help icon click
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const isFocused = ref(false)
const localValue = ref(props.modelValue)
const hasError = ref(false)
const errorMessage = ref('')

watch(
  () => props.modelValue,
  (newValue) => {
    if (!isFocused.value) {
      localValue.value = newValue
    }
  }
)

function emitHelpClick() {
  emit('help-click')
}

function handleFocus() {
  isFocused.value = true
}

function handleBlur() {
  isFocused.value = false
  validateInput(localValue.value)
  if (!hasError.value) {
    emit('update:modelValue', localValue.value)
  }
}

function validateInput(value: string) {
  hasError.value = false
  errorMessage.value = ''

  if (props.required && value.trim() === '') {
    hasError.value = true
    errorMessage.value = 'This field is required'
    return
  }

  if (props.minLength !== undefined && value.length < props.minLength) {
    hasError.value = true
    errorMessage.value = `Must be at least ${props.minLength} characters`
    return
  }

  if (props.maxLength !== undefined && value.length > props.maxLength) {
    hasError.value = true
    errorMessage.value = `Must be at most ${props.maxLength} characters`
    return
  }

  if (props.pattern !== undefined && !props.pattern.test(value)) {
    hasError.value = true
    errorMessage.value = 'Invalid format'
    return
  }
}
</script>

<style>
.text-input {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  width: 275px;
}

.text-input__label-container {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.text-input__label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #e2e8f0;
}

.text-input__help-icon {
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

.text-input__help-icon:hover {
  color: #CBD5E1;
}

.text-input__container {
  position: relative;
}

.text-input__field {
  width: 100%;
  padding: 10px 12px;
  font-size: 0.9375rem;
  color: #f8fafc;
  background-color: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.text-input__field:focus {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 2px rgba(124, 58, 237, 0.2);
}

.text-input__field--error {
  border-color: #f43f5e;
}

.text-input__field--error:focus {
  box-shadow: 0 0 0 2px rgba(244, 63, 94, 0.2);
}

.text-input__error {
  margin-top: 6px;
  font-size: 0.8125rem;
  color: #f43f5e;
}
</style>
