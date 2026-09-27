import { reactive } from 'vue'

// 全局 UI 状态（登录/注册弹窗）
const ui = reactive({
  authModalOpen: false,
  authMode: 'login'
})

export function useUI() {
  return {
    ui,
    openAuth(mode = 'login') {
      ui.authMode = mode
      ui.authModalOpen = true
    },
    closeAuth() {
      ui.authModalOpen = false
    }
  }
}
