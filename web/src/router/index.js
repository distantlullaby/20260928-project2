import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuth } from '../store/auth'
import FeedView from '../views/FeedView.vue'
import ProfileView from '../views/ProfileView.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/feed' },
    { path: '/feed', component: FeedView },
    { path: '/profile', component: ProfileView, meta: { requiresAuth: true } }
  ]
})

router.beforeEach((to) => {
  const auth = useAuth()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return { path: '/feed', query: { login: '1' } }
  }
})

export default router
