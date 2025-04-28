<template>
  <section class="container">
    <div class="page-header">
      <h1>{{ t('stats_title') }}</h1>
      <p>{{ t('stats_subtitle') }} {{ channel }}</p>
    </div>

    <div class="stats-grid">
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <p>{{ t('loading') }}</p>
      </div>

      <div v-else-if="error" class="error">
        <p>{{ t('error_loading') }}</p>
        <button @click="fetchStats" class="retry-btn">{{ t('retry') }}</button>
      </div>

      <template v-else>
        <div v-for="(value, key) in stats" :key="key" class="stat-card">
          <div class="stat-value">{{ value }}</div>
          <div class="stat-label">{{ t(`stats.${key}`) }}</div>
        </div>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import {computed, onMounted, ref} from "vue";
import {useRoute} from "vue-router";
import {useI18n} from "vue-i18n";
import axios from "axios";
import {useNotifications} from "@/services/notifications.ts";

const notifications = useNotifications();
const {t} = useI18n()

const route = useRoute();
const channel = computed(() => route.params.channel)

const loading = ref(false)
const error = ref(false)
const stats = ref<{ [key: string]: number }>({})

function fetchStats() {
  loading.value = true;
  error.value = false;

  axios.get(`/api/stats/${channel.value}`)
    .then(response => {
      stats.value = response.data;
      loading.value = false;
    })
    .catch((e) => {
      notifications.error('Error fetching stats', e?.toString?.() ?? '');
      error.value = true;
    }).finally(() => {
      loading.value = false;
    })
}

onMounted(() => {
  fetchStats();
});
</script>

<style scoped>

.page-header {
  margin-bottom: 2rem;
  text-align: center;
}

.page-header h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.page-header p {
  color: var(--text-secondary);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-top: 2rem;
  padding: 2rem;
}

.stat-card {
  background: var(--card);
  border-radius: var(--radius);
  padding: 1.5rem;
  border: 1px solid var(--border);
  transition: transform 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-5px);
}

.stat-value {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--primary);
  margin-bottom: 0.5rem;
}

.stat-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
  text-transform: capitalize;
}

.loading {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  gap: 1rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--primary);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error {
  grid-column: 1 / -1;
  text-align: center;
  padding: 2rem;
  color: #ff6b6b;
}

.retry-btn {
  background: var(--primary);
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: var(--radius);
  margin-top: 1rem;
  cursor: pointer;
  transition: background 0.2s ease;
}

.retry-btn:hover {
  background: var(--primary-dark);
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  }

  .page-header h1 {
    font-size: 1.5rem;
  }
}
</style>
