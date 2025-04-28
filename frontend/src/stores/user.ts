import {ref} from 'vue'
import {defineStore} from 'pinia'
import type {User} from "@/lib/types.ts";
import axios from 'axios'
import {useNotifications} from "@/services/notifications.ts";

const localStorageTokenKey = 'legionbot-token'
const localStorageUserKey = 'legionbot-user'

export const useUserStore = defineStore('user', () => {
  const notifications = useNotifications()

  const localTokenStr = localStorage.getItem(localStorageTokenKey)
  const token = ref<string | null>(localTokenStr || null);

  const localUserStr = localStorage.getItem(localStorageUserKey)
  const user = ref<User | null>(localUserStr ? JSON.parse(localUserStr) : null)

  function login(newToken: string) {
    validateToken(newToken).then((newUser) => {
      token.value = newToken
      user.value = newUser

      localStorage.setItem(localStorageTokenKey, newToken)
      localStorage.setItem(localStorageUserKey, JSON.stringify(newUser))
    }).catch(() => {
      notifications.error('Failed to validate token', 'Login again to continue')
      logout()
    }).finally(() => {
      window.history.replaceState({}, '', window.location.pathname);
    })
  }

  async function validateToken(token: string): Promise<User> {
    const response = await axios.get<User>('/api/validate', {
      headers: { 'Authorization': `Bearer ${token}` }
    });

    if (!response.data) {
      throw new Error("invalid user data")
    }

    return response.data;
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem(localStorageUserKey)
    localStorage.removeItem(localStorageTokenKey)
    notifications.warning('Logged out')
  }

  if (token.value) {
    validateToken(token.value).then((user) => {
      console.log(user)
    }).catch(() => {
      notifications.error('Failed to validate token', 'Login again to continue')
      logout()
    })
  }

  return {
    user,
    token,
    login,
    logout
  }
})
