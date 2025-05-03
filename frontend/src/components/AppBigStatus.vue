<template>
  <div class="api-status" :class="`api-status--${status}`">
    <div class="api-status__icon-container">
      <div class="api-status__icon">
        <!-- Idle state (breathing animation) -->
        <div v-if="status === 'idle'" class="api-status__idle-circle"></div>

        <!-- Loading state (wave animation) -->
        <div v-if="status === 'loading'" class="api-status__wave-container">
          <div class="api-status__wave" v-for="i in 3" :key="i" :style="`--delay: ${i * 0.33}s`"></div>
        </div>

        <!-- Success state -->
        <svg v-else-if="status === 'success'" viewBox="0 0 24 24" fill="none">
          <path d="M20 6L9 17L4 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>

        <!-- Error state -->
        <svg v-else-if="status === 'error'" viewBox="0 0 24 24" fill="none">
          <path d="M18 6L6 18M6 6L18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
    </div>

    <div class="api-status__content">
      <h3 class="api-status__title">{{ title }}</h3>
      <p v-if="subtitle" class="api-status__subtitle">{{ subtitle }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
type ApiStatus = 'idle' | 'loading' | 'success' | 'error'

interface Props {
  status?: ApiStatus
  title: string
  subtitle?: string
}

const props = withDefaults(defineProps<Props>(), {
  status: 'idle'
})
</script>

<style scoped>
.api-status {
  --success-color: hsl(142, 71%, 45%);
  --error-color: hsl(0, 84%, 60%);
  --loading-color: hsl(210, 100%, 56%);
  --idle-color: hsl(240, 5%, 60%);
  --surface: hsl(240, 5%, 26%);
  --text-primary: hsl(0, 0%, 98%);
  --text-secondary: hsl(240, 1%, 70%);

  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 2rem;
  border-radius: 1rem;
  background-color: var(--card);
  color: var(--text-primary);
  max-width: 600px;
  margin: 0 auto;
  transition: all 0.3s ease;
}

.api-status__icon-container {
  display: flex;
  align-items: center;
  justify-content: center;
}

.api-status__icon {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  border-radius: 50%;
  background-color: var(--surface);
  color: var(--text-primary);
  overflow: hidden;
}

/* Idle State */
.api-status--idle .api-status__icon {
  background-color: hsla(240, 5%, 60%, 0.1);
}

.api-status__idle-circle {
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  background-color: var(--idle-color);
  opacity: 0.8;
  animation: api-status-breathe 3s ease-in-out infinite;
}

/* Loading State */
.api-status--loading .api-status__icon {
  background-color: hsla(210, 100%, 56%, 0.1);
  color: var(--loading-color);
}

.api-status__wave-container {
  position: relative;
  width: 100%;
  height: 100%;
}

.api-status__wave {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  margin: auto;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  background-color: var(--loading-color);
  opacity: 0;
  animation: api-status-wave 2s linear infinite;
  animation-delay: var(--delay);
}

/* Success State */
.api-status--success .api-status__icon {
  background-color: hsla(142, 71%, 45%, 0.1);
  color: var(--success-color);
}

.api-status--success .api-status__icon svg {
  width: 2rem;
  height: 2rem;
  animation: api-status-pulse 2s ease-in-out infinite;
}

/* Error State */
.api-status--error .api-status__icon {
  background-color: hsla(0, 84%, 60%, 0.1);
  color: var(--error-color);
}

.api-status--error .api-status__icon svg {
  width: 2rem;
  height: 2rem;
  animation: api-status-shake 0.8s ease infinite;
}

/* Content */
.api-status__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.api-status__title {
  font-size: 1.5rem;
  font-weight: 600;
  line-height: 1.3;
  color: var(--text-primary);
  margin: 0;
}

.api-status__subtitle {
  font-size: 1.1rem;
  line-height: 1.5;
  color: var(--text-secondary);
  margin: 0;
  opacity: 0.9;
}

/* Animations */
@keyframes api-status-breathe {
  0%, 100% {
    transform: scale(1);
    opacity: 0.8;
  }
  50% {
    transform: scale(1.2);
    opacity: 0.4;
  }
}

@keyframes api-status-wave {
  0% {
    transform: scale(0.5);
    opacity: 0.8;
  }
  100% {
    transform: scale(3);
    opacity: 0;
  }
}

@keyframes api-status-pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.2);
  }
}

@keyframes api-status-shake {
  0%, 100% {
    transform: translateX(0) rotate(0deg);
  }
  20%, 60% {
    transform: translateX(-4px) rotate(-2deg);
  }
  40%, 80% {
    transform: translateX(4px) rotate(2deg);
  }
}
</style>
