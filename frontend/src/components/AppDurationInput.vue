<template>
  <div class="duration-input">
    <label v-if="label" class="duration-input__label">{{ label }}</label>
    <div class="duration-input__container">
      <input
        ref="inputRef"
        v-model="displayValue"
        type="text"
        class="duration-input__field"
        :class="{ 'duration-input__field--error': hasError }"
        :placeholder="placeholder"
        @focus="handleFocus"
        @blur="handleBlur"
        @keydown="handleKeyDown"
      />
      <div v-if="hasError" class="duration-input__error">
        {{ errorMessage }}
      </div>
      <div class="duration-input__time-symbol">⏳</div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'

const MIN_DURATION_NS = 1_000_000 // 1ms in nanoseconds
const MAX_DURATION_NS = 356_400_000_000_000 // 99h in nanoseconds (99 * 60 * 60 * 1e9)

const props = withDefaults(
  defineProps<{
    modelValue: number
    label?: string
    min?: number
    max?: number
    placeholder?: string
  }>(),
  {
    label: '',
    min: MIN_DURATION_NS,
    max: MAX_DURATION_NS,
    placeholder: 'HH:mm:ss.zzz',
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

const displayValue = computed({
  get() {
    if (isFocused.value) {
      return localValue.value
    }
    return formatDuration(props.modelValue)
  },
  set(value: string) {
    localValue.value = value
    const parsed = parseDuration(value)
    if (parsed !== null) {
      validateDuration(parsed)
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
      localValue.value = formatDuration(newValue)
    }
  }
)

function handleFocus() {
  isFocused.value = true
  localValue.value = formatDuration(props.modelValue)
}

function handleBlur() {
  isFocused.value = false
  // Validate and format on blur
  const parsed = parseDuration(localValue.value)
  if (parsed !== null) {
    validateDuration(parsed)
    if (!hasError.value) {
      emit('update:modelValue', parsed)
      localValue.value = formatDuration(parsed)
    }
  } else {
    // Revert to last valid value
    localValue.value = formatDuration(props.modelValue)
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

  // Allow : and . at appropriate positions
  const currentPos = (e.target as HTMLInputElement).selectionStart
  if (currentPos !== null) {
    if (
      (e.key === ':' && [2, 5].includes(currentPos)) ||
      (e.key === '.' && currentPos === 8)
    ) {
      return
    }
  }

  e.preventDefault()
}

function formatDuration(ns: number): string {
  // Ensure ns is within safe integer range and positive
  ns = Math.max(0, Math.min(ns, Number.MAX_SAFE_INTEGER))

  // Constants in nanoseconds
  const NS_PER_HOUR = 3_600_000_000_000
  const NS_PER_MINUTE = 60_000_000_000
  const NS_PER_SECOND = 1_000_000_000
  const NS_PER_MILLISECOND = 1_000_000

  // Calculate each component
  const hours = Math.floor(ns / NS_PER_HOUR)
  const remainingAfterHours = ns % NS_PER_HOUR

  const minutes = Math.floor(remainingAfterHours / NS_PER_MINUTE)
  const remainingAfterMinutes = remainingAfterHours % NS_PER_MINUTE

  const seconds = Math.floor(remainingAfterMinutes / NS_PER_SECOND)
  const remainingAfterSeconds = remainingAfterMinutes % NS_PER_SECOND

  const milliseconds = Math.floor(remainingAfterSeconds / NS_PER_MILLISECOND)

  // Format with leading zeros
  return [
    hours.toString().padStart(2, '0'),
    minutes.toString().padStart(2, '0'),
    seconds.toString().padStart(2, '0')
  ].join(':') + `.${milliseconds.toString().padStart(3, '0')}`
}

function parseDuration(value: string): number | null {
  if (!value) return null

  const regex = /^(\d{1,2}):(\d{2}):(\d{2})\.(\d{3})$/
  const match = value.match(regex)
  if (!match) return null

  const [, hoursStr, minutesStr, secondsStr, msStr] = match
  const hours = Number(hoursStr)
  const minutes = Number(minutesStr)
  const seconds = Number(secondsStr)
  const milliseconds = Number(msStr)

  return (
    hours * 3_600_000_000_000 +
    minutes * 60_000_000_000 +
    seconds * 1_000_000_000 +
    milliseconds * 1_000_000
  )
}

function validateDuration(ns: number) {
  hasError.value = false
  errorMessage.value = ''

  if (ns < props.min) {
    hasError.value = true
    errorMessage.value = `Duration must be at least ${formatDuration(props.min)}`
    return
  }

  if (ns > props.max) {
    hasError.value = true
    errorMessage.value = `Duration must be at most ${formatDuration(props.max)}`
    return
  }

  if (ns < MIN_DURATION_NS) {
    hasError.value = true
    errorMessage.value = `Minimum duration is ${formatDuration(MIN_DURATION_NS)}`
    return
  }

  if (ns > MAX_DURATION_NS) {
    hasError.value = true
    errorMessage.value = `Maximum duration is ${formatDuration(MAX_DURATION_NS)}`
    return
  }
}
</script>

<style>
.duration-input {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  max-width: 300px;
}

.duration-input__label {
  display: block;
  margin-bottom: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  color: #e2e8f0;
}

.duration-input__container {
  position: relative;
}

.duration-input__field {
  width: 100%;
  padding: 10px 12px;
  font-size: 0.9375rem;
  font-family: monospace;
  color: #f8fafc;
  background-color: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.duration-input__field:focus {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 2px rgba(124, 58, 237, 0.2);
}

.duration-input__field--error {
  border-color: #f43f5e;
}

.duration-input__field--error:focus {
  box-shadow: 0 0 0 2px rgba(244, 63, 94, 0.2);
}

.duration-input__error {
  margin-top: 6px;
  font-size: 0.8125rem;
  color: #f43f5e;
}

.duration-input__time-symbol {
  position: absolute;
  right: 10px;
  top: 10px;
  color: #94a3b8;
  pointer-events: none;
}
</style>
