import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import ChannelStatsView from "@/views/ChannelStatsView.vue";

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/stats/:channel',
      name: 'channel-stats',
      component: ChannelStatsView,
    },
  ],
})

export default router
