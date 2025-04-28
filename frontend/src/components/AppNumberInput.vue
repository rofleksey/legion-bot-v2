<template>
  <div class="number-input">
    <div v-if="label" class="number-input__label-container">
      <label class="number-input__label">{{ label }}</label>
      <button
        v-if="showHelpIcon"
        class="number-input__help-icon"
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
    <div class="number-input__container">
      <input
        ref="inputRef"
        v-model="displayValue"
        type="text"
        class="number-input__field"
        :class="{ 'number-input__field--error': hasError }"
        :placeholder="placeholder"
        @focus="handleFocus"
        @blur="handleBlur"
        @keydown="handleKeyDown"
      />
      <div v-if="hasError" class="number-input__error">
        {{ errorMessage }}
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: number
    label?: string
    min?: number
    max?: number
    placeholder?: string
    step?: number
    format?: boolean // whether to add thousands separators
    showHelpIcon?: boolean // whether to show the help icon
  }>(),
  {
    label: '',
    min: Number.MIN_SAFE_INTEGER,
    max: Number.MAX_SAFE_INTEGER,
    placeholder: 'Enter a number',
    step: 1,
    format: false,
    showHelpIcon: false
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
  (e: 'help-click'): void // new custom event for help icon click
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const isFocused = ref(false)
const localValue = ref('')
const hasError = ref(false)
const errorMessage = ref('')

const displayValue = computed({
  get() {
    if (isFocused.value) {
      return localValue.value
    }
    return props.format ? formatNumber(props.modelValue) : props.modelValue.toString()
  },
  set(value: string) {
    localValue.value = value
    const parsed = parseNumber(value)
    if (!isNaN(parsed)) {
      validateNumber(parsed)
      if (!hasError.value) {
        emit('update:modelValue', parsed)
      }
    } else if (value === '') {
      hasError.value = false
    }
  },
})

watch(
  () => props.modelValue,
  (newValue) => {
    if (!isFocused.value) {
      localValue.value = props.format ? formatNumber(newValue) : newValue.toString()
    }
  }
)

function emitHelpClick() {
  emit('help-click')
}

function handleFocus() {
  isFocused.value = true
  localValue.value = props.modelValue.toString()
}

function handleBlur() {
  isFocused.value = false
  const parsed = parseNumber(localValue.value)
  if (!isNaN(parsed)) {
    validateNumber(parsed)
    if (!hasError.value) {
      emit('update:modelValue', parsed)
      localValue.value = props.format ? formatNumber(parsed) : parsed.toString()
    }
  } else {
    localValue.value = props.format ? formatNumber(props.modelValue) : props.modelValue.toString()
    hasError.value = false
  }
}

function handleKeyDown(e: KeyboardEvent) {
  // Allow navigation keys, backspace, delete, tab
  if (
    [
      'ArrowLeft',
      'ArrowRight',
      'ArrowUp',
      'ArrowDown',
      'Backspace',
      'Delete',
      'Tab',
      'Home',
      'End',
    ].includes(e.key)
  ) {
    return
  }

  // Allow numbers
  if (/^\d$/.test(e.key)) {
    return
  }

  // Allow minus sign only at start
  if (e.key === '-' && (e.target as HTMLInputElement).selectionStart === 0) {
    return
  }

  // Allow decimal if step requires it
  if (e.key === '.' && props.step % 1 !== 0) {
    const currentValue = (e.target as HTMLInputElement).value
    if (!currentValue.includes('.')) {
      return
    }
  }

  e.preventDefault()
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat(undefined, {
    maximumFractionDigits: props.step % 1 === 0 ? 0 : 2
  }).format(value)
}

function parseNumber(value: string): number {
  // Remove thousands separators while parsing
  const cleanValue = value.replace(/,/g, '')
  const parsed = parseFloat(cleanValue)
  return isNaN(parsed) ? NaN : Math.round(parsed / props.step) * props.step
}

function validateNumber(value: number) {
  hasError.value = false
  errorMessage.value = ''

  if (value < props.min) {
    hasError.value = true
    errorMessage.value = `Value must be at least ${props.min}`
    return
  }

  if (value > props.max) {
    hasError.value = true
    errorMessage.value = `Value must be at most ${props.max}`
    return
  }

  if (!Number.isInteger(value / props.step)) {
    hasError.value = true
    errorMessage.value = `Value must be a multiple of ${props.step}`
    return
  }
}
</script>

<style>
.number-input {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  width: 275px;
}

.number-input__label-container {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.number-input__label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #e2e8f0;
}

.number-input__help-icon {
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

.number-input__help-icon:hover {
  color: #CBD5E1;
}

.number-input__container {
  position: relative;
}

.number-input__field {
  width: 100%;
  padding: 10px 12px;
  font-size: 0.9375rem;
  color: #f8fafc;
  background-color: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.number-input__field:focus {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 2px rgba(124, 58, 237, 0.2);
}

.number-input__field--error {
  border-color: #f43f5e;
}

.number-input__field--error:focus {
  box-shadow: 0 0 0 2px rgba(244, 63, 94, 0.2);
}

.number-input__error {
  margin-top: 6px;
  font-size: 0.8125rem;
  color: #f43f5e;
}
</style>
