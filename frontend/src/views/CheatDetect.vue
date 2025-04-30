<template>
  <div class="cheat-detection-container">
    <div class="cheat-detection-header">
      <h1 class="cheat-detection-title">
        {{t('cheat_detection.title')}}
      </h1>
      <p class="cheat-detection-subtitle">
        {{t('cheat_detection.subtitle')}}
      </p>
    </div>

    <div class="cheat-detection-search">
      <div class="search-input-container">
        <input
          v-model="query"
          @keyup.enter="searchUsers"
          class="search-input"
          :placeholder="t('cheat_detection.enter_username')"
          type="text"
        />
        <button @click="searchUsers" class="search-button" :disabled="isLoading">
          <span v-if="!isLoading">{{t('cheat_detection.search')}}</span>
          <span v-else class="spinner"></span>
        </button>
      </div>
    </div>

    <div v-if="hasSearched" class="results-container">
      <div v-if="isLoading" class="loading-state">
        <div class="spinner"></div>
        <p>{{t('cheat_detection.searching')}}</p>
      </div>

      <div v-else-if="results.length === 0" class="empty-state">
        <svg class="empty-icon" viewBox="0 0 24 24">
          <path fill="currentColor" d="M12,2A10,10 0 0,1 22,12A10,10 0 0,1 12,22A10,10 0 0,1 2,12A10,10 0 0,1 12,2M12,4A8,8 0 0,0 4,12A8,8 0 0,0 12,20A8,8 0 0,0 20,12A8,8 0 0,0 12,4M11,7H13V9H11V7M11,11H13V17H11V11Z" />
        </svg>
        <h3>{{t('cheat_detection.no_results')}}</h3>
        <p>{{t('cheat_detection.username_clean')}}</p>
      </div>

      <div v-else class="results-list">
        <div v-for="(user, index) in results" :key="index" class="result-item">
          <div class="result-username">{{ user.username }}</div>
          <div class="result-site">{{ user.site }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {useNotifications} from "@/services/notifications.ts";
import axios from "axios";
import {useI18n} from "vue-i18n";
import {errorToString} from "@/lib/misc.ts";

interface DetectedUser {
  username: string
  site: string
}

const notifications = useNotifications()
const {t} = useI18n()

const query = ref('')
const isLoading = ref(false)
const results = ref<DetectedUser[]>([])
const hasSearched = ref(false)

const searchUsers = async () => {
  if (!query.value.trim()) return

  isLoading.value = true
  hasSearched.value = true

  try {
    const response = await axios.post('/api/cheatDetect', { username: query.value })
    results.value = response.data
  } catch (e) {
    notifications.error('Failed to get a list of users', errorToString(e));
    results.value = []
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.cheat-detection-container {
  max-width: 42rem;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.cheat-detection-header {
  margin-bottom: 1.5rem;
  text-align: center;
}

.cheat-detection-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.cheat-detection-subtitle {
  color: var(--text-secondary);
  font-size: 1rem;
}

.cheat-detection-search {
  margin-bottom: 2rem;
}

.search-input-container {
  display: flex;
  gap: 0.5rem;
}

.search-input {
  flex: 1;
  padding: 0.75rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: 0.375rem;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--primary);
}

.search-button {
  padding: 0 1.5rem;
  background-color: var(--primary);
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.search-button:hover {
  background-color: var(--primary-dark);
}

.search-button:disabled {
  background-color: var(--primary-disabled);
  cursor: not-allowed;
}

.results-container {
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  overflow: hidden;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 2rem;
  color: var(--text-secondary);
}

.empty-state {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary);
}

.empty-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto 1rem;
  color: var(--text-tertiary);
}

.empty-state h3 {
  font-size: 1.125rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.results-list {
  display: flex;
  flex-direction: column;
}

.result-item {
  display: flex;
  justify-content: space-between;
  padding: 1rem;
  border-bottom: 1px solid var(--border-color);
}

.result-item:last-child {
  border-bottom: none;
}

.result-username {
  font-weight: 500;
  color: var(--text-primary);
}

.result-site {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.spinner {
  width: 1.25rem;
  height: 1.25rem;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s ease-in-out infinite;
  margin: 0 auto;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
