<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal-card">
      <button class="modal-close" @click="$emit('close')">✕</button>
      <h3 class="modal-title">{{ isLogin ? '欢迎回来' : '注册领尝鲜币' }}</h3>
      <p class="modal-sub">
        {{ isLogin ? '登录后即可替别人去吃、赚取悬赏' : '注册即赠送 100 枚尝鲜币，马上发布你的第一份心愿' }}
      </p>

      <label class="label">用户名</label>
      <input v-model="form.username" class="input" placeholder="2-32 位用户名" maxlength="32" />
      <label class="label">密码</label>
      <input v-model="form.password" class="input" type="password" placeholder="4-32 位密码" maxlength="32" @keyup.enter="submit" />
      <template v-if="!isLogin">
        <label class="label">昵称（可选）</label>
        <input v-model="form.nickname" class="input" placeholder="给自己起个吃货昵称" maxlength="32" @keyup.enter="submit" />
      </template>

      <p v-if="error" class="err">{{ error }}</p>

      <button class="btn btn-primary btn-block" style="margin-top: 18px" :disabled="loading" @click="submit">
        {{ loading ? '请稍候…' : isLogin ? '登录' : '注册并领取 100 🪙' }}
      </button>
      <p class="switch">
        {{ isLogin ? '还没有账号？' : '已有账号？' }}
        <a href="#" @click.prevent="toggleMode">{{ isLogin ? '去注册' : '去登录' }}</a>
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed } from 'vue'
import { login, register } from '../api'
import { useAuth } from '../store/auth'
import { toast } from '../store/toast'

const props = defineProps({ mode: { type: String, default: 'login' } })
const emit = defineEmits(['close', 'ok'])

const auth = useAuth()
const mode = ref(props.mode)
const isLogin = computed(() => mode.value === 'login')
const form = reactive({ username: '', password: '', nickname: '' })
const error = ref('')
const loading = ref(false)

function toggleMode() {
  error.value = ''
  mode.value = isLogin.value ? 'register' : 'login'
}

async function submit() {
  error.value = ''
  if (form.username.trim().length < 2 || form.password.length < 4) {
    error.value = '用户名至少 2 位，密码至少 4 位'
    return
  }
  loading.value = true
  try {
    const payload = { username: form.username.trim(), password: form.password }
    if (!isLogin.value) payload.nickname = form.nickname.trim()
    const res = isLogin.value ? await login(payload) : await register(payload)
    auth.setSession(res)
    toast(isLogin.value ? `欢迎回来，${res.user.nickname}` : `注册成功，${res.user.nickname} 获得 100 枚尝鲜币！`)
    emit('ok')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.err {
  color: var(--red);
  font-size: 13px;
  margin: 10px 0 0;
}
.switch {
  text-align: center;
  font-size: 13px;
  color: var(--ink-3);
  margin: 14px 0 0;
}
</style>
