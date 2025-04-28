import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'

import App from './App.vue'
import router from './router'

import Notifications from './services/notifications'
import NotificationContainer from './components/AppNotificationContainer.vue'

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages: {
    en: {
      twitch_login: 'Login with Twitch',
      app_title: 'Bring The Entity\'s Terror To Your Chat',
      app_description: 'Legion Bot stalks your chat like a Dead by Daylight killer, striking unpredictably and timing out chatters on the fifth hit. Customize your killer, attack frequency, and AI responses for maximum suspense.',
      get_started: 'Get started',
      learn_more: 'Learn more',
      powerful_features: 'Powerful features',
      feature1_title: 'Killer AI Responses',
      feature1_desc: 'Our AI generates creepy killer-themed responses when users mention the bot, enhancing the Dead by Daylight atmosphere.',
      feature2_title: 'Multiple Killers',
      feature2_desc: 'Choose from different Dead by Daylight killers like Legion, Myers, and Ghost Face, each with unique stalking behaviors.',
      feature3_title: 'Interactive Stalking',
      feature3_desc: 'The bot appears periodically to "stab" chatters, with the 5th hit triggering a customizable timeout.',
      feature4_title: 'Full Customization',
      feature4_desc: 'Adjust attack frequency, timeout duration, killer selection, and other parameters to fit your stream\'s vibe.',
      feature5_title: 'Chat Engagement',
      feature5_desc: 'Keeps chat active and entertained with unpredictable killer appearances and reactions between matches.',
      feature6_title: 'DbD Features',
      feature6_desc: 'Special Dead by Daylight themed commands, stats tracking, and more killer features coming soon.',
      support: 'Support',
      privacy_policy: 'Privacy Policy',
      terms_of_service: 'Terms of Service',
      stats_title: 'Channel Statistics',
      stats_subtitle: 'Statistics for',
      loading: 'Loading statistics...',
      error_loading: 'Failed to load statistics',
      retry: 'Try Again',
      stats: {
        bleedOuts: 'Bleed Outs',
        bodyBlock: 'Body Blocks',
        fail: 'Failure count',
        success: 'Success count',
        total: 'Total count',
        hits: 'Hits',
        miss: 'Misses',
        stuns: 'Stuns',
      }
    },
    ru: {
      twitch_login: 'Войти через Twitch',
      app_title: 'Вселите Ужас Сущности в Ваш Чат',
      app_description: 'Легион Бот охотится за вашими зрителями как убийца из Dead by Daylight — нападает неожиданно, а после пятого удара отправляет нарушителей в таймаут. Настраивайте персонажа, частоту атак и ответы ИИ для максимального погружения.',
      get_started: 'Начать',
      learn_more: 'Подробнее',
      powerful_features: 'Возможности бота',
      feature1_title: 'Ответы ИИ',
      feature1_desc: 'ИИ генерирует пугающие ответы в стиле убийц, когда зрители упоминают бота, усиливая атмосферу Dead by Daylight.',
      feature2_title: 'Разные убийцы',
      feature2_desc: 'Несколько персонажей из Dead by Daylight: Легион, Тень, Крик — у каждого уникальная манера преследования.',
      feature3_title: 'Интерактивная охота',
      feature3_desc: 'Бот периодически "атакует" зрителей, а после 5-го удара активирует настраиваемый таймаут.',
      feature4_title: 'Гибкие настройки',
      feature4_desc: 'Регулируйте частоту атак, длительность таймаута, выбор убийцы и другие параметры под стиль вашего стрима.',
      feature5_title: 'Вовлечение чата',
      feature5_desc: 'Поддерживает активность чата неожиданными появлениями убийцы и реакциями между матчами.',
      feature6_title: 'Особенности DBD',
      feature6_desc: 'Уникальные тематические команды, статистика и новые функции убийц в разработке.',
      support: 'Поддержка',
      privacy_policy: 'Политика конфиденциальности',
      terms_of_service: 'Условия использования',
      stats_title: 'Статистика канала',
      stats_subtitle: 'Статистика для',
      loading: 'Загрузка статистики...',
      error_loading: 'Ошибка загрузки статистики',
      retry: 'Попробовать снова',
      stats: {
        bleedOuts: 'Истеканий',
        bodyBlock: 'Боди Блоков',
        fail: 'Провалено',
        success: 'Успехов',
        total: 'Всего',
        hits: 'Ударов',
        miss: 'Промахов',
        stuns: 'Оглушений',
      }
    }
  }
})

const app = createApp(App)

app.component('NotificationContainer', NotificationContainer)

app.use(i18n)
app.use(createPinia())
app.use(router)
app.use(Notifications)

app.mount('#app')
