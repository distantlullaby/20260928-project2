import { createRouter, createWebHashHistory } from 'vue-router'
import { auth } from './store'

const routes = [
  { path: '/', name: 'feed', component: () => import('./pages/FeedPage.vue') },
  { path: '/login', name: 'login', component: () => import('./pages/LoginPage.vue') },
  { path: '/profile', name: 'profile', component: () => import('./pages/ProfilePage.vue'), meta: { requiresAuth: true } }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  }
})

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'login' && auth.isLoggedIn) {
    return { name: 'feed' }
  }
  return true
})

export default router
