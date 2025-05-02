<template>
  <section class="hero">
    <h1>{{ t('app_title') }}</h1>
    <p>{{ t('app_description') }}</p>
    <div class="cta-buttons">
      <button class="primary-btn" @click="onStart">{{ t('get_started') }}</button>
      <button class="secondary-btn">{{ t('learn_more') }}</button>
    </div>
  </section>

  <section class="features">
    <h2 class="section-title">{{ t('powerful_features')}} </h2>
    <div class="feature-grid">
      <div class="feature-card scroll-animate" :class="{visible: featuresVisible[0]}"
           ref="feature1">
        <div class="feature-icon">🤖</div>
        <h3 class="feature-title">{{ t('feature1_title') }}</h3>
        <p class="feature-desc">{{ t('feature1_desc') }}</p>
      </div>

      <div class="feature-card scroll-animate" :class="{visible: featuresVisible[1]}"
           ref="feature2">
        <div class="feature-icon">🎮</div>
        <h3 class="feature-title">{{ t('feature2_title') }}</h3>
        <p class="feature-desc">{{ t('feature2_desc') }}</p>
      </div>

      <div class="feature-card scroll-animate" :class="{visible: featuresVisible[2]}"
           ref="feature3">
        <div class="feature-icon">✨</div>
        <h3 class="feature-title">{{ t('feature3_title') }}</h3>
        <p class="feature-desc">{{ t('feature3_desc') }}</p>
      </div>

      <div class="feature-card scroll-animate" :class="{visible: featuresVisible[3]}"
           ref="feature4">
        <div class="feature-icon">📊</div>
        <h3 class="feature-title">{{ t('feature4_title') }}</h3>
        <p class="feature-desc">{{ t('feature4_desc') }}</p>
      </div>

      <div class="feature-card scroll-animate" :class="{visible: featuresVisible[4]}"
           ref="feature5">
        <div class="feature-icon">🤝</div>
        <h3 class="feature-title">{{ t('feature5_title') }}</h3>
        <p class="feature-desc">{{ t('feature5_desc') }}</p>
      </div>

      <div class="feature-card scroll-animate" :class="{visible: featuresVisible[5]}"
           ref="feature6">
        <div class="feature-icon">🔔</div>
        <h3 class="feature-title">{{ t('feature6_title') }}</h3>
        <p class="feature-desc">{{ t('feature6_desc') }}</p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from "vue";
import {useRouter} from "vue-router";
import {useUserStore} from "@/stores/user.ts";
import axios from "axios";
import {useI18n} from "vue-i18n";
import {useNotifications} from "@/services/notifications.ts";
import {errorToString, ymReachGoal} from "@/lib/misc.ts";

const {t} = useI18n()
const notifications = useNotifications()
const router = useRouter()

const userStore = useUserStore();
const user = computed(() => userStore.user);

const featuresVisible = ref(Array(6).fill(false));

function handleScroll() {
  const featureElements = document.querySelectorAll('.feature-card');
  featureElements.forEach((el, index) => {
    if (el) {
      const rect = el.getBoundingClientRect();
      featuresVisible.value[index] = rect.top < window.innerHeight - 100;
    }
  });
}

function onStart() {
  if (user.value) {
    router.push('/settings')
  } else {
    loginWithTwitch().catch((e) => {
      notifications.error('Twitch loging error', errorToString(e));
    })
  }
}

async function loginWithTwitch() {
  ymReachGoal('login')
  const response = await axios.get('/api/auth/login');
  if (response.data?.authUrl) {
    localStorage.setItem('twitch_auth_state', response.data.state);
    window.location.href = response.data.authUrl;
  } else {
    notifications.error('No auth URL');
  }
}

onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get('token');
  const state = urlParams.get('state');

  if (token && state) {
    const storedState = localStorage.getItem('twitch_auth_state');
    if (state === storedState) {
      localStorage.removeItem('twitch_auth_state');
      userStore.login(token)
    } else {
      notifications.error('State mismatch');
    }
  }

  window.addEventListener('scroll', handleScroll, {passive: true});
  handleScroll();
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
});
</script>

<style scoped>
.hero {
  padding: 6rem 2rem;
  text-align: center;
  max-width: 800px;
  margin: 0 auto;
  position: relative;
  z-index: 5;
}

.hero h1 {
  font-size: 3rem;
  margin-bottom: 1.5rem;
  line-height: 1.2;
}

.hero p {
  font-size: 1.2rem;
  color: var(--text-secondary);
  margin-bottom: 2rem;
  line-height: 1.6;
}

.cta-buttons {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.primary-btn {
  background-color: var(--primary);
  color: white;
  border: none;
  padding: 0.8rem 1.6rem;
  border-radius: 0.25rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.primary-btn:hover {
  background-color: var(--primary-dark);
  transform: translateY(-1px);
}

.secondary-btn {
  background-color: transparent;
  color: var(--text);
  border: 1px solid var(--text-secondary);
  padding: 0.8rem 1.6rem;
  border-radius: 0.25rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.secondary-btn:hover {
  border-color: var(--text);
  transform: translateY(-1px);
}

.features {
  padding: 6rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.section-title {
  text-align: center;
  font-size: 2rem;
  margin-bottom: 4rem;
  position: relative;
}

.section-title::after {
  content: '';
  position: absolute;
  bottom: -1rem;
  left: 50%;
  transform: translateX(-50%);
  width: 80px;
  height: 4px;
  background-color: var(--primary);
  border-radius: 2px;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
}

.feature-card {
  background-color: var(--card);
  border-radius: 0.5rem;
  padding: 2rem;
  transition: all 0.3s ease;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.feature-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.2);
}

.feature-icon {
  font-size: 2rem;
  margin-bottom: 1rem;
  color: var(--primary);
}

.feature-title {
  font-size: 1.25rem;
  margin-bottom: 1rem;
}

.feature-desc {
  color: var(--text-secondary);
  line-height: 1.6;
}

.scroll-animate {
  opacity: 0;
  transform: translateY(20px);
  transition: all 0.6s ease;
}

.scroll-animate.visible {
  opacity: 1;
  transform: translateY(0);
}

@media (max-width: 768px) {
  .hero h1 {
    font-size: 2.2rem;
  }

  .hero p {
    font-size: 1rem;
  }

  .cta-buttons {
    flex-direction: column;
  }
}
</style>
