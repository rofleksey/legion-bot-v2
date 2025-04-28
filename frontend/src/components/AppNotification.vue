<!-- src/components/ui/Notification.vue -->
<template>
  <Transition name="notification">
    <div
      v-if="isVisible"
      class="notification"
      :class="`notification--${type}`"
      role="alert"
    >
      <div class="notification__icon">
        <component :is="iconComponent" />
      </div>
      <div class="notification__content">
        <h3 class="notification__title">{{ title }}</h3>
        <p class="notification__subtitle">{{ subtitle }}</p>
      </div>
      <button class="notification__close" @click="close">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>
  </Transition>
</template>

<script lang="ts" setup>
import { computed, ref, onMounted } from 'vue'
import {
  InfoIcon,
  AlertTriangleIcon,
  XCircleIcon,
  LightbulbIcon,
} from 'lucide-vue-next'

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  subtitle: {
    type: String,
    default: '',
  },
  type: {
    type: String,
    default: 'info',
    validator: (value: string) =>
      ['info', 'warning', 'error', 'hint'].includes(value),
  },
  duration: {
    type: Number,
    default: 3500,
  },
})

const emit = defineEmits(['close'])

const isVisible = ref(false)

const iconComponent = computed(() => {
  switch (props.type) {
    case 'warning':
      return AlertTriangleIcon
    case 'error':
      return XCircleIcon
    case 'hint':
      return LightbulbIcon
    default:
      return InfoIcon
  }
})

onMounted(() => {
  isVisible.value = true
  if (props.duration > 0) {
    setTimeout(() => close(), props.duration)
  }
})

function close() {
  isVisible.value = false
  setTimeout(() => emit('close'), 300)
}
</script>

<style scoped>
.notification {
  position: relative;
  display: flex;
  width: 350px;
  padding: 1rem;
  margin-bottom: 1rem;
  border-radius: 0.5rem;
  box-shadow:
    0 10px 15px -3px rgb(0 0 0 / 0.1),
    0 4px 6px -4px rgb(0 0 0 / 0.1);
  background-color: white;
  border-left: 4px solid;
  gap: 0.75rem;
  align-items: flex-start;
}

.notification--info {
  border-left-color: #3b82f6;
}

.notification--warning {
  border-left-color: #f59e0b;
}

.notification--error {
  border-left-color: #ef4444;
}

.notification--hint {
  border-left-color: #10b981;
}

.notification__icon {
  padding-top: 0.125rem;
  flex-shrink: 0;
}

.notification__icon svg {
  width: 1.25rem;
  height: 1.25rem;
}

.notification--info .notification__icon svg {
  color: #3b82f6;
}

.notification--warning .notification__icon svg {
  color: #f59e0b;
}

.notification--error .notification__icon svg {
  color: #ef4444;
}

.notification--hint .notification__icon svg {
  color: #10b981;
}

.notification__content {
  flex-grow: 1;
}

.notification__title {
  font-weight: 600;
  font-size: 0.875rem;
  line-height: 1.25rem;
  margin-bottom: 0.25rem;
  color: #111827;
}

.notification__subtitle {
  font-size: 0.875rem;
  line-height: 1.25rem;
  color: #6b7280;
}

.notification__close {
  padding: 0.25rem;
  color: #9ca3af;
  background: none;
  border: none;
  cursor: pointer;
  border-radius: 0.25rem;
  flex-shrink: 0;
}

.notification__close:hover {
  color: #6b7280;
  background-color: #f3f4f6;
}

.notification__close svg {
  width: 1rem;
  height: 1rem;
}

.notification-enter-active,
.notification-leave-active {
  transition:
    opacity 0.3s ease,
    transform 0.3s ease;
}

.notification-enter-from,
.notification-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.notification-enter-to,
.notification-leave-from {
  opacity: 1;
  transform: translateX(0);
}
</style>
