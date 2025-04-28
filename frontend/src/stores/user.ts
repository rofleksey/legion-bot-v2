import {ref} from 'vue'
import {defineStore} from 'pinia'
import type {User} from "@/lib/types.ts";
import axios from 'axios'

const localStorageTokenKey = 'legionbot-token'
const localStorageUserKey = 'legionbot-user'

export const useUserStore = defineStore('user', () => {
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
    }).catch((e) => {
      console.error(e)
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
  }

  if (token.value) {
    validateToken(token.value).catch((e) => {
      console.error(e)
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
