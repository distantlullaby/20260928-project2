import { reactive, watch } from 'vue'

// 极简全局状态：当前登录用户与令牌，持久化到 localStorage
const saved = (() => {
  try {
    return JSON.parse(localStorage.getItem('taste_auth') || 'null')
  } catch {
    return null
  }
})()

export const auth = reactive({
  token: saved?.token || '',
  user: saved?.user || null,

  login(token, user) {
    this.token = token
    this.user = user
  },
  setUser(user) {
    this.user = user
  },
  logout() {
    this.token = ''
    this.user = null
  },
  get isLoggedIn() {
    return !!this.token
  }
})

watch(
  () => [auth.token, auth.user],
  () => {
    if (auth.token) {
      localStorage.setItem('taste_auth', JSON.stringify({ token: auth.token, user: auth.user }))
    } else {
      localStorage.removeItem('taste_auth')
    }
  },
  { deep: true }
)
