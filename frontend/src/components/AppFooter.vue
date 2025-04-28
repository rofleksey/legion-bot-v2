<template>
  <footer>
    <div class="footer-links">
      <a href="https://t.me/rofleksey" class="footer-link">{{ t('support')}}</a>
      <a href="#" class="footer-link">{{ t('privacy_policy')}}</a>
      <a href="#" class="footer-link">{{ t('terms_of_service')}}</a>
    </div>
    <p class="copyright">© 2024-2025 DBD Legion Bot</p>
  </footer>
</template>

<script setup lang="ts">
import {useI18n} from "vue-i18n";

const {t} = useI18n()
</script>

<style scoped>

footer {
  padding: 3rem 2rem;
  text-align: center;
  background-color: rgba(0, 0, 0, 0.5);
  margin-top: 4rem;
}

.footer-links {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  justify-content: center;
  gap: 2rem;
  margin-bottom: 1.5rem;
}

.footer-link {
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.2s ease;
}

.footer-link:hover {
  color: var(--text);
}
</style>
