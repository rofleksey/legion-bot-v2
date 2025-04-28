<template>
  <div class="language-switcher">
    <button @click="setLanguage('en')" :class="{ active: locale === 'en' }">EN</button>
    <button @click="setLanguage('ru')" :class="{ active: locale === 'ru' }">RU</button>
  </div>
</template>

<script setup lang="ts">
import {useI18n} from 'vue-i18n'
import {onBeforeMount} from "vue";

const { locale } = useI18n()

function setLanguage(lang: string) {
  locale.value = lang
  localStorage.setItem('legionbot_lang', lang)
}

function detectLanguage(): string {
  const localLang = localStorage.getItem('legionbot_lang')
  if (localLang) {
    return localLang
  }

  // @ts-ignore
  const browserLang = navigator.language || navigator.userLanguage;
  if (browserLang.startsWith('ru')) {
    return 'ru';
  }

  return 'en';
}

onBeforeMount(() => {
  locale.value = detectLanguage()
})
</script>

<style scoped>
.language-switcher {
  display: none;
  gap: 0.5rem;
}

@media (min-width: 768px) {
  .language-switcher {
    display: flex;
  }
}

.language-switcher button {
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border);
  padding: 0.25rem 0.75rem;
  border-radius: var(--radius);
  cursor: pointer;
  transition: all 0.2s ease;
}

.language-switcher button:hover {
  color: var(--text);
  border-color: var(--text-secondary);
}

.language-switcher button.active {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}
</style>
