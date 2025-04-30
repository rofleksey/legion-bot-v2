import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import ChannelStatsView from "@/views/ChannelStatsView.vue";
import SettingsView from "@/views/SettingsView.vue";
import CheatDetectView from "@/views/CheatDetectView.vue";
import {useUserStore} from "@/stores/user.ts";
import UserListView from "@/views/admin/UserListView.vue";

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/settings',
      name: 'settings',
      component: SettingsView,
      beforeEnter: () => {
        const userStore = useUserStore();
        if (userStore.user) {
          return
        }
        return '/'
      },
    },
    {
      path: '/stats/:channel',
      name: 'channel-stats',
      component: ChannelStatsView,
    },
    {
      path: '/cheat_detector',
      name: 'cheat-detector',
      component: CheatDetectView,
    },
    {
      path: '/admin/userList',
      name: 'admin-user-list',
      component: UserListView
    },
  ],
})

export default router
