<template>
  <div class="user-management-container">
    <div class="user-management-header">
      <h1 class="user-management-title">
        User list
      </h1>
      <p class="user-management-subtitle">
        Login as another user
      </p>
    </div>

    <div class="user-management-search">
      <div class="search-input-container">
        <input
          v-model="searchQuery"
          class="search-input"
          placeholder="Search query"
          type="text"
        />
      </div>
    </div>

    <div class="users-container">
      <div v-if="isLoading" class="loading-state">
        <div class="spinner"></div>
        <p>Loading...</p>
      </div>

      <div v-else-if="filteredUsers.length === 0" class="empty-state">
        <svg class="empty-icon" viewBox="0 0 24 24">
          <path fill="currentColor" d="M12,2A10,10 0 0,1 22,12A10,10 0 0,1 12,22A10,10 0 0,1 2,12A10,10 0 0,1 12,2M12,4A8,8 0 0,0 4,12A8,8 0 0,0 12,20A8,8 0 0,0 20,12A8,8 0 0,0 12,4M11,7H13V9H11V7M11,11H13V17H11V11Z" />
        </svg>
        <h3>No users found</h3>
        <p>Try again</p>
      </div>

      <div v-else class="users-list">
        <div
          v-for="user in filteredUsers"
          :key="user.login"
          class="user-item"
          @click="loginAsUser(user)"
        >
          <div class="user-login">{{ user.login }}</div>
          <div v-if="isLoggingIn && currentLogin === user.login" class="user-login-spinner">
            <span class="spinner"></span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useNotifications } from "@/services/notifications.ts";
import axios from "axios";
import { useI18n } from "vue-i18n";
import { errorToString } from "@/lib/misc.ts";
import { useUserStore } from "@/stores/user";

interface TwitchUser {
  login: string;
}

interface LoginResponse {
  token: string;
  user: {
    login: string;
    displayName: string;
    profileImageUrl: string;
  };
}

const notifications = useNotifications()

const userStore = useUserStore()
const token = computed(() => userStore.token)

const users = ref<TwitchUser[]>([])
const isLoading = ref(false)
const isLoggingIn = ref(false)
const currentLogin = ref('')
const searchQuery = ref('')

const filteredUsers = computed(() => {
  if (!searchQuery.value) {
    return users.value
  }

  const query = searchQuery.value.toLowerCase().trim()
  return users.value.filter((user) => user.login.toLowerCase().includes(query))
})

onMounted(async () => {
  isLoading.value = true

  try {
    const response = await axios.get('/api/admin/users', {
      headers: {Authorization: `Bearer ${token.value}`}
    })
    users.value = response.data
  } catch (e) {
    notifications.error('Failed to load users', errorToString(e))
    users.value = []
  } finally {
    isLoading.value = false
  }
})

const loginAsUser = async (user: TwitchUser) => {
  isLoggingIn.value = true
  currentLogin.value = user.login

  try {
    const response = await axios.post('/api/admin/loginAs', {
      login: user.login
    }, {
      headers: {Authorization: `Bearer ${token.value}`}
    })

    const data: LoginResponse = response.data
    userStore.login(data.token)

    notifications.info(`Logged in as ${user.login}`)
  } catch (e) {
    notifications.error(`Failed to login as ${user.login}`, errorToString(e))
  } finally {
    isLoggingIn.value = false
    currentLogin.value = ''
  }
}
</script>

<style scoped>
.user-management-container {
  max-width: 42rem;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.user-management-header {
  margin-bottom: 1.5rem;
  text-align: center;
}

.user-management-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.user-management-subtitle {
  color: var(--text-secondary);
  font-size: 1rem;
}

.user-management-search {
  margin-bottom: 1rem;
}

.search-input-container {
  display: flex;
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

.users-container {
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

.users-list {
  display: flex;
  flex-direction: column;
  max-height: 500px;
  overflow-y: auto;
}

.user-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background-color 0.2s;
}

.user-item:hover {
  background-color: var(--background-hover);
}

.user-item:last-child {
  border-bottom: none;
}

.user-login {
  font-weight: 500;
  color: var(--text-primary);
}

.user-login-spinner {
  width: 1.25rem;
  height: 1.25rem;
}

.spinner {
  width: 1.25rem;
  height: 1.25rem;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: var(--primary);
  animation: spin 1s ease-in-out infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
