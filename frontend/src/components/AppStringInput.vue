<template>
  <div class="text-input">
    <label v-if="label" class="text-input__label">{{ label }}</label>
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
  }>(),
  {
    label: '',
    placeholder: 'Enter text',
    maxLength: undefined,
    minLength: undefined,
    pattern: undefined,
    required: false
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
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

.text-input__label {
  display: block;
  margin-bottom: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  color: #e2e8f0;
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
