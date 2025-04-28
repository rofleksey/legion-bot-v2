<script setup lang="ts">
import {RouterView} from 'vue-router'
import AppHeader from "@/components/AppHeader.vue";
import AppFooter from "@/components/AppFooter.vue";
import AppBackground from "@/components/AppBackground.vue";
</script>

<template>
  <AppHeader />
  <AppBackground />
  <RouterView />
  <AppFooter />
</template>

<style scoped>
</style>
