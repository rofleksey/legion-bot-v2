<template>
  <header>
    <AppLogo />
    <LanguageSwitcher class="app-header-lang-switcher" />
    <div class="grower"/>
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
  align-items: center;
  position: relative;
  z-index: 10;
}

.app-header-lang-switcher {
  padding-left: 40px;
}

.grower {
  flex-grow: 1;
}
</style>
