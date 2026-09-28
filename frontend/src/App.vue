<template>
  <div class="app-shell">
    <header class="topbar">
      <RouterLink to="/" class="brand">
        <span class="brand-logo">🍜</span>
        <span class="brand-text">
          <strong>替你尝一口</strong>
          <small>让别人的味蕾，先替你探路</small>
        </span>
      </RouterLink>

      <div class="top-right">
        <template v-if="auth.isLoggedIn">
          <RouterLink to="/profile" class="wallet">
            <span class="wallet-avatar">{{ auth.user.avatar_emoji }}</span>
            <span class="coin">{{ auth.user.coins }}</span>
          </RouterLink>
        </template>
        <template v-else>
          <RouterLink to="/login" class="btn btn-ghost btn-sm">登录 / 注册</RouterLink>
        </template>
      </div>
    </header>

    <main class="app-main">
      <RouterView />
    </main>

    <nav class="tabbar">
      <RouterLink to="/" class="tab" active-class="active">
        <span class="tab-ico">🧭</span><span>首页 Feed</span>
      </RouterLink>
      <RouterLink :to="auth.isLoggedIn ? '/profile' : '/login'" class="tab" active-class="active">
        <span class="tab-ico">👤</span><span>个人中心</span>
      </RouterLink>
    </nav>

    <ToastHost />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { auth } from './store'
import http from './api'
import ToastHost from './ToastHost.vue'

// 每次进入应用时刷新用户最新余额
onMounted(async () => {
  if (auth.isLoggedIn) {
    try {
      const { data } = await http.get('/api/me')
      auth.setUser(data.user)
    } catch {
      /* 401 拦截器会处理 */
    }
  }
})
</script>

<style scoped>
.app-shell { min-height: 100vh; padding-bottom: 72px; }

.topbar {
  position: sticky;
  top: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: rgba(253, 248, 243, 0.88);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--line);
}
.brand { display: flex; align-items: center; gap: 10px; }
.brand-logo {
  font-size: 28px;
  filter: drop-shadow(0 3px 5px rgba(232, 85, 45, 0.3));
}
.brand-text { display: flex; flex-direction: column; line-height: 1.15; }
.brand-text strong { font-size: 17px; color: var(--brand-dark); letter-spacing: 0.5px; }
.brand-text small { font-size: 11px; color: var(--ink-soft); }

.wallet {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  border: 1.5px solid var(--line);
  border-radius: 999px;
  padding: 4px 14px 4px 5px;
  box-shadow: 0 3px 10px rgba(180, 90, 40, 0.08);
}
.wallet-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--brand-soft);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 17px;
}

.app-main { max-width: 880px; margin: 0 auto; padding: 0 16px; }

.tabbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 64px;
  background: #fff;
  border-top: 1px solid var(--line);
  display: flex;
  z-index: 50;
}
.tab {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  font-size: 12px;
  color: var(--ink-soft);
}
.tab-ico { font-size: 21px; }
.tab.active { color: var(--brand); font-weight: 700; }
.tab.active .tab-ico { transform: translateY(-1px) scale(1.08); }

@media (min-width: 880px) {
  .tabbar { max-width: 880px; left: 50%; transform: translateX(-50%); border-radius: 18px 18px 0 0; }
}
</style>
