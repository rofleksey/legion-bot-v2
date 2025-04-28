import {type App, inject, ref} from 'vue'
import type {LNotification, LNotificationWithID} from "@/lib/ui-types.ts";

interface NotificationInstance {
  show: (options: LNotification) => void
  info: (title: string, subtitle?: string, duration?: number) => void
  warning: (title: string, subtitle?: string, duration?: number) => void
  error: (title: string, subtitle?: string, duration?: number) => void
  hint: (title: string, subtitle?: string, duration?: number) => void
}

const NOTIFICATION_SYMBOL = Symbol('notifications')
const notificationsStore = ref<LNotificationWithID[]>([])

export function useNotifications(): NotificationInstance {
  const notifications = inject<NotificationInstance>(NOTIFICATION_SYMBOL)
  if (!notifications) {
    throw new Error('Notifications plugin not installed')
  }
  return notifications
}

export function useNotificationStore() {
  return {
    notifications: notificationsStore,
    addNotification: (notification: LNotification) => {
      const id = Date.now().toString()
      notificationsStore.value.push({ id, ...notification })
    },
    removeNotification: (id: string) => {
      notificationsStore.value = notificationsStore.value.filter(
        (notification) => notification.id !== id
      )
    }
  }
}

export default {
  install: (app: App) => {
    const { addNotification } = useNotificationStore()

    const notifications: NotificationInstance = {
      show(options) {
        addNotification(options)
      },
      info(title, subtitle, duration = 3500) {
        this.show({ title, subtitle, type: 'info', duration })
      },
      warning(title, subtitle, duration = 3500) {
        this.show({ title, subtitle, type: 'warning', duration })
      },
      error(title, subtitle, duration = 3500) {
        this.show({ title, subtitle, type: 'error', duration })
      },
      hint(title, subtitle, duration = 3500) {
        this.show({ title, subtitle, type: 'hint', duration })
      },
    }

    app.provide(NOTIFICATION_SYMBOL, notifications)
    app.config.globalProperties.$notifications = notifications
  },
}
