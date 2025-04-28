<template>
  <button class="login-btn" @click="loginWithTwitch" v-if="!user">
    {{ t('twitch_login') }}
  </button>
  <div v-else class="user-info">
    <img :src="user.profileImageUrl" class="user-avatar" width="32" height="32" style="border-radius: 50%;">
    <span>{{ user.displayName }}</span>
  </div>
</template>

<script setup lang="ts">
import {useI18n} from 'vue-i18n'
import {useUserStore} from "@/stores/user";
import {computed} from "vue";
import axios from "axios";

const { t } = useI18n()

const userStore = useUserStore()
const user = computed(() => userStore.user)

async function loginWithTwitch() {
  try {
    const response = await axios.get('/api/auth/login');
    if (response.data?.authUrl) {
      localStorage.setItem('twitch_auth_state', response.data.state);
      window.location.href = response.data.authUrl;
    }
  } catch (error) {
    console.error('Error during Twitch login:', error);
    // Consider adding user feedback here
  }
}
</script>

<style scoped>
.login-btn {
  background-color: var(--primary);
  color: white;
  border: none;
  padding: 0.6rem 1.2rem;
  border-radius: 0.25rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.login-btn:hover {
  background-color: var(--primary-dark);
  transform: translateY(-1px);
}

.user-info {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 5px;
}

@media (max-width: 768px) {
  .user-info span {
    display: none;
  }
}
</style>
