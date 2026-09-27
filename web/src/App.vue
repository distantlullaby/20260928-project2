<template>
  <div class="app">
    <header class="topbar">
      <div class="topbar-inner">
        <router-link to="/feed" class="brand" @click="$router.push('/feed')">
          <span class="brand-logo">🍜</span>
          <span>替你尝一口</span>
        </router-link>
        <div class="topnav">
          <router-link to="/feed" class="nav-link" active-class="active">求代吃广场</router-link>
          <template v-if="auth.isLoggedIn">
            <router-link to="/profile" class="nav-link" active-class="active">个人中心</router-link>
            <span class="coin-chip">🪙 {{ auth.user?.coin_balance ?? 0 }}</span>
            <button class="btn btn-ghost btn-sm" @click="auth.logout(); $router.push('/feed')">
              退出
            </button>
          </template>
          <template v-else>
            <button class="btn btn-outline btn-sm" @click="ui.openAuth('login')">登录</button>
            <button class="btn btn-primary btn-sm" @click="ui.openAuth('register')">注册领币</button>
          </template>
        </div>
      </div>
    </header>

    <main class="page">
      <router-view />
    </main>

    <footer class="footer">替你尝一口 · 先看别人替你吃的现场，再决定要不要亲自去</footer>

    <AuthModal v-if="ui.ui.authModalOpen" :mode="ui.ui.authMode" @close="ui.closeAuth()" @ok="onAuthOk" />

    <div v-if="toastState.text" class="toast">{{ toastState.text }}</div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuth } from './store/auth'
import { useToast } from './store/toast'
import { useUI } from './store/ui'
import { getMe } from './api'
import AuthModal from './components/AuthModal.vue'

const auth = useAuth()
const toastState = useToast()
const ui = useUI()
const route = useRoute()

function onAuthOk() {
  ui.closeAuth()
}

onMounted(async () => {
  if (route.query.login) ui.openAuth('login')
  // 已登录时刷新一次用户信息（余额可能在其他标签页变动）
  if (auth.isLoggedIn) {
    try {
      const { user } = await getMe()
      auth.setUser(user)
    } catch {
      /* token 失效时拦截器会清理 */
    }
  }
})
</script>

<style scoped>
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.topbar {
  position: sticky;
  top: 0;
  z-index: 100;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--line);
}
.topbar-inner {
  max-width: 760px;
  margin: 0 auto;
  padding: 12px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 20px;
  font-weight: 800;
  color: var(--brand);
}
.brand-logo {
  font-size: 24px;
}
.topnav {
  display: flex;
  align-items: center;
  gap: 10px;
}
.nav-link {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink-2);
  padding: 5px 8px;
  border-radius: 8px;
}
.nav-link.active {
  color: var(--brand);
  background: var(--brand-soft);
}
.coin-chip {
  background: #fff7e6;
  border: 1px solid #ffe0a3;
  color: #9a6b00;
  font-size: 13px;
  font-weight: 700;
  padding: 4px 11px;
  border-radius: 999px;
}
.page {
  flex: 1;
  width: 100%;
  max-width: 760px;
  margin: 0 auto;
  padding: 20px 16px 48px;
}
.footer {
  text-align: center;
  color: var(--ink-3);
  font-size: 12px;
  padding: 20px;
  border-top: 1px solid var(--line);
}
</style>
