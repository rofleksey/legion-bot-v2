document.addEventListener('DOMContentLoaded', () => {
  const { createApp } = Vue;

  createApp({
    data() {
      return {
        stats: {},
        loading: true,
        error: false,
        channel: window.location.pathname.split('/stats/')[1] || '',
        language: this.detectLanguage(),
        translations: {
          en: {
            stats_title: 'Channel Statistics',
            stats_subtitle: 'Statistics for',
            loading: 'Loading statistics...',
            error_loading: 'Failed to load statistics',
            retry: 'Try Again',
          },
          ru: {
            stats_title: 'Статистика канала',
            stats_subtitle: 'Статистика для',
            loading: 'Загрузка статистики...',
            error_loading: 'Ошибка загрузки статистики',
            retry: 'Попробовать снова',
          }
        }
      };
    },
    methods: {
      fetchStats() {
        this.loading = true;
        this.error = false;

        fetch(`/api/stats/${this.channel}`)
          .then(response => {
            if (!response.ok) {
              throw new Error('Network response was not ok');
            }
            return response.json();
          })
          .then(data => {
            this.stats = data;
            this.loading = false;
          })
          .catch(error => {
            console.error('Error fetching stats:', error);
            this.error = true;
            this.loading = false;
          });
      },
      detectLanguage() {
        const browserLang = navigator.language || navigator.userLanguage;
        if (browserLang.startsWith('ru')) {
          return 'ru';
        }

        return 'en';
      },
      translate(key) {
        return this.translations[this.language]?.[key] || this.translations['en'][key] || key;
      },
      setLanguage(lang) {
        this.language = lang;
        // You could also save the language preference to localStorage
        // localStorage.setItem('preferredLanguage', lang);
      }
    },
    mounted() {
      // Check for saved language preference
      // const savedLang = localStorage.getItem('preferredLanguage');
      // if (savedLang) this.language = savedLang;

      this.fetchStats();
    }
  }).mount('#app');
});
