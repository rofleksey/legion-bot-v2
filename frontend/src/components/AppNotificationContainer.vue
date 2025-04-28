<template>
  <div class="notification-container">
    <AppNotification
      v-for="notification in notifications"
      :key="notification.id"
      v-bind="notification"
      @close="removeNotification(notification.id)"
    />
  </div>
</template>

<script lang="ts" setup>
import AppNotification from "@/components/AppNotification.vue";
import {useNotificationStore} from '@/services/notifications'

const { notifications, removeNotification } = useNotificationStore()
</script>

<style>
.notification-container {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 9999;
}
</style>
