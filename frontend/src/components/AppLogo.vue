<template>
  <div class="logo" @click="router.push('/')">
    DBD <span>Legion</span>Bot
  </div>
</template>

<script setup lang="ts">
import {useRouter} from "vue-router";

const router = useRouter();
</script>

<style scoped>
.logo {
  font-size: 1.8rem;
  font-weight: 700;
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 0.5rem;
  user-select: none;
  cursor: pointer;
}

.logo span {
  color: var(--primary);
}
</style>
