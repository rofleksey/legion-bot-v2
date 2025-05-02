<template>
  <button class="login-btn" @click="loginWithTwitch" v-if="!user">
    {{ t('login') }}
  </button>
  <div v-else class="user-info" @click="router.push('/settings')">
    <img :src="user.profileImageUrl" class="user-avatar" width="32" height="32" style="border-radius: 50%;">
    <span>{{ user.displayName }}</span>
  </div>
</template>

<script setup lang="ts">
import {useI18n} from 'vue-i18n'
import {useUserStore} from "@/stores/user";
import {computed} from "vue";
import axios from "axios";
import {useRouter} from "vue-router";
import {useNotifications} from "@/services/notifications.ts";
import {errorToString, ymReachGoal} from "@/lib/misc.ts";

const notifications = useNotifications();
const { t } = useI18n()
const router = useRouter()

const userStore = useUserStore()
const user = computed(() => userStore.user)

async function loginWithTwitch() {
  try {
    ymReachGoal('login')
    const response = await axios.get('/api/auth/login');
    if (response.data?.authUrl) {
      localStorage.setItem('twitch_auth_state', response.data.state);
      window.location.href = response.data.authUrl;
    }
  } catch (e) {
    notifications.error('Error during Twitch login', errorToString(e));
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
  cursor: pointer;
}

@media (max-width: 768px) {
  .user-info span {
    display: none;
  }
}
</style>
