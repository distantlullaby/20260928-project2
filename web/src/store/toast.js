import { reactive } from 'vue'

// 全局轻提示
const state = reactive({ text: '', timer: null })

export function toast(text) {
  state.text = text
  if (state.timer) clearTimeout(state.timer)
  state.timer = setTimeout(() => (state.text = ''), 2200)
}

export function useToast() {
  return state
}
