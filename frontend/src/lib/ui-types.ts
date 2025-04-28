export type NotificationType = 'info' | 'warning' | 'error' | 'hint'

export interface LNotification {
  title: string;
  subtitle?: string;
  type?: NotificationType;
  duration?: number;
}

export interface LNotificationWithID extends LNotification {
  id: string
}
