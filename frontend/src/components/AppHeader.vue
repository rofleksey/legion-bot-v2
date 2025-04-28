<template>
  <header>
    <AppLogo />
    <LanguageSwitcher />
    <LoginButton />
  </header>
</template>

<script setup lang="ts">
import LoginButton from "@/components/LoginButton.vue";
import AppLogo from "@/components/AppLogo.vue";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
</script>

<style scoped>
header {
  padding: 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  z-index: 10;
}
</style>
