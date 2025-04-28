<template>
  <div class="chance-input">
    <label v-if="label" class="chance-input__label">{{ label }}</label>
    <div class="chance-input__container">
      <div class="chance-input__field-wrapper">
        <input
          ref="inputRef"
          v-model="displayValue"
          type="text"
          class="chance-input__field"
          :class="{ 'chance-input__field--error': hasError }"
          placeholder="0-100"
          @focus="handleFocus"
          @blur="handleBlur"
          @keydown="handleKeyDown"
        />
      </div>
      <div v-if="hasError" class="chance-input__error">
        {{ errorMessage }}
      </div>
      <div class="chance-input__percentage">🎲</div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: number // 0-1
    label?: string
  }>(),
  {
    label: ''
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const isFocused = ref(false)
const localValue = ref('')
const hasError = ref(false)
const errorMessage = ref('')

// Convert 0-1 to 0-100 for display
const displayValue = computed({
  get() {
    if (isFocused.value) {
      return localValue.value
    }
    return Math.round(props.modelValue * 100).toString()
  },
  set(value: string) {
    localValue.value = value
    const parsed = parseChance(value)
    if (!isNaN(parsed)) {
      validateChance(parsed)
      if (!hasError.value) {
        emit('update:modelValue', parsed / 100)
      }
    } else if (value === '') {
      hasError.value = false
    }
  }
})

watch(
  () => props.modelValue,
  (newValue) => {
    if (!isFocused.value) {
      localValue.value = Math.round(newValue * 100).toString()
    }
  }
)

function handleFocus() {
  isFocused.value = true
  localValue.value = Math.round(props.modelValue * 100).toString()
}

function handleBlur() {
  isFocused.value = false
  const parsed = parseChance(localValue.value)
  if (!isNaN(parsed)) {
    validateChance(parsed)
    if (!hasError.value) {
      emit('update:modelValue', parsed / 100)
      localValue.value = parsed.toString()
    }
  } else {
    localValue.value = Math.round(props.modelValue * 100).toString()
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

  e.preventDefault()
}

function parseChance(value: string): number {
  const parsed = parseInt(value, 10)
  return isNaN(parsed) ? NaN : Math.min(100, Math.max(0, parsed))
}

function validateChance(value: number) {
  hasError.value = false
  errorMessage.value = ''

  if (isNaN(value)) {
    hasError.value = true
    errorMessage.value = 'Please enter a valid number'
    return
  }

  if (value < 0) {
    hasError.value = true
    errorMessage.value = 'Value must be at least 0'
    return
  }

  if (value > 100) {
    hasError.value = true
    errorMessage.value = 'Value must be at most 100'
    return
  }
}
</script>

<style>
.chance-input {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  max-width: 300px;
}

.chance-input__label {
  display: block;
  margin-bottom: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  color: #e2e8f0;
}

.chance-input__container {
  position: relative;
}

.chance-input__field-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.chance-input__field {
  width: 100%;
  padding: 10px 40px 10px 12px;
  font-size: 0.9375rem;
  color: #f8fafc;
  background-color: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.chance-input__field:focus {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 2px rgba(124, 58, 237, 0.2);
}

.chance-input__field--error {
  border-color: #f43f5e;
}

.chance-input__field--error:focus {
  box-shadow: 0 0 0 2px rgba(244, 63, 94, 0.2);
}

.chance-input__percentage {
  position: absolute;
  right: 10px;
  top: 10px;
  color: #94a3b8;
  pointer-events: none;
}

.chance-input__error {
  margin-top: 6px;
  font-size: 0.8125rem;
  color: #f43f5e;
}
</style>
