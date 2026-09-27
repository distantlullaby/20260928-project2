import { reactive, watch } from 'vue'

const STORAGE_KEY = 'taste_for_you_auth'

function load() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY)) || { token: '', user: null }
  } catch {
    return { token: '', user: null }
  }
}

const state = reactive(load())

watch(
  state,
  (v) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ token: v.token, user: v.user }))
  },
  { deep: true }
)

export function useAuth() {
  return {
    get token() {
      return state.token
    },
    get user() {
      return state.user
    },
    get isLoggedIn() {
      return !!state.token
    },
    setSession({ token, user }) {
      state.token = token
      state.user = user
    },
    setUser(user) {
      state.user = user
    },
    logout() {
      state.token = ''
      state.user = null
    }
  }
}
