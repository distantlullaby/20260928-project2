<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-logo">🍜</div>
      <h1>{{ mode === 'login' ? '欢迎回来' : '加入替你尝一口' }}</h1>
      <p class="muted">
        {{ mode === 'login' ? '登录后发求代吃、替别人去吃赚尝鲜币' : '注册即送 100 尝鲜币，开启云探味之旅' }}
      </p>

      <div class="mode-switch">
        <button :class="{ on: mode === 'login' }" @click="mode = 'login'">登录</button>
        <button :class="{ on: mode === 'register' }" @click="mode = 'register'">注册</button>
      </div>

      <form @submit.prevent="submit">
        <div class="field">
          <label>用户名</label>
          <input v-model="form.username" minlength="3" maxlength="32" placeholder="3 位以上用户名" required />
        </div>
        <div v-if="mode === 'register'" class="field">
          <label>昵称</label>
          <input v-model="form.nickname" maxlength="32" placeholder="大家会怎么称呼你" required />
        </div>
        <div class="field">
          <label>密码</label>
          <input v-model="form.password" type="password" minlength="6" maxlength="64" placeholder="至少 6 位" required />
        </div>
        <div v-if="mode === 'register'" class="field">
          <label>选个头像表情</label>
          <div class="emoji-row">
            <button
              v-for="e in emojis"
              :key="e"
              type="button"
              :class="{ on: form.avatar_emoji === e }"
              @click="form.avatar_emoji = e"
            >{{ e }}</button>
          </div>
        </div>

        <button class="btn btn-primary btn-block" :disabled="loading">
          {{ loading ? '请稍候…' : mode === 'login' ? '登录' : '注册并领取 100 币' }}
        </button>
      </form>

      <div class="demo-hint">
        <p class="small muted">🎁 演示账号（密码均为 123456）：</p>
        <div class="demo-users">
          <button type="button" v-for="u in demo" :key="u" class="demo-chip" @click="fillDemo(u)">{{ u }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import http, { errMsg } from '../api'
import { auth } from '../store'
import { toast } from '../toast'

const router = useRouter()
const route = useRoute()
const mode = ref('login')
const loading = ref(false)
const emojis = ['🍜', '🍢', '🍰', '🍔', '🍣', '🥟', '🧋', '🍕', '🌮', '🍦']
const demo = ['ali', 'bo', 'ci']

const form = reactive({ username: '', nickname: '', password: '', avatar_emoji: '🍜' })

function fillDemo(u) {
  mode.value = 'login'
  form.username = u
  form.password = '123456'
}

async function submit() {
  loading.value = true
  try {
    const url = mode.value === 'login' ? '/api/login' : '/api/register'
    const payload = mode.value === 'login'
      ? { username: form.username, password: form.password }
      : form
    const { data } = await http.post(url, payload)
    auth.login(data.token, data.user)
    toast.show(mode.value === 'login' ? '登录成功，开吃！' : '注册成功，100 尝鲜币已到账 🎁', 'ok')
    router.push(route.query.redirect || '/')
  } catch (e) {
    toast.show(errMsg(e), 'err')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: calc(100vh - 64px);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 30px 16px;
}
.auth-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 24px;
  padding: 30px 26px;
  box-shadow: var(--shadow);
  border: 1px solid var(--line);
}
.auth-logo { font-size: 46px; text-align: center; margin-bottom: 6px; }
h1 { margin: 0 0 6px; font-size: 22px; text-align: center; }
.auth-card > p { text-align: center; margin: 0 0 20px; font-size: 13px; }

.mode-switch {
  display: flex;
  background: #f6ede5;
  border-radius: 999px;
  padding: 4px;
  margin-bottom: 22px;
}
.mode-switch button {
  flex: 1;
  border: none;
  background: transparent;
  padding: 9px;
  border-radius: 999px;
  font-weight: 700;
  font-size: 14px;
  color: var(--ink-soft);
  cursor: pointer;
}
.mode-switch button.on { background: #fff; color: var(--brand-dark); box-shadow: 0 2px 8px rgba(180,90,40,0.15); }

.emoji-row { display: flex; flex-wrap: wrap; gap: 6px; }
.emoji-row button {
  width: 38px; height: 38px;
  border: 2px solid var(--line);
  background: #fff;
  border-radius: 12px;
  font-size: 20px;
  cursor: pointer;
}
.emoji-row button.on { border-color: var(--brand); background: var(--brand-soft); transform: scale(1.08); }

.demo-hint { margin-top: 22px; text-align: center; }
.demo-users { display: flex; gap: 8px; justify-content: center; margin-top: 8px; }
.demo-chip {
  border: 1px dashed var(--brand);
  color: var(--brand-dark);
  background: var(--brand-soft);
  border-radius: 999px;
  padding: 5px 14px;
  font-size: 13px;
  cursor: pointer;
  font-family: monospace;
}
.demo-chip:hover { background: #ffe1d3; }
</style>
