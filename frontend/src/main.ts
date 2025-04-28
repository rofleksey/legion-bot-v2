import './assets/main.css'

import {createApp} from 'vue'
import {createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'

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
      },
      "settings": {
        "title": "Settings",
        "weight": "Weight",
        "save_button": "Save Changes",
        "saving": "Saving...",
        "save_success": "Settings applied successfully!",
        "save_failed": "Failed to apply settings",
        "general_title": "General Settings",
        "killers_title": "Killers Settings",
        "general": "⚙️ General",
        "legion": "🔪 Legion",
        "ghostface": "👻 Ghost Face",
        "doctor": "🧠 Doctor",
        "doctor_description": "While the doctor is in the chat, he has a chance to replace a user's message with a jumbled set of letters (he deletes the message and writes it himself, the author of the message will be lost)",
        "enabled": "Enabled",
        "disabled": "Disabled",
        "language": "Chat Language",
        "delay_between_killers": "Delay Between Killers",
        "delay_at_the_stream_start": "Delay At The Stream Start",
        "min_number_of_viewers": "Min Number Of Viewers",
        "legion_description": "Has a chance to 'hit' users that send messages. Affected users are inflicted with 'deep wound' status effect and need to !mend, otherwise they 'bleed out' and receive a timeout. If it manages to hit 'Fatal Hit' number of users - the last one is 'hooked' and receives a timeout. If it gets no hits for 'Frenzy Timeout' duration - the killer goes away. Can be body blocked (by 'deep wound'-ed users), !pallet stunned, !locker stunned, !tbag-ged.",
        "ghostface_description": "Permanently marks users who send messages. If during any of the next rounds the Ghost Face sees a message from a marked user, he 'kills' themm, 'hooks' them and leaves (the user gets a timeout and loses the mark). Gamers who were not marked during the current round also lose their mark. Comes for a fixed amount of time, silently marks users and leaves. Always reports the number of users who were marked in this round before leaving. Can be !reveal-ed and !tbag-ed",
        "fatal_hit": "Fatal Hit",
        "frenzy_timeout": "Frenzy Timeout",
        "deep_wound_timeout": "Deep Wound Timeout",
        "react_chance": "Message React Chance",
        "hit_chance": "Hit Chance",
        "min_delay_hits": "Min Delay Between Reactions",
        "hook_ban_time": "Hook Ban Time",
        "bleedout_ban_time": "Bleed Out Ban Time",
        "bodyblock_chance": "Body Block Success Chance",
        "locker_grab_chance": "Locker Grab Chance",
        "locker_stun_chance": "Locker Stun Chance",
        "pallet_stun_chance": "Pallet Stun Chance"
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
      },
      "settings": {
        "title": "Настройки",
        "weight": "Вес",
        "save_button": "Сохранить изменения",
        "saving": "Сохранение...",
        "save_success": "Настройки успешно применены!",
        "save_failed": "Не удалось применить настройки",
        "general_title": "Основные настройки",
        "killers_title": "Настройки убийц",
        "general": "⚙️ Общие",
        "legion": "🔪 Легион",
        "ghostface": "👻 Крик",
        "doctor": "🧠 Доктор",
        "enabled": "Включено",
        "disabled": "Отключено",
        "language": "Язык Чата",
        "delay_between_killers": "Задержка между убийцами",
        "delay_at_the_stream_start": "Задержка в начале стрима",
        "min_number_of_viewers": "Мин. Кол-во Зрителей",
        "legion_description": "Имеет шанс 'ударить' пользователей, которые отправляют сообщения. Пораженные пользователи получают эффект 'глубокая рана' и должны использовать команду !mend, иначе они 'истекают кровью' и получают таймаут. Если убийца достигает нужного количества 'ударов' - последний пользователь 'вешается на крюк' и получает таймаут. Если убийца не может нанести ни одного удара в течение 'Времени ярости' - он уходит. Легиона можно бодиблочить (пользователями с 'глубокой раной'), оглушить палетой (!pallet), шкафом (!locker) или тибегнуть ему (!tbag).",
        "ghostface_description": "Помечает пользователей, которые отправляют сообщения. Если в следующем раунде Крик увидит сообщение от помеченного пользователя, он 'убивает' его, 'вешает на крюк' и уходит (пользователь получает таймаут и теряет марку). Пользователи, которые не были помечены за текущий раунд также теряют метку. Приходит на фиксированное время, молча помечает пользователей и уходит. Перед уходом всегда сообщает количество пользователей, которые были помечены в этом раунде. Крика можно обнаружить (!reveal) и тибегнуть (!tbag).",
        "doctor_description": "Пока доктор в чате, имеет шанс заменить сообщение пользователя на перемешанный набор букв (удаляет сообщение и пишет его сам, автор сообщения будет утерян)",
        "fatal_hit": "Номер смертельного удара",
        "frenzy_timeout": "Таймаут ярости",
        "deep_wound_timeout": "Таймаут глубокой раны",
        "react_chance": "Шанс реакции на сообщение",
        "hit_chance": "Шанс попадания",
        "min_delay_between_hits": "Мин. задержка между реакциями",
        "hook_ban_time": "Время бана на крюке",
        "bleedout_ban_time": "Время бана при истекании кровью",
        "bodyblock_chance": "Шанс бодиблока",
        "locker_grab_chance": "Шанс хватания из шкафа",
        "locker_stun_chance": "Шанс оглушения шкафом",
        "pallet_stun_chance": "Шанс оглушения палетой",
        "timeout": "Таймаут"
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
